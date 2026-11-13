package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/access/appconnector"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/access/data_sources/appconnectorgroup"
	appconnectorsds "github.com/catonetworks/terraform-provider-cato/internal/provider/access/data_sources/appconnectors"
	privateappsds "github.com/catonetworks/terraform-provider-cato/internal/provider/access/data_sources/privateapps"
	usersds "github.com/catonetworks/terraform-provider-cato/internal/provider/access/data_sources/users"
	usersgroupds "github.com/catonetworks/terraform-provider-cato/internal/provider/access/data_sources/usersgroup"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/access/privateapp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/access/user"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/account"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/admin"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/licensinginfo"
	mailinglistds "github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/mailinglist"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/permissioncatalog"
	roleds "github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/role"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/roles"
	subscriptiongroupds "github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/subscriptiongroup"
	webhookds "github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/webhook"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/license"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/mailinglist"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/rbacassignment"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/role"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/subscriptiongroup"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/webhook"
	allocatedipres "github.com/catonetworks/terraform-provider-cato/internal/provider/network/allocatedip"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/cloudinterconnectphysicalconnection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/cloudinterconnectsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/accountsnapshotsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/allocatedip"
	dhcprelaygroupds "github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/dhcprelaygroup"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/host"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/networkinterfaces"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/networkranges"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/sitelocation"
	sitesds "github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/sites"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/dhcpoption"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/dhcprelaygroup"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/dhcpsettings"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/dnsforwardingrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/dnssettings"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/dnssitesettings"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/ipsecsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/laninterface"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/laninterfacelagmember"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkrange"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/secondaryawsvsocket"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/secondaryazurevsocket"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/socketsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/wifissid"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/antimalware/amfhindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/antimalware/amfhpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/antimalware/amfhrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/antimalware/amfhsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/clientconnectivity/ccindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/clientconnectivity/ccpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/clientconnectivity/ccrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/clientconnectivity/ccsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/dynamicip/diaindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/dynamicip/diapolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/dynamicip/diarule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/dynamicip/diasection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/data_sources/ifrulesections"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/data_sources/ifwrulesindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifsubpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/lan/lanindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/lan/lansection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/lan/lfsubpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/lan/socketlanfirewallrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/lan/socketlannetworkrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/privateaccess/papolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/privateaccess/parule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/privateaccess/parulebulk"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/privateaccess/pasection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/remoteportfwd/rpfindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/remoteportfwd/rpfpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/remoteportfwd/rpfrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/remoteportfwd/rpfsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip/ssiaindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip/ssiapolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip/ssiarule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip/ssiasection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy/swpindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy/swppolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy/swprule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy/swpsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass/sbindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass/sbpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass/sbrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass/sbsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel/stindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel/stpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel/strule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel/stsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver/tsindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver/tspolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver/tsrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver/tssection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tls/data_sources/tlsrulesindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tls/tlsindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tls/tlsrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tls/tlssection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/data_sources/wfrulesections"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/data_sources/wfrulesindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wanindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfsubpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wnwindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wnwpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wnwrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wnwsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/bgppeer"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/customapplication"
	customapplicationsds "github.com/catonetworks/terraform-provider-cato/internal/provider/resource/data_sources/customapplications"
	fqdncontainerds "github.com/catonetworks/terraform-provider-cato/internal/provider/resource/data_sources/fqdncontainer"
	globaliprangesds "github.com/catonetworks/terraform-provider-cato/internal/provider/resource/data_sources/globalipranges"
	groupdata "github.com/catonetworks/terraform-provider-cato/internal/provider/resource/data_sources/group"
	iprangecontainerds "github.com/catonetworks/terraform-provider-cato/internal/provider/resource/data_sources/ipaddressrangecontainer"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/fqdncontainer"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/globalipranges"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/group"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/groupmembers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/ipaddressrangecontainer"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/statichost"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/waninterface"
	secrpfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/security/remoteportforward/rpfrule"
	secrpfsection "github.com/catonetworks/terraform-provider-cato/internal/provider/security/remoteportforward/rpfsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ provider.Provider = &catoProvider{}
)

