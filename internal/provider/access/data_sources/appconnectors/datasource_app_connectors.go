package appconnectors

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// terraform field names
const (
	tfNameFilter      = "name_filter"
	tfGroupFilter     = "group_filter"
	tfItems           = "items"
	tfID              = "id"
	tfName            = "name"
	tfDescription     = "description"
	tfGroupName       = "group_name"
	tfType            = "type"
	tfIsRegistered    = "is_registered"
	tfPrivateAppCount = "private_app_count"
	tfSerialNumber    = "serial_number"
	tfSocketID        = "socket_id"
)

// pageLimit is the page size used when paginating ztnaAppConnectorList to exhaustion.
const pageLimit = 1000

var attrTypes = map[string]attr.Type{
	tfID:              types.StringType,
	tfName:            types.StringType,
	tfDescription:     types.StringType,
	tfGroupName:       types.StringType,
	tfType:            types.StringType,
	tfIsRegistered:    types.BoolType,
	tfPrivateAppCount: types.Int64Type,
	tfSerialNumber:    types.StringType,
	tfSocketID:        types.StringType,
}

func DataSource() datasource.DataSource {
	return &appConnectorsDataSource{}
}

type appConnectorsDataSource struct {
	client *catoclient.Service
}

func (d *appConnectorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_connectors"
}

func (d *appConnectorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves ZTNA app connectors.",
		Attributes: map[string]schema.Attribute{
			tfNameFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of app connector names to filter",
				Optional:    true,
			},
			tfGroupFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of app connector group names to filter",
				Optional:    true,
			},

			tfItems: schema.ListNestedAttribute{
				Description: "List of app connectors",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						tfID: schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						tfName: schema.StringAttribute{
							Description: "Name",
							Computed:    true,
						},
						tfDescription: schema.StringAttribute{
							Description: "Description",
							Computed:    true,
						},
						tfGroupName: schema.StringAttribute{
							Description: "App connector group name",
							Computed:    true,
						},
						tfType: schema.StringAttribute{
							Description: "Connector type",
							Computed:    true,
						},
						tfIsRegistered: schema.BoolAttribute{
							Description: "Whether the connector is registered",
							Computed:    true,
						},
						tfPrivateAppCount: schema.Int64Attribute{
							Description: "Number of private applications served by the connector",
							Computed:    true,
						},
						tfSerialNumber: schema.StringAttribute{
							Description: "Socket serial number",
							Computed:    true,
						},
						tfSocketID: schema.StringAttribute{
							Description: "Socket ID",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *appConnectorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *appConnectorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var lookup Lookup
	if diags := req.Config.Get(ctx, &lookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	nameFilter := makeSet(lookup.NameFilter)
	groupFilter := makeSet(lookup.GroupFilter)
	var objects []attr.Value

	from := int64(0)
	for {
		result, err := d.client.Catov2.ZtnaAppConnectorZtnaAppConnectorList(ctx, d.client.AccountId,
			cato_models.ZtnaAppConnectorListInput{
				Paging: &cato_models.PagingInput{Limit: pageLimit, From: from},
				// sort is a non-nullable input on this API even though every
				// field inside it is optional -- pass an empty object.
				Sort: &cato_models.ZtnaAppConnectorListSortInput{},
			})
		tflog.Debug(ctx, "Read.ZtnaAppConnectorZtnaAppConnectorList.response", map[string]interface{}{
			"from":     from,
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API ZtnaAppConnectorList error", err.Error())
			return
		}

		list := result.GetZtnaAppConnector().GetZtnaAppConnectorList()
		if list == nil {
			break
		}
		page := list.ZtnaAppConnector
		for _, item := range page {
			if item == nil {
				continue
			}
			if !matchesFilters(item.Name, item.GroupName, nameFilter, groupFilter) {
				continue
			}

			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					tfID:              types.StringValue(item.ID),
					tfName:            types.StringValue(item.Name),
					tfDescription:     types.StringValue(strValue(item.Description)),
					tfGroupName:       types.StringValue(item.GroupName),
					tfType:            types.StringValue(string(item.Type)),
					tfIsRegistered:    types.BoolValue(item.IsRegistered),
					tfPrivateAppCount: types.Int64Value(item.PrivateAppCount),
					tfSerialNumber:    types.StringValue(strValue(item.SerialNumber)),
					tfSocketID:        types.StringValue(strValue(item.SocketID)),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			objects = append(objects, obj)
		}

		from += int64(len(page))
		total := list.GetPageInfo().GetTotal()
		if len(page) == 0 || from >= total {
			break
		}
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	lookup.Items = list
	if diags := resp.State.Set(ctx, &lookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

// makeSet builds a lookup set from a Terraform list; nil = no filter.
func makeSet(value types.List) map[string]struct{} {
	if value.IsNull() {
		return nil
	}
	set := make(map[string]struct{})
	for _, element := range value.Elements() {
		set[strings.Trim(element.String(), "\"")] = struct{}{}
	}
	return set
}

func strValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func matchesFilters(name, group string, names, groups map[string]struct{}) bool {
	if names != nil {
		if _, ok := names[name]; !ok {
			return false
		}
	}
	if groups != nil {
		if _, ok := groups[group]; !ok {
			return false
		}
	}
	return true
}
