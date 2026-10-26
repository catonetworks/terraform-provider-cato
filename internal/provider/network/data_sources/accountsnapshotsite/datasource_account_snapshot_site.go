package accountsnapshotsite

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ datasource.DataSource              = &accountSnapshotSiteDataSource{}
	_ datasource.DataSourceWithConfigure = &accountSnapshotSiteDataSource{}
)

func NewAccountSnapshotSiteDataSource() datasource.DataSource {
	return &accountSnapshotSiteDataSource{}
}

type accountSnapshotSiteDataSource struct {
	client *catoclient.Service
}

func (d *accountSnapshotSiteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accountSnapshotSite"
}

func (d *accountSnapshotSiteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves account snapshot site socket serial number and primary status information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for the site. Cannot be used with site_name.",
				Optional:    true,
			},
			"site_name": schema.StringAttribute{
				Description: "Name of a site. Cannot be used with id.",
				Optional:    true,
			},
			"info": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Site Name",
						Computed:    true,
					},
					"sockets": schema.ListNestedAttribute{
						Description: "List of sockets attached to the site",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Socket id",
									Computed:    true,
								},
								"serial": schema.StringAttribute{
									Description: "Socket serial number",
									Computed:    true,
								},
								"is_primary": schema.BoolAttribute{
									Description: "Socket is primary",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *accountSnapshotSiteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

//nolint:gocyclo
func (d *accountSnapshotSiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SiteSnapshot
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure exactly one of 'id' or 'site_name' is set
	if state.ID != nil && state.SiteName != nil {
		resp.Diagnostics.AddError(
			"Configuration Error",
			"Only one of 'id' or 'site_name' can be set, not both.",
		)
		return
	}

	if state.ID == nil && state.SiteName == nil {
		resp.Diagnostics.AddError(
			"Configuration Error",
			"Either 'id' or 'site_name' must be specified.",
		)
		return
	}

	accountSnapshotSite, err := d.client.Catov2.AccountSnapshot(ctx, nil, nil, &d.client.AccountId)
	tflog.Debug(ctx, "Read.AccountSnapshot.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(accountSnapshotSite),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// Find the site that matches the provided id or site_name
	var foundSite *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites

	for _, site := range accountSnapshotSite.AccountSnapshot.Sites {
		// Check if site matches by ID
		if state.ID != nil && site.ID != nil && *site.ID == *state.ID {
			foundSite = site
			break
		}
		// Check if site matches by name
		if state.SiteName != nil &&
			site.InfoSiteSnapshot != nil &&
			site.InfoSiteSnapshot.Name != nil &&
			*site.InfoSiteSnapshot.Name == *state.SiteName {
			foundSite = site
			break
		}
	}

	if foundSite == nil {
		if state.ID != nil {
			resp.Diagnostics.AddError(
				"Site Not Found",
				fmt.Sprintf("No site found with ID: %s", *state.ID),
			)
		} else {
			resp.Diagnostics.AddError(
				"Site Not Found",
				fmt.Sprintf("No site found with name: %s", *state.SiteName),
			)
		}
		return
	}

	state = SiteSnapshot{
		ID:       foundSite.ID,
		SiteName: foundSite.InfoSiteSnapshot.Name,
		Info: &SiteInfo{
			Name: foundSite.InfoSiteSnapshot.Name,
		},
	}

	for _, socket := range foundSite.InfoSiteSnapshot.Sockets {
		state.Info.Sockets = append(state.Info.Sockets, SocketInfo{
			ID:        socket.ID,
			Serial:    socket.Serial,
			IsPrimary: socket.IsPrimary,
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
