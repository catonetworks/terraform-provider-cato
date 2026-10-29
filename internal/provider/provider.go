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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/access/privateapp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/account"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/admin"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/licensinginfo"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/permissioncatalog"
	roleds "github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/role"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/data_sources/roles"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/license"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/rbacassignment"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/account/role"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/accountsnapshotsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/allocatedip"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/dhcprelaygroup"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/host"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/networkinterfaces"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/networkranges"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/data_sources/sitelocation"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/ipsecsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/laninterface"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/laninterfacelagmember"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkrange"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/socketsite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acpolicy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acsection"
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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atsection"
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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wnwrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wnwsection"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/bgppeer"
	groupdata "github.com/catonetworks/terraform-provider-cato/internal/provider/resource/data_sources/group"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/globalipranges"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/group"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/groupmembers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/statichost"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/resource/waninterface"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/security/antimalwarefilehash/amfhrule"
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
				Description: "API Key for the Cato API. Can be provided using CATO_BASEURL environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"account_id": schema.StringAttribute{
				Description: "AccountId for the Cato API",
				Required:    true,
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

	accountID := config.AccountID.ValueString()

	if baseurl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("baseurl"),
			"Missing Cato API Base URL ",
			"The provider cannot create the CATO API client as there is a missing or empty value for the CATO API URL. ",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Cato API Token ",
			"The provider cannot create the CATO API client as there is a missing or empty value for the CATO API Token. ",
		)
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
		dhcprelaygroup.DataSource,
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
		appconnectorgroup.GroupDataSource,
	}
}

func (p *catoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		account.NewAccountResource,
		role.NewResource,
		rbacassignment.NewResource,
		admin.NewAdminResource,
		alwaysonrule.NewAlwaysOnRuleResource,

		amfhrule.NewResource,
		bgppeer.NewBgpPeerResource,
		group.NewGroupResource,
		ifrule.NewInternetFwRuleResource,
		ifsection.NewInternetFwSectionResource,
		ifsubpolicy.NewIfSubPolicyResource,
		laninterface.NewLanInterfaceResource,
		laninterfacelagmember.NewLanInterfaceLagMemberResource,
		license.NewLicenseResource,
		networkrange.NewNetworkRangeResource,
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
		atsection.NewAppTenantRestrictionSectionResource,
		acpolicy.NewApplicationControlPolicyResource,
		acrule.NewApplicationControlRuleResource,
		acsection.NewApplicationControlSectionResource,
		lansection.NewSocketLanSectionResource,
		socketlannetworkrule.NewSocketLanNetworkRuleResource,
		socketlanfirewallrule.NewSocketLanFirewallRuleResource,
		globalipranges.NewGlobalIPRangesResource,
		lfsubpolicy.NewLfSubPolicyResource,
		lanindex.NewLanRulesIndexResource,
	}
}