const (
	defaultRetryMax                   int64 = 5
	defaultRetryWaitMinSeconds        int64 = 1
	defaultRetryWaitMaxSeconds        int64 = 30
	defaultVersionCheckTimeoutSeconds int64 = 2
	policyRevisionNotFound                  = "PolicyRevisionNotFound"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &catoProvider{
			version:        version,
			versionChecker: newRegistryVersionChecker(http.DefaultClient),
		}
	}
}

type catoProvider struct {
	version            string
	hasBeenInitialized atomic.Bool
	hasCheckedVersion  atomic.Bool
	versionChecker     registryVersionChecker
}

func (p *catoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cato"
	resp.Version = p.version
}

func (p *catoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"baseurl": schema.StringAttribute{
				Description: "URL for the Cato API. Can be provided using CATO_BASEURL environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API Key for the Cato API (can be provided via the CATO_TOKEN environment " +
					"variable). Part of the explicit-credentials method; mutually exclusive with `profile`.",
				Optional:  true,
				Sensitive: true,
			},
			"account_id": schema.StringAttribute{
				Description: "AccountId for the Cato API (can be provided via the CATO_ACCOUNT_ID " +
					"environment variable). Part of the explicit-credentials method; mutually exclusive " +
					"with `profile`.",
				Optional: true,
			},
			"profile": schema.StringAttribute{
				Description: "Name of a catocli profile to source credentials from. When set (or when " +
					"CATO_PROFILE is set), the provider reads token, account_id, and endpoint from the " +
					"named section of the catocli credentials file (~/.cato/credentials by default; " +
					"override the path with the CATO_CREDENTIALS_FILE environment variable). This is the " +
					"local-development auth method and is MUTUALLY EXCLUSIVE with token/account_id " +
					"attributes in the provider block. A selected profile overrides ambient CATO_TOKEN, " +
					"CATO_ACCOUNT_ID, and CATO_BASEURL values; an explicit baseurl attribute still wins. For CI use " +
					"the explicit-credentials env vars instead. See the provider README's Authentication " +
					"section.",
				Optional: true,
			},
			"retry_max": schema.Int64Attribute{
				Description: "Maximum number of retries for retryable API requests. " +
					"Defaults to 5. Can be provided using CATO_RETRY_MAX environment variable.",
				Optional: true,
			},
			"retry_wait_min_seconds": schema.Int64Attribute{
				Description: "Minimum backoff between retry attempts, in seconds. " +
					"Defaults to 1. Can be provided using CATO_RETRY_WAIT_MIN_SECONDS environment variable.",
				Optional: true,
			},
			"retry_wait_max_seconds": schema.Int64Attribute{
				Description: "Maximum backoff between retry attempts, in seconds. " +
					"Defaults to 30. Can be provided using CATO_RETRY_WAIT_MAX_SECONDS environment variable.",
				Optional: true,
			},
			"version_check_disabled": schema.BoolAttribute{
				Description: "Disables the advisory Terraform Registry version check. " +
					"Can be provided using CATO_VERSION_CHECK_DISABLED environment variable; any non-empty value disables the check.",
				Optional: true,
			},
			"version_check_timeout_seconds": schema.Int64Attribute{
				Description: "Maximum time for the advisory Terraform Registry version check, in seconds. " +
					"Defaults to 2. Can be provided using CATO_VERSION_CHECK_TIMEOUT_SECONDS environment variable.",
				Optional: true,
			},
		},
	}
}

