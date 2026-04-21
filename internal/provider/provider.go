package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &catoProvider{}
)

const (
	defaultRetryMax            int64 = 5
	defaultRetryWaitMinSeconds int64 = 1
	defaultRetryWaitMaxSeconds int64 = 30
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &catoProvider{
			version: version,
		}
	}
}

type catoProvider struct {
	version string
}

type catoProviderModel struct {
	BaseURL             types.String `tfsdk:"baseurl"`
	Token               types.String `tfsdk:"token"`
	AccountId           types.String `tfsdk:"account_id"`
	RetryMax            types.Int64  `tfsdk:"retry_max"`
	RetryWaitMinSeconds types.Int64  `tfsdk:"retry_wait_min_seconds"`
	RetryWaitMaxSeconds types.Int64  `tfsdk:"retry_wait_max_seconds"`
}

// added by JF to support use of two different clients (long story....)
type catoClientData struct {
	BaseURL   string
	Token     string
	AccountId string
	catov2    *cato.Client
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
				Description: "Maximum number of retries for retryable API requests. Defaults to 5. Can be provided using CATO_RETRY_MAX environment variable.",
				Optional:    true,
			},
			"retry_wait_min_seconds": schema.Int64Attribute{
				Description: "Minimum backoff between retry attempts, in seconds. Defaults to 1. Can be provided using CATO_RETRY_WAIT_MIN_SECONDS environment variable.",
				Optional:    true,
			},
			"retry_wait_max_seconds": schema.Int64Attribute{
				Description: "Maximum backoff between retry attempts, in seconds. Defaults to 30. Can be provided using CATO_RETRY_WAIT_MAX_SECONDS environment variable.",
				Optional:    true,
			},
		},
	}
}

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

	if config.AccountId.IsUnknown() {
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

	if resp.Diagnostics.HasError() {
		return
	}

	baseurl := os.Getenv("CATO_BASEURL")
	token := os.Getenv("CATO_TOKEN")
	retryMax, retryMaxErr := int64FromEnv("CATO_RETRY_MAX")
	retryWaitMinSeconds, retryWaitMinErr := int64FromEnv("CATO_RETRY_WAIT_MIN_SECONDS")
	retryWaitMaxSeconds, retryWaitMaxErr := int64FromEnv("CATO_RETRY_WAIT_MAX_SECONDS")

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

	accountId := config.AccountId.ValueString()

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

	if resp.Diagnostics.HasError() {
		return
	}

	// newer client:
	headers := map[string]string{}
	headers["User-Agent"] = "cato-terraform-" + p.version
	retryConfig := buildRetryConfig(retryMax, retryWaitMinSeconds, retryWaitMaxSeconds)
	httpClient := buildRetryHTTPClient(retryConfig)
	catoClient, err := cato.New(baseurl, token, accountId, httpClient, headers)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Cato API Client",
			err.Error(),
		)
		return
	}

	dataSourceData := &catoClientData{
		BaseURL:   baseurl,
		Token:     token,
		AccountId: accountId,
		catov2:    catoClient,
	}

	resp.DataSourceData = dataSourceData
	resp.ResourceData = dataSourceData

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

type retryClientConfig struct {
	retryMax     int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
}

type retryableResponseErrors struct {
	Errors        []retryableResponseError `json:"errors"`
	NetworkErrors []retryableResponseError `json:"networkErrors"`
	GraphQLErrors []retryableResponseError `json:"graphqlErrors"`
}

type retryableResponseError struct {
	Message string `json:"message"`
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
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy
	retryClient.RetryMax = retryConfig.retryMax
	retryClient.RetryWaitMin = retryConfig.retryWaitMin
	retryClient.RetryWaitMax = retryConfig.retryWaitMax

	return retryClient.StandardClient()
}

func (p *catoProvider) cleanupDrafts(ctx context.Context, d *catoClientData) {
	if os.Getenv("DISABLE_POLICY_RULE_CLEANUP") == "true" {
		return
	}
	resp, err := d.catov2.PolicyPrivateAccessDiscardRevision(ctx, d.AccountId)
	if err != nil {
		tflog.Error(ctx, "failed to discard draft private-access policy", map[string]any{"err": err})
		return
	}
	errors := resp.GetPolicy().GetPrivateAccess().DiscardPolicyRevision.Errors
	if len(errors) > 0 {
		if errors[0].ErrorCode != nil && *errors[0].ErrorCode == "PolicyRevisionNotFound" {
			return // no policy draft to discard; OK
		}
		tflog.Error(ctx, "failed to discard draft private-access policy", map[string]any{"errors": errors})
	}
}

func (p *catoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountSnapshotSiteDataSource,
		AllocatedIpDataSource,
		DhcpRelayDataSource,
		GroupDataSource,
		LicensingInfoDataSource,
		NetworkInterfacesDataSource,
		SiteLocationDataSource,
		IfwRulesIndexDataSource,
		WanRulesIndexDataSource,
		TlsRulesIndexDataSource,
		IfRuleSectionsDataSource,
		WfRuleSectionsDataSource,
		NetworkRangesDataSource,
		HostDataSource,
		AppConnectorGroupDataSource,
	}
}

func (p *catoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccountResource,
		NewAdminResource,
		NewBgpPeerResource,
		NewGroupResource,
		NewInternetFwRuleResource,
		NewInternetFwSectionResource,
		NewLanInterfaceResource,
		NewLanInterfaceLagMemberResource,
		NewLicenseResource,
		NewNetworkRangeResource,
		NewGroupMembersResource,
		NewSiteIpsecResource,
		NewSocketSiteResource,
		NewStaticHostResource,
		NewTlsInspectionRuleResource,
		NewTlsInspectionSectionResource,
		NewWanFwRuleResource,
		NewWanFwSectionResource,
		NewWanInterfaceResource,
		NewWanNetworkRuleResource,
		NewWanNetworkSectionResource,
		NewIfwRulesIndexResource,
		NewWanRulesIndexResource,
		NewWanNetworkRulesIndexResource,
		NewTlsRulesIndexResource,
		NewAppConnectorResource,
		NewPrivateAppResource,
		NewPrivAccessPolicyResource,
		NewPrivAccessRuleResource,
		NewPrivAccessRuleBulkResource,
		NewSocketLanSectionResource,
		NewSocketLanNetworkRuleResource,
		NewSocketLanFirewallRuleResource,
	}
}
