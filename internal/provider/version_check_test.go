package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/stretchr/testify/require"
)

func TestRegistryVersionCheckerCheckFindsNewestStableVersion(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Accept"))
		_, err := w.Write([]byte(`{
			"versions": [
				{"version": "1.3.0-beta.1"},
				{"version": "1.2.0"},
				{"version": "1.3.0"},
				{"version": "not-a-version"}
			]
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	checker := registryVersionChecker{client: server.Client(), url: server.URL}
	update, err := checker.check(context.Background(), "1.2.0", time.Second)

	require.NoError(t, err)
	require.NotNil(t, update)
	require.Equal(t, "1.2.0", update.installed.String())
	require.Equal(t, "1.3.0", update.latest.String())
	require.Equal(t, "minor", releaseType(update.installed, update.latest))
}

func TestRegistryVersionCheckerCheckIgnoresPrereleasesAndOlderVersions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`{
			"versions": [
				{"version": "2.0.0-beta.1"},
				{"version": "1.4.0"},
				{"version": "1.3.9"}
			]
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	checker := registryVersionChecker{client: server.Client(), url: server.URL}
	update, err := checker.check(context.Background(), "1.4.0", time.Second)

	require.NoError(t, err)
	require.Nil(t, update)
}

func TestRegistryVersionCheckerCheckReturnsErrorForUnexpectedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	checker := registryVersionChecker{client: server.Client(), url: server.URL}
	update, err := checker.check(context.Background(), "1.2.0", time.Second)

	require.Error(t, err)
	require.Nil(t, update)
}

func TestRegistryVersionCheckerCheckHonorsTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	checker := registryVersionChecker{client: server.Client(), url: server.URL}
	update, err := checker.check(context.Background(), "1.2.0", 10*time.Millisecond)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Nil(t, update)
}

func TestVersionUpdateDiagnosticIncludesVersionInformation(t *testing.T) {
	t.Parallel()

	update, err := newVersionUpdate("1.2.3", "2.0.0")
	require.NoError(t, err)

	diagnostic := versionUpdateDiagnostic(update)

	require.Equal(t, "New Cato Terraform Provider Version Available", diagnostic.Summary())
	require.Contains(t, diagnostic.Detail(), "Version 1.2.3 is installed; version 2.0.0 is available (major release).")
	require.Contains(t, diagnostic.Detail(), "Terraform will continue using your configured provider version.")
}

func TestCatoProviderWarnIfNewVersionAvailableAddsNonBlockingDiagnosticOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`{"versions": [{"version": "1.2.4"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	p := catoProvider{
		version:        "1.2.3",
		versionChecker: registryVersionChecker{client: server.Client(), url: server.URL},
	}
	resp := &frameworkprovider.ConfigureResponse{}

	p.warnIfNewVersionAvailable(context.Background(), resp, time.Second)
	p.warnIfNewVersionAvailable(context.Background(), resp, time.Second)

	require.Len(t, resp.Diagnostics, 1)
	require.Equal(t, "New Cato Terraform Provider Version Available", resp.Diagnostics[0].Summary())
}

func TestCatoProviderWarnIfNewVersionAvailableReportsMinorUpdate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`{
			"versions": [
				{"version": "1.4.0"},
				{"version": "1.5.0"},
				{"version": "1.6.0-beta.1"}
			]
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	p := catoProvider{
		version:        "1.4.0",
		versionChecker: registryVersionChecker{client: server.Client(), url: server.URL},
	}
	resp := &frameworkprovider.ConfigureResponse{}

	p.warnIfNewVersionAvailable(context.Background(), resp, time.Second)

	require.Len(t, resp.Diagnostics, 1)
	require.Contains(
		t,
		resp.Diagnostics[0].Detail(),
		"Version 1.4.0 is installed; version 1.5.0 is available (minor release).",
	)
	require.NotContains(t, resp.Diagnostics[0].Detail(), "1.6.0-beta.1")
}

func TestCatoProviderWarnIfNewVersionAvailableReportsMajorUpdate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`{
			"versions": [
				{"version": "1.5.0"},
				{"version": "2.0.0"},
				{"version": "3.0.0-beta.1"}
			]
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	p := catoProvider{
		version:        "1.5.0",
		versionChecker: registryVersionChecker{client: server.Client(), url: server.URL},
	}
	resp := &frameworkprovider.ConfigureResponse{}

	p.warnIfNewVersionAvailable(context.Background(), resp, time.Second)

	require.Len(t, resp.Diagnostics, 1)
	require.Contains(
		t,
		resp.Diagnostics[0].Detail(),
		"Version 1.5.0 is installed; version 2.0.0 is available (major release).",
	)
	require.NotContains(t, resp.Diagnostics[0].Detail(), "3.0.0-beta.1")
}

func TestCatoProviderWarnIfNewVersionAvailableIgnoresRegistryFailures(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	p := catoProvider{
		version:        "1.2.3",
		versionChecker: registryVersionChecker{client: server.Client(), url: server.URL},
	}
	resp := &frameworkprovider.ConfigureResponse{}

	p.warnIfNewVersionAvailable(context.Background(), resp, time.Second)

	require.Empty(t, resp.Diagnostics)
}

func TestCatoProviderSchemaExposesVersionCheckTimeout(t *testing.T) {
	t.Parallel()

	resp := &frameworkprovider.SchemaResponse{}
	(&catoProvider{}).Schema(context.Background(), frameworkprovider.SchemaRequest{}, resp)

	attribute, ok := resp.Schema.Attributes["version_check_timeout_seconds"].(schema.Int64Attribute)
	require.True(t, ok)
	require.True(t, attribute.Optional)
	require.Contains(t, attribute.Description, "CATO_VERSION_CHECK_TIMEOUT_SECONDS")
}

func TestDefaultVersionCheckTimeout(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(2), defaultVersionCheckTimeoutSeconds)
}