//nolint:gocyclo,funlen
func (p *catoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config catoProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("baseurl"),
			"Unknown Cato API Base URL ",
			"The provider cannot create the CATO API client as there is an unknown configuration value for the CATO API base URL. ",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Cato API Token",
			"The provider cannot create the CATO API client as there is an unknown configuration value for the CATO API token. ",
		)
	}

	if config.AccountID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Unknown Cato API account_id",
			"The provider cannot create the CATO API client as there is an unknown configuration value for the CATO API account_id. ",
		)
	}

	if config.Profile.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("profile"),
			"Unknown Cato CLI Profile",
			"The provider cannot resolve credentials as there is an unknown configuration value for profile.",
		)
	}

	if config.RetryMax.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_max"),
			"Unknown Retry Max",
			"The provider cannot create the CATO API client as there is an unknown configuration value for retry_max.",
		)
	}

	if config.RetryWaitMinSeconds.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_min_seconds"),
			"Unknown Retry Minimum Wait",
			"The provider cannot create the CATO API client as there is an unknown configuration value for retry_wait_min_seconds.",
		)
	}

	if config.RetryWaitMaxSeconds.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_max_seconds"),
			"Unknown Retry Maximum Wait",
			"The provider cannot create the CATO API client as there is an unknown configuration value for retry_wait_max_seconds.",
		)
	}

	if config.VersionCheckDisabled.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("version_check_disabled"),
			"Unknown Version Check Disabled",
			"The provider cannot determine whether to disable the version check from an unknown configuration value.",
		)
	}

	if config.VersionCheckTimeoutSeconds.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("version_check_timeout_seconds"),
			"Unknown Version Check Timeout",
			"The provider cannot determine the version check timeout from an unknown configuration value.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	baseurl := os.Getenv("CATO_BASEURL")
	token := os.Getenv("CATO_TOKEN")
	retryMax, retryMaxErr := int64FromEnv("CATO_RETRY_MAX")
	retryWaitMinSeconds, retryWaitMinErr := int64FromEnv("CATO_RETRY_WAIT_MIN_SECONDS")
	retryWaitMaxSeconds, retryWaitMaxErr := int64FromEnv("CATO_RETRY_WAIT_MAX_SECONDS")
	versionCheckDisabled := os.Getenv("CATO_VERSION_CHECK_DISABLED") != ""
	versionCheckTimeoutSeconds, versionCheckTimeoutErr := int64FromEnv("CATO_VERSION_CHECK_TIMEOUT_SECONDS")

	if !config.BaseURL.IsNull() {
		baseurl = config.BaseURL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.RetryMax.IsNull() {
		value := config.RetryMax.ValueInt64()
		retryMax = &value
	}

	if !config.RetryWaitMinSeconds.IsNull() {
		value := config.RetryWaitMinSeconds.ValueInt64()
		retryWaitMinSeconds = &value
	}

	if !config.RetryWaitMaxSeconds.IsNull() {
		value := config.RetryWaitMaxSeconds.ValueInt64()
		retryWaitMaxSeconds = &value
	}

	if !config.VersionCheckDisabled.IsNull() {
		versionCheckDisabled = config.VersionCheckDisabled.ValueBool()
	}

	if !config.VersionCheckTimeoutSeconds.IsNull() {
		value := config.VersionCheckTimeoutSeconds.ValueInt64()
		versionCheckTimeoutSeconds = &value
	}

	if retryMax == nil {
		value := defaultRetryMax
		retryMax = &value
	}

	if retryWaitMinSeconds == nil {
		value := defaultRetryWaitMinSeconds
		retryWaitMinSeconds = &value
	}

	if retryWaitMaxSeconds == nil {
		value := defaultRetryWaitMaxSeconds
		retryWaitMaxSeconds = &value
	}

	if versionCheckTimeoutSeconds == nil {
		value := defaultVersionCheckTimeoutSeconds
		versionCheckTimeoutSeconds = &value
	}

	accountID := os.Getenv("CATO_ACCOUNT_ID")
	if !config.AccountID.IsNull() {
		accountID = config.AccountID.ValueString()
	}

	// A configured `profile` is AUTHORITATIVE. An explicitly-set profile (the
	// provider-block `profile` attribute, or CATO_PROFILE) wins over ambient
	// CATO_TOKEN / CATO_ACCOUNT_ID / CATO_BASEURL environment variables -- a stray
	// shell credential (e.g. a leftover export for another tenant) must never
	// override or conflict with the profile the configuration intentionally
	// selected. The mutual-exclusion ERROR is reserved for a genuine CONFIG
	// conflict: a selected profile and explicit `token`/`account_id` attributes.
	profileName := ""
	if !config.Profile.IsNull() {
		profileName = config.Profile.ValueString()
	} else if envProfile := os.Getenv("CATO_PROFILE"); envProfile != "" {
		profileName = envProfile
	}
	hasProfile := profileName != ""
	explicitAttrSet := !config.Token.IsNull() || !config.AccountID.IsNull()

	const authHelp = "Authenticate using exactly ONE of: (A) explicit credentials -- set " +
		"`token` and `account_id` in the provider block, or the CATO_TOKEN / CATO_ACCOUNT_ID " +
		"(and CATO_BASEURL) environment variables (use this for CI such as Azure DevOps or " +
		"Terraform Cloud, where there is no local credentials file); or (B) a catocli profile " +
		"-- set `profile` in the provider block (or the CATO_PROFILE environment variable) to a " +
		"profile name from the catocli credentials file (~/.cato/credentials by default, " +
		"overridable with CATO_CREDENTIALS_FILE)."

	// A selected profile conflicts with explicit credential attributes.
	// Ambient CATO_TOKEN/CATO_ACCOUNT_ID env vars are NOT a conflict --
	// the profile wins and they are ignored.
	if hasProfile && explicitAttrSet {
		resp.Diagnostics.AddError(
			"Conflicting Cato API Credentials",
			"A profile is selected through `profile` or CATO_PROFILE, and the provider block sets `token`/`account_id`. These "+
				"authentication methods are mutually exclusive -- set only one in the "+
				"provider block. "+authHelp,
		)
		return
	}

	// Profile wins: resolve it and REPLACE any ambient token/account_id/baseurl
	// (an explicit `baseurl` attribute still overrides the profile endpoint).
	if hasProfile {
		credsPath := credentialsFilePath()
		profileToken, profileAccountID, profileEndpoint, profileErr := ResolveProfileCredentials(profileName, credsPath)
		if profileErr != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("profile"),
				"Unable to Resolve Cato CLI Profile",
				profileErr.Error(),
			)
		} else {
			token = profileToken
			accountID = profileAccountID
			baseurl = profileEndpoint
			if !config.BaseURL.IsNull() {
				baseurl = config.BaseURL.ValueString()
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if token == "" && accountID == "" {
		resp.Diagnostics.AddError(
			"Unable to Authenticate to the Cato API",
			"No Cato API credentials were found in the provider configuration or environment. "+authHelp,
		)
	} else {
		if baseurl == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("baseurl"),
				"Missing Cato API Base URL ",
				"The provider cannot create the CATO API client as there is a missing or empty value for the CATO API URL. "+authHelp,
			)
		}

		if token == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("token"),
				"Missing Cato API Token ",
				"The provider cannot create the CATO API client as there is a missing or empty value for the CATO API Token. "+authHelp,
			)
		}

		if accountID == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("account_id"),
				"Missing Cato API account_id",
				"The provider cannot create the CATO API client as there is a missing or empty value for account_id. "+authHelp,
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if retryMaxErr != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_max"),
			"Invalid Retry Max Environment Variable",
			retryMaxErr.Error(),
		)
	}

	if retryWaitMinErr != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_min_seconds"),
			"Invalid Retry Minimum Wait Environment Variable",
			retryWaitMinErr.Error(),
		)
	}

	if retryWaitMaxErr != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_max_seconds"),
			"Invalid Retry Maximum Wait Environment Variable",
			retryWaitMaxErr.Error(),
		)
	}

	if versionCheckTimeoutErr != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("version_check_timeout_seconds"),
			"Invalid Version Check Timeout Environment Variable",
			versionCheckTimeoutErr.Error(),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if retryMax != nil && *retryMax < 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_max"),
			"Invalid Retry Max",
			"The provider retry_max value must be greater than or equal to 0.",
		)
	}

	if retryWaitMinSeconds != nil && *retryWaitMinSeconds <= 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_min_seconds"),
			"Invalid Retry Minimum Wait",
			"The provider retry_wait_min_seconds value must be greater than 0.",
		)
	}

	if retryWaitMaxSeconds != nil && *retryWaitMaxSeconds <= 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_max_seconds"),
			"Invalid Retry Maximum Wait",
			"The provider retry_wait_max_seconds value must be greater than 0.",
		)
	}

	if retryWaitMinSeconds != nil && retryWaitMaxSeconds != nil && *retryWaitMinSeconds > *retryWaitMaxSeconds {
		resp.Diagnostics.AddAttributeError(
			path.Root("retry_wait_min_seconds"),
			"Invalid Retry Backoff Range",
			"The provider retry_wait_min_seconds value must be less than or equal to retry_wait_max_seconds.",
		)
	}

	if versionCheckTimeoutSeconds != nil && *versionCheckTimeoutSeconds <= 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("version_check_timeout_seconds"),
			"Invalid Version Check Timeout",
			"The provider version_check_timeout_seconds value must be greater than 0.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// newer client:
	headers := map[string]string{}
	headers["User-Agent"] = "cato-terraform-" + p.version
	retryConfig := buildRetryConfig(retryMax, retryWaitMinSeconds, retryWaitMaxSeconds)
	httpClient := buildRetryHTTPClient(retryConfig)
	catoClient, err := cato.New(baseurl, token, accountID, httpClient, headers)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Cato API Client",
			err.Error(),
		)
		return
	}

	dataSourceData := &catoclient.Service{
		BaseURL:              baseurl,
		Token:                token,
		AccountId:            accountID,
		Catov2:               catoclient.NewProviderSDKClient(catoClient),
		AccountSnapshotCache: catoclient.NewAccountSnapshotCache(),
	}

	resp.DataSourceData = dataSourceData
	resp.ResourceData = dataSourceData

	if !versionCheckDisabled {
		p.warnIfNewVersionAvailable(ctx, resp, time.Duration(*versionCheckTimeoutSeconds)*time.Second)
	}

	// cleanup stale rules
	p.cleanupDrafts(ctx, dataSourceData)
}

