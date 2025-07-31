package provider

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

type dhcpRelayGroupLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type DhcpRelayGroup struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DhcpRelayServers types.List   `tfsdk:"dhcp_relay_servers"`
}

func DhcpRelayDataSource() datasource.DataSource {
	return &dhcpRelayGroupDataSource{}
}

type dhcpRelayGroupDataSource struct {
	client *catoClientData
}

func (d *dhcpRelayGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcpRelayGroup"
}

func (d *dhcpRelayGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves DHCP relay groups.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of names to filter",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of dhcpRelayGroups",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name",
							Computed:    true,
						},
						"dhcp_relay_servers": schema.ListAttribute{
							Description: "DHCP Relay Servers",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *dhcpRelayGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *dhcpRelayGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var dhcpRelayGroupLookup dhcpRelayGroupLookup
	if diags := req.Config.Get(ctx, &dhcpRelayGroupLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	zeroInt64 := int64(0)
	result, err := d.client.catov2.EntityLookupMinimal(ctx, d.client.AccountId, cato_models.EntityTypeDhcpRelayGroup, &zeroInt64, nil, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	filterByName := !dhcpRelayGroupLookup.NameFilter.IsNull() && dhcpRelayGroupLookup.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range dhcpRelayGroupLookup.NameFilter.Elements() {
			// Trim any quotes if present
			valueStr := strings.Trim(value.String(), "\"")
			namesMap[valueStr] = struct{}{}
		}
	}

	attrTypes := map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"dhcp_relay_servers": types.ListType{
			ElemType: types.StringType,
		},
	}
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		name := cast.ToString(item.GetEntity().GetName())
		if !filterByName || contains(namesMap, name) {
			dhcpRelayServers := strings.Split(item.Description, ",")
			var dhcpRelayServersValues []attr.Value
			for _, s := range dhcpRelayServers {
				dhcpRelayServersValues = append(dhcpRelayServersValues, types.StringValue(strings.TrimSpace(s)))
			}
			dhcpRelayServersList, diags := types.ListValue(types.StringType, dhcpRelayServersValues)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					"id":                 types.StringValue(item.GetEntity().GetID()),
					"name":               types.StringValue(name),
					"dhcp_relay_servers": dhcpRelayServersList,
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			objects = append(objects, obj)
		}
	}

	list, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: attrTypes,
		},
		objects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	dhcpRelayGroupLookup.Items = list
	if diags := resp.State.Set(ctx, &dhcpRelayGroupLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
