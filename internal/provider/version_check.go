package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

const (
	registryProviderVersionsURL = "https://registry.terraform.io/v1/providers/catonetworks/cato/versions"
	registryVersionCheckMaxBody = 1 << 20
)

type registryVersionChecker struct {
	client *http.Client
	url    string
}

type registryVersionsResponse struct {
	Versions []registryProviderVersion `json:"versions"`
}

type registryProviderVersion struct {
	Version string `json:"version"`
}

type versionUpdate struct {
	installed *semver.Version
	latest    *semver.Version
}

func newRegistryVersionChecker(client *http.Client) registryVersionChecker {
	return registryVersionChecker{
		client: client,
		url:    registryProviderVersionsURL,
	}
}

func (p *catoProvider) warnIfNewVersionAvailable(
	ctx context.Context,
	resp *provider.ConfigureResponse,
	timeout time.Duration,
) {
	if !p.hasCheckedVersion.CompareAndSwap(false, true) {
		return
	}

	update, err := p.versionChecker.check(ctx, p.version, timeout)
	if err != nil || update == nil {
		return
	}

	resp.Diagnostics.Append(versionUpdateDiagnostic(update))
}

func (c registryVersionChecker) check(
	ctx context.Context,
	installedVersion string,
	timeout time.Duration,
) (*versionUpdate, error) {
	installed, err := semver.NewVersion(installedVersion)
	if err != nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create Terraform Registry version request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	client := c.client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Terraform Registry provider versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Terraform Registry provider versions returned unexpected status %d", resp.StatusCode)
	}

	var body registryVersionsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, registryVersionCheckMaxBody)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode Terraform Registry provider versions: %w", err)
	}

	latest, err := latestStableVersion(body.Versions)
	if err != nil {
		return nil, err
	}

	return versionUpdateFromVersions(installed, latest), nil
}

func latestStableVersion(versions []registryProviderVersion) (*semver.Version, error) {
	var latest *semver.Version

	for _, version := range versions {
		candidate, err := semver.NewVersion(version.Version)
		if err != nil || candidate.Prerelease() != "" {
			continue
		}
		if latest == nil || latest.LessThan(candidate) {
			latest = candidate
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("Terraform Registry returned no valid stable provider versions")
	}

	return latest, nil
}

func newVersionUpdate(installedVersion, latestVersion string) (*versionUpdate, error) {
	installed, err := semver.NewVersion(installedVersion)
	if err != nil {
		return nil, fmt.Errorf("parse installed provider version: %w", err)
	}
	latest, err := semver.NewVersion(latestVersion)
	if err != nil {
		return nil, fmt.Errorf("parse latest provider version: %w", err)
	}

	return versionUpdateFromVersions(installed, latest), nil
}

func versionUpdateFromVersions(installed, latest *semver.Version) *versionUpdate {
	if latest == nil || !installed.LessThan(latest) {
		return nil
	}

	return &versionUpdate{
		installed: installed,
		latest:    latest,
	}
}

func versionUpdateDiagnostic(update *versionUpdate) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		"New Cato Terraform Provider Version Available",
		fmt.Sprintf(
			"Version %s is installed; version %s is available (%s release). "+
				"Review the release notes at https://github.com/catonetworks/terraform-provider-cato/releases/tag/v%s before upgrading. "+
				"Terraform will continue using your configured provider version.",
			update.installed,
			update.latest,
			releaseType(update.installed, update.latest),
			update.latest,
		),
	)
}

func releaseType(installed, latest *semver.Version) string {
	switch {
	case installed.Major() != latest.Major():
		return "major"
	case installed.Minor() != latest.Minor():
		return "minor"
	case installed.Patch() != latest.Patch():
		return "patch"
	default:
		return "stable"
	}
}
