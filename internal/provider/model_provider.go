package provider

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type catoProviderModel struct {
	BaseURL                    types.String `tfsdk:"baseurl"`
	Token                      types.String `tfsdk:"token"`
	AccountID                  types.String `tfsdk:"account_id"`
	RetryMax                   types.Int64  `tfsdk:"retry_max"`
	RetryWaitMinSeconds        types.Int64  `tfsdk:"retry_wait_min_seconds"`
	RetryWaitMaxSeconds        types.Int64  `tfsdk:"retry_wait_max_seconds"`
	VersionCheckDisabled       types.Bool   `tfsdk:"version_check_disabled"`
	VersionCheckTimeoutSeconds types.Int64  `tfsdk:"version_check_timeout_seconds"`
}

type retryClientConfig struct {
	retryMax     int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
}