func int64FromEnv(key string) (*int64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return nil, nil
	}

	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s must be a valid integer, got %q", key, raw)
	}

	return &val, nil
}

func buildRetryConfig(retryMax, retryWaitMinSeconds, retryWaitMaxSeconds *int64) *retryClientConfig {
	if retryMax == nil && retryWaitMinSeconds == nil && retryWaitMaxSeconds == nil {
		return nil
	}

	config := &retryClientConfig{}
	if retryMax != nil {
		config.retryMax = int(*retryMax)
	}
	if retryWaitMinSeconds != nil {
		config.retryWaitMin = time.Duration(*retryWaitMinSeconds) * time.Second
	}
	if retryWaitMaxSeconds != nil {
		config.retryWaitMax = time.Duration(*retryWaitMaxSeconds) * time.Second
	}

	return config
}

func buildRetryHTTPClient(retryConfig *retryClientConfig) *http.Client {
	if retryConfig == nil {
		return nil
	}

	retryClient := retryablehttp.NewClient()
	retryClient.CheckRetry = cato.BaseRetryPolicy
	retryClient.RetryMax = retryConfig.retryMax
	retryClient.RetryWaitMin = retryConfig.retryWaitMin
	retryClient.RetryWaitMax = retryConfig.retryWaitMax

	return retryClient.StandardClient()
}

