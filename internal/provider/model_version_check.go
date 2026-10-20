package provider

import (
	"github.com/Masterminds/semver/v3"
)

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
