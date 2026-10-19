package appconnector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type GroupDataSourceModel struct {
	Items []types.String `tfsdk:"items"`
	Total types.Int64    `tfsdk:"total"`
}

func GroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	client *catoclient.Service
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_connector_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "AppConnector group schema definition",
		Attributes: map[string]schema.Attribute{
			"items": schema.ListAttribute{
				Description: "AppConnector groups",
				Computed:    true,
				ElementType: types.StringType,
			},
			"total": schema.Int64Attribute{
				Description: "Total number of groups",
				Computed:    true,
			},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var groups GroupDataSourceModel
	if diags := req.Config.Get(ctx, &groups); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	result, err := d.client.Catov2.AppConnectorReadGroups(ctx, d.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch app-connector groups", err.Error())
		return
	}

	for _, g := range result.GetZtnaAppConnector().GetZtnaAppConnectorGroupList().GetItems() {
		groups.Items = append(groups.Items, types.StringValue(g))
	}
	groups.Total = types.Int64Value(result.GetZtnaAppConnector().GetZtnaAppConnectorGroupList().GetPageInfo().GetTotal())

	resp.Diagnostics.Append(resp.State.Set(ctx, &groups)...)
	if diags := resp.State.Set(ctx, &groups); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