func (p *catoProvider) cleanupDrafts(ctx context.Context, d *catoclient.Service) {
	if os.Getenv("DISABLE_POLICY_RULE_CLEANUP") == "true" || p.hasBeenInitialized.Load() {
		return
	}
	p.hasBeenInitialized.Store(true)
	resp, err := d.Catov2.PolicyPrivateAccessDiscardRevision(ctx, d.AccountId)
	if err != nil {
		tflog.Error(ctx, "failed to discard draft private-access policy", map[string]any{"err": err})
		return
	}
	errors := resp.GetPolicy().GetPrivateAccess().DiscardPolicyRevision.Errors
	if len(errors) > 0 {
		if errors[0].ErrorCode != nil && *errors[0].ErrorCode == policyRevisionNotFound {
			return // no policy draft to discard; OK
		}
		tflog.Error(ctx, "failed to discard draft private-access policy", map[string]any{"errors": errors})
	}
}

func (p *catoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		accountsnapshotsite.NewAccountSnapshotSiteDataSource,
		roleds.NewDataSource,
		roles.NewDataSource,
		permissioncatalog.NewDataSource,
		allocatedip.Source,
		dhcprelaygroupds.DataSource,
		groupdata.DataSource,
		licensinginfo.NewDataSource,
		networkinterfaces.NetworkInterfacesDataSource,
		sitelocation.DataSource,
		ifwrulesindex.IndexDataSource,
		wfrulesindex.WanRulesIndexDataSource,
		tlsrulesindex.RulesIndexDataSource,
		ifrulesections.IfRuleSectionsDataSource,
		wfrulesections.WfRuleSectionsDataSource,
		networkranges.DataSource,
		host.DataSource,
		sitesds.DataSource,
		usersds.DataSource,
		usersgroupds.DataSource,
		globaliprangesds.DataSource,
		customapplicationsds.DataSource,
		mailinglistds.NewDataSource,
		subscriptiongroupds.NewDataSource,
		webhookds.NewDataSource,
		fqdncontainerds.DataSource,
		iprangecontainerds.DataSource,
		appconnectorgroup.GroupDataSource,
		privateappsds.DataSource,
		appconnectorsds.DataSource,
	}
}

