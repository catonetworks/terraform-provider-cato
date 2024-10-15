package provider

import (
	"context"
	"net/http"
	"os"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &catoProvider{}
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
	BaseURL   types.String `tfsdk:"baseurl"`
	Token     types.String `tfsdk:"token"`
	AccountId types.String `tfsdk:"account_id"`
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

	if resp.Diagnostics.HasError() {
		return
	}

	baseurl := os.Getenv("CATO_BASEURL")
	token := os.Getenv("CATO_TOKEN")

	if !config.BaseURL.IsNull() {
		baseurl = config.BaseURL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
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

	// newer client:
	catoClient, _ := cato.New(baseurl, token, *http.DefaultClient, "User-Agent", "cato-terraform-"+p.version)

	dataSourceData := &catoClientData{
		BaseURL:   baseurl,
		Token:     token,
		AccountId: accountId,
		catov2:    catoClient,
	}

	resp.DataSourceData = dataSourceData
	resp.ResourceData = dataSourceData

}

func (p *catoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountSnapshotSiteDataSource,
		// NewInternetFwPolicyDataSource,
	}
}

func (p *catoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSocketSiteResource,
		NewWanInterfaceResource,
		NewInternetFwRuleResource,
		NewInternetFwSectionResource,
		NewWanFwRuleResource,
		NewWanFwSectionResource,
		// NewSiteIpsecResource,
		// NewSiteIpsecTunnelResource,
		NewStaticHostResource,
		NewNetworkRangeResource,
	}
}
