package provider

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/adapters/catoapi"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/adapters/memory"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources"
	privateaccesspolicyrule "github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources/policy/privateaccess/rule"
	statichost "github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources/static_host"
)

var _ provider.Provider = (*catoProvider)(nil)

// NewProvider initializes the standalone POC provider. No production provider is wrapped
// or configured, and configuration never mutates or discards policy revisions.
func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider { return &catoProvider{version: version} }
}

type catoProvider struct {
	version string
}

type configModel struct {
	BaseURL   types.String `tfsdk:"baseurl"`
	Token     types.String `tfsdk:"token"`
	AccountID types.String `tfsdk:"account_id"`
}

func (p *catoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cato"
	resp.Version = p.version
}

func (p *catoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Proof of concept for private-access rule deletion and static-host creation.",
		Attributes: map[string]schema.Attribute{
			"baseurl":    schema.StringAttribute{Optional: true, Description: "Cato GraphQL HTTPS endpoint. Defaults to CATO_BASEURL."},
			"token":      schema.StringAttribute{Optional: true, Sensitive: true, Description: "API token. Defaults to CATO_TOKEN."},
			"account_id": schema.StringAttribute{Required: true, Description: "Cato account containing the resources."},
		},
	}
}

func (p *catoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config configModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for name, value := range map[string]types.String{
		"baseurl": config.BaseURL, "token": config.Token, "account_id": config.AccountID,
	} {
		if value.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root(name), "Unknown provider configuration",
				name+" must be known before configuring the provider.")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	baseURL := configurationValue(config.BaseURL, "CATO_BASEURL")
	token := configurationValue(config.Token, "CATO_TOKEN")
	accountID := config.AccountID.ValueString()
	for name, value := range map[string]string{"baseurl": baseURL, "token": token, "account_id": accountID} {
		if strings.TrimSpace(value) == "" {
			resp.Diagnostics.AddAttributeError(path.Root(name), "Missing provider configuration", name+" must be configured.")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint, err := url.Parse(baseURL)
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.User != nil {
		resp.Diagnostics.AddAttributeError(path.Root("baseurl"), "Invalid Cato endpoint", "Use an HTTPS URL without embedded credentials.")
		return
	}
	const requestTimeout = 30 * time.Second
	client, err := cato.New(baseURL, token, accountID, &http.Client{Timeout: requestTimeout}, map[string]string{
		"User-Agent": "cato-terraform-refactor/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to initialize Cato client", "The API client could not be initialized.")
		return
	}
	resp.ResourceData = &resources.Dependencies{
		CatoAPI:      catoapi.NewAdapter(client),
		AccountID:    accountID,
		ResourceLock: memory.NewResourceLock(),
	}
}

func configurationValue(value types.String, environment string) string {
	if value.IsNull() {
		return os.Getenv(environment)
	}
	return value.ValueString()
}

func (p *catoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		privateaccesspolicyrule.NewResource,
		statichost.NewResource,
	}
}

func (p *catoProvider) DataSources(_ context.Context) []func() datasource.DataSource { return nil }
