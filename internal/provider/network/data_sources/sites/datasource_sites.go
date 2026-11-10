package sites

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// terraform field names
const (
	tfNameFilter        = "name_filter"
	tfItems             = "items"
	tfID                = "id"
	tfName              = "name"
	tfDescription       = "description"
	tfType              = "type"
	tfConnectionType    = "connection_type"
	tfOperationalStatus = "operational_status"
)

var attrTypes = map[string]attr.Type{
	tfID:                types.StringType,
	tfName:              types.StringType,
	tfDescription:       types.StringType,
	tfType:              types.StringType,
	tfConnectionType:    types.StringType,
	tfOperationalStatus: types.StringType,
}

func DataSource() datasource.DataSource {
	return &sitesDataSource{}
}

type sitesDataSource struct {
	client *catoclient.Service
}

func (d *sitesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

func (d *sitesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves socket sites.",
		Attributes: map[string]schema.Attribute{
			tfNameFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of site names to filter",
				Optional:    true,
			},

			tfItems: schema.ListNestedAttribute{
				Description: "List of sites",
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
						tfType: schema.StringAttribute{
							Description: "Site type (e.g. BRANCH, HEADQUARTERS, CLOUD_DC)",
							Computed:    true,
						},
						tfConnectionType: schema.StringAttribute{
							Description: "Site connection type (e.g. socket model)",
							Computed:    true,
						},
						tfOperationalStatus: schema.StringAttribute{
							Description: "Site operational status",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *sitesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *sitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var siteLookup Lookup
	if diags := req.Config.Get(ctx, &siteLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	filter := makeFilter(&siteLookup)
	var objects []attr.Value

	// NOTE: entityLookup(type=site) treats limit=0 literally (returns zero
	// items), unlike type=host where 0 means "all". Page explicitly with a
	// non-zero limit + from offset and accumulate until total is reached, so
	// large accounts aren't silently truncated.
	pageSize := int64(1000)
	from := int64(0)
	for {
		result, err := d.client.Catov2.EntityLookup(
			ctx, d.client.AccountId, cato_models.EntityTypeSite, &pageSize, &from, nil, nil, nil, nil, nil, nil,
		)
		tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
			"from":     from,
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
			return
		}

		page := result.GetEntityLookup().GetItems()
		for _, item := range page {
			entity := item.GetEntity()
			name := cast.ToString(entity.Name)
			if !filter.accept(name) {
				continue
			}

			helperFields := item.GetHelperFields()
			siteType := cast.ToString(helperFields["type"])
			connectionType := cast.ToString(helperFields["connectionType"])
			operationalStatus := cast.ToString(helperFields["operationalStatus"])

			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					tfID:                types.StringValue(entity.ID),
					tfName:              types.StringValue(name),
					tfDescription:       types.StringValue(item.GetDescription()),
					tfType:              types.StringValue(siteType),
					tfConnectionType:    types.StringValue(connectionType),
					tfOperationalStatus: types.StringValue(operationalStatus),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			objects = append(objects, obj)
		}

		from += int64(len(page))
		total := result.GetEntityLookup().GetTotal()
		if len(page) == 0 || (total != nil && from >= *total) {
			break
		}
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	siteLookup.Items = list
	if diags := resp.State.Set(ctx, &siteLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

// makeFilter prepares a lookup map for names;
// if IsNull() then the map is nil
func makeFilter(siteLookup *Lookup) *siteFilter {
	var f siteFilter

	if !siteLookup.NameFilter.IsNull() {
		f.namesMap = make(map[string]struct{})
		for _, value := range siteLookup.NameFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"") // Trim quotes
			f.namesMap[valueStr] = struct{}{}
		}
	}
	return &f
}

// accept returns true if the item should be returned
func (f *siteFilter) accept(name string) bool {
	if f.namesMap == nil {
		return true // no filter applied
	}
	_, ok := f.namesMap[name]
	return ok
}