//nolint:funlen // Keep the declarative resource registry together.
func (p *catoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		account.NewAccountResource,
		dnssettings.NewDNSSettingsResource,
		dnssitesettings.NewDNSSiteSettingsResource,
		dnsforwardingrule.NewDNSForwardingRuleResource,
		dhcpsettings.NewDHCPSettingsResource,
		dhcpoption.NewDHCPOptionResource,
		dhcprelaygroup.NewDHCPRelayGroupResource,
		role.NewResource,
		mailinglist.NewResource,
		subscriptiongroup.NewResource,
		webhook.NewResource,
		cloudinterconnectsite.NewResource,
		secondaryawsvsocket.NewResource,
		secondaryazurevsocket.NewResource,
		wifissid.NewResource,
		cloudinterconnectphysicalconnection.NewResource,
		customapplication.NewResource,
		fqdncontainer.NewResource,
		ipaddressrangecontainer.NewResource,
		rbacassignment.NewResource,
		admin.NewAdminResource,
		user.NewResource,
		alwaysonrule.NewAlwaysOnRuleResource,

		sbrule.NewResource,
		sbsection.NewResource,
		pasection.NewResource,
		// MERGE NOTE (v1.5.0-prerelease.1 + feat/bulk-policy-families consolidation):
		// Remote Port Forward is deliberately registered under BOTH naming families:
		//   tag side:    secrpfrule/secrpfsection -> cato_remote_port_forward_rule/_section
		//   branch side: rpfrule/rpfsection       -> cato_remote_port_fwd_rule/_section
		// Likewise cato_private_access_section (tag) and cato_pa_section (branch) coexist.
		// One naming must be retired before GA.
		// For the 13 type names implemented on BOTH sides (client connectivity,
		// dynamic ip allocation, sdp static ip allocation, split tunnel, terminal
		// server, anti-malware file hash rule+section,
		// site web proxy section) the BRANCH implementations are registered.
		// Socket Bypass retains its established schema and implementation.
		// The other implementations under policy/access, security and
		// resource/sitewebproxy remain unregistered pending review.
		secrpfrule.NewResource,
		secrpfsection.NewResource,
		alwaysonindex.NewAlwaysOnRuleBulkResource,
		alwaysonpolicy.NewAlwaysOnPolicyResource,
		alwaysonsection.NewAlwaysOnSectionResource,
		bgppeer.NewBgpPeerResource,
		group.NewGroupResource,
		ifrule.NewInternetFwRuleResource,
		ifsection.NewInternetFwSectionResource,
		ifsubpolicy.NewIfSubPolicyResource,
		laninterface.NewLanInterfaceResource,
		laninterfacelagmember.NewLanInterfaceLagMemberResource,
		license.NewLicenseResource,
		networkrange.NewNetworkRangeResource,
		allocatedipres.NewAllocatedIPResource,
		groupmembers.NewGroupMembersResource,
		ipsecsite.NewSiteIpsecResource,
		socketsite.NewSocketSiteResource,
		statichost.NewStaticHostResource,
		tlsrule.NewTLSInspectionRuleResource,
		tlssection.NewTLSInspectionSectionResource,
		wfrule.NewWanFwRuleResource,
		wfsection.NewWanFwSectionResource,
		wfsubpolicy.NewWfSubPolicyResource,
		waninterface.NewWanInterfaceResource,
		wnwrule.NewWanNetworkRuleResource,
		wnwsection.NewWanNetworkSectionResource,
		ifindex.NewIfwRulesIndexResource,
		wanindex.NewWanRulesIndexResource,
		wnwindex.NewWanNetworkRulesIndexResource,
		tlsindex.NewTLSRulesIndexResource,
		appconnector.NewAppConnectorResource,
		privateapp.NewPrivateAppResource,
		papolicy.NewPrivAccessPolicyResource,
		parule.NewPrivAccessRuleResource,
		parulebulk.NewPrivAccessRuleBulkResource,
		atrule.NewAppTenantRestrictionRuleResource,
		atindex.NewAtrRulesIndexResource,
		atsection.NewAppTenantRestrictionSectionResource,
		acpolicy.NewApplicationControlPolicyResource,
		tsindex.NewTsRulesIndexResource,
		tspolicy.NewTerminalServerPolicyResource,
		amfhindex.NewAmfhRulesIndexResource,
		amfhpolicy.NewAntiMalwareFileHashPolicyResource,
		amfhrule.NewAntiMalwareFileHashRuleResource,
		amfhsection.NewAntiMalwareFileHashSectionResource,
		ccrule.NewClientConnectivityRuleResource,
		ccsection.NewClientConnectivitySectionResource,
		diarule.NewDynamicIPAllocationRuleResource,
		diasection.NewDynamicIPAllocationSectionResource,
		ssiarule.NewSdpStaticIPAllocationRuleResource,
		ssiasection.NewSdpStaticIPAllocationSectionResource,
		swpsection.NewSiteWebProxySectionResource,
		strule.NewSplitTunnelRuleResource,
		stsection.NewSplitTunnelSectionResource,
		tsrule.NewTerminalServerRuleResource,
		tssection.NewTerminalServerSectionResource,
		ssiaindex.NewSsiaRulesIndexResource,
		ssiapolicy.NewSdpStaticIPAllocationPolicyResource,
		diaindex.NewDiaRulesIndexResource,
		diapolicy.NewDynamicIPAllocationPolicyResource,
		ccindex.NewCcRulesIndexResource,
		ccpolicy.NewClientConnectivityPolicyResource,
		stindex.NewStRulesIndexResource,
		stpolicy.NewSplitTunnelPolicyResource,
		rpfrule.NewRemotePortFwdRuleResource,
		rpfsection.NewRemotePortFwdSectionResource,
		rpfindex.NewRpfRulesIndexResource,
		rpfpolicy.NewRemotePortFwdPolicyResource,
		swprule.NewSiteWebProxyRuleResource,
		swpindex.NewSwpRulesIndexResource,
		swppolicy.NewSiteWebProxyPolicyResource,
		sbindex.NewSbRulesIndexResource,
		sbpolicy.NewSocketBypassPolicyResource,
		atpolicy.NewAppTenantRestrictionPolicyResource,
		wnwpolicy.NewWanNetworkPolicyResource,
		pasection.NewPrivateAccessSectionResource,
		acrule.NewApplicationControlRuleResource,
		acindex.NewApcRulesIndexResource,
		acsection.NewApplicationControlSectionResource,
		lansection.NewSocketLanSectionResource,
		socketlannetworkrule.NewSocketLanNetworkRuleResource,
		socketlanfirewallrule.NewSocketLanFirewallRuleResource,
		globalipranges.NewGlobalIPRangesResource,
		lfsubpolicy.NewLfSubPolicyResource,
		lanindex.NewLanRulesIndexResource,
	}
}
