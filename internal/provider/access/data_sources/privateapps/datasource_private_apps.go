package privateapps

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
	tfNameFilter         = "name_filter"
	tfIDFilter           = "id_filter"
	tfItems              = "items"
	tfID                 = "id"
	tfName               = "name"
	tfDescription        = "description"
	tfInternalAppAddress = "internal_app_address"
	tfAllowICMPProtocol  = "allow_icmp_protocol"
	tfProbingEnabled     = "probing_enabled"
	tfPublished          = "published"
	tfConnectivityStatus = "connectivity_status"
)

// pageLimit is the page size used when paginating privateApplicationList to exhaustion.
const pageLimit = 1000

var attrTypes = map[string]attr.Type{
	tfID:                 types.StringType,
	tfName:               types.StringType,
	tfDescription:        types.StringType,
	tfInternalAppAddress: types.StringType,
	tfAllowICMPProtocol:  types.BoolType,
	tfProbingEnabled:     types.BoolType,
	tfPublished:          types.BoolType,
	tfConnectivityStatus: types.StringType,
}

func DataSource() datasource.DataSource {
	return &privateAppsDataSource{}
}

type privateAppsDataSource struct {
	client *catoclient.Service
}

func (d *privateAppsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_apps"
}

func (d *privateAppsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves private applications (ZTNA private access apps).",
		Attributes: map[string]schema.Attribute{
			tfNameFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of private application names to filter",
				Optional:    true,
			},
			tfIDFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of private application IDs to filter",
				Optional:    true,
			},

			tfItems: schema.ListNestedAttribute{
				Description: "List of private applications",
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
						tfInternalAppAddress: schema.StringAttribute{
							Description: "Internal application address (IPv4 or FQDN)",
							Computed:    true,
						},
						tfAllowICMPProtocol: schema.BoolAttribute{
							Description: "Whether ICMP is allowed to the application",
							Computed:    true,
						},
						tfProbingEnabled: schema.BoolAttribute{
							Description: "Whether connectivity probing is enabled",
							Computed:    true,
						},
						tfPublished: schema.BoolAttribute{
							Description: "Whether the application is published",
							Computed:    true,
						},
						tfConnectivityStatus: schema.StringAttribute{
							Description: "Connectivity status",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *privateAppsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *privateAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var lookup Lookup
	if diags := req.Config.Get(ctx, &lookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	filter := makeFilter(&lookup)
	var objects []attr.Value

	from := 0
	for {
		limit, offset := pageLimit, from
		result, err := d.client.Catov2.PrivateApplicationPrivateApplicationList(ctx, d.client.AccountId,
			&cato_models.PrivateApplicationListInput{
				Paging: &cato_models.PagingInput{Limit: int64(limit), From: int64(offset)},
			})
		tflog.Debug(ctx, "Read.PrivateApplicationPrivateApplicationList.response", map[string]interface{}{
			"from":     from,
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PrivateApplicationList error", err.Error())
			return
		}

		list := result.GetPrivateApplication().GetPrivateApplicationList()
		if list == nil {
			break
		}
		page := list.Applications
		for _, item := range page {
			if item == nil {
				continue
			}
			if !filter.accept(item.Name, item.ID) {
				continue
			}
			description := ""
			if item.Description != nil {
				description = *item.Description
			}
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					tfID:                 types.StringValue(item.ID),
					tfName:               types.StringValue(item.Name),
					tfDescription:        types.StringValue(description),
					tfInternalAppAddress: types.StringValue(item.InternalAppAddress),
					tfAllowICMPProtocol:  types.BoolValue(item.AllowICMPProtocol),
					tfProbingEnabled:     types.BoolValue(item.ProbingEnabled),
					tfPublished:          types.BoolValue(item.Published),
					tfConnectivityStatus: types.StringValue(string(item.ConnectivityStatus)),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			objects = append(objects, obj)
		}

		from += len(page)
		total := list.GetPageInfo().GetTotal()
		if len(page) == 0 || int64(from) >= total {
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

// makeFilter prepares lookup maps for names and ids; nil map = no filter.
func makeFilter(lookup *Lookup) *appFilter {
	var f appFilter
	if !lookup.NameFilter.IsNull() {
		f.namesMap = make(map[string]struct{})
		for _, value := range lookup.NameFilter.Elements() {
			f.namesMap[strings.Trim(value.String(), "\"")] = struct{}{}
		}
	}
	if !lookup.IDFilter.IsNull() {
		f.idsMap = make(map[string]struct{})
		for _, value := range lookup.IDFilter.Elements() {
			f.idsMap[strings.Trim(value.String(), "\"")] = struct{}{}
		}
	}
	return &f
}

type appFilter struct {
	namesMap map[string]struct{}
	idsMap   map[string]struct{}
}

// accept returns true if the item passes the configured filters.
func (f *appFilter) accept(name, id string) bool {
	if f.namesMap != nil {
		if _, ok := f.namesMap[name]; !ok {
			return false
		}
	}
	if f.idsMap != nil {
		if _, ok := f.idsMap[id]; !ok {
			return false
		}
	}
	return true
}
