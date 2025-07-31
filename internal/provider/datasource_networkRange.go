package provider

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

func NetworkRangesDataSource() datasource.DataSource {
	return &networkRangesDataSource{}
}

type networkRangesDataSource struct {
	client *catoClientData
}

func (d *networkRangesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networkRanges"
}

func (d *networkRangesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"site_id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of site IDs to filter",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of network range names to filter",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of network ranges",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						"interface_name": schema.StringAttribute{
							Description: "Interface Name",
							Computed:    true,
						},
						"mdns_reflector": schema.BoolAttribute{
							Description: "mDNS Reflector enabled",
							Computed:    true,
						},
						"microsegmentation": schema.BoolAttribute{
							Description: "Microsegmentation enabled",
							Computed:    true,
						},
						"site_id": schema.StringAttribute{
							Description: "Site ID",
							Computed:    true,
						},
						"site_name": schema.StringAttribute{
							Description: "Site Name",
							Computed:    true,
						},
						"subnet": schema.StringAttribute{
							Description: "Subnet",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *networkRangesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *networkRangesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var networkRangesDataSource NetworkRangeLookup
	if diags := req.Config.Get(ctx, &networkRangesDataSource); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	zeroInt64 := int64(0)
	result, err := d.client.catov2.EntityLookupMinimal(ctx, d.client.AccountId, cato_models.EntityTypeSiteRange, &zeroInt64, nil, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	filterByName := !networkRangesDataSource.NameFilter.IsNull() && networkRangesDataSource.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range networkRangesDataSource.NameFilter.Elements() {
			// Trim any quotes if present
			valueStr := strings.Trim(value.String(), "\"")
			namesMap[valueStr] = struct{}{}
		}
	}

	filterBySiteId := !networkRangesDataSource.SiteIdFilter.IsNull() && networkRangesDataSource.SiteIdFilter.Elements() != nil
	siteIdsMap := make(map[string]struct{})
	if filterBySiteId {
		for _, value := range networkRangesDataSource.SiteIdFilter.Elements() {
			// Trim any quotes if present
			valueStr := strings.Trim(value.String(), "\"")
			siteIdsMap[valueStr] = struct{}{}
		}
	}

	attrTypes := map[string]attr.Type{
		"id":                types.StringType,
		"interface_name":    types.StringType,
		"mdns_reflector":    types.BoolType,
		"microsegmentation": types.BoolType,
		"site_id":           types.StringType,
		"site_name":         types.StringType,
		"subnet":            types.StringType,
		"name":              types.StringType,
	}
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		// Split the interface_name value on " \\ " and pop the last element as name
		interfaceNameRaw := cast.ToString(item.GetEntity().GetName())
		parts := strings.Split(interfaceNameRaw, " \\ ")
		nameValue := ""
		if len(parts) > 0 {
			nameValue = parts[len(parts)-1]
			parts = parts[:len(parts)-1]
		}
		helperFields := item.GetHelperFields()
		siteId := types.StringValue(cast.ToString(helperFields["siteId"]))

		// Check filters: combined logic for name and site ID filters
		// If both siteIds and names are specified, both must match
		// If only names are specified (no siteIds), only name filter applies
		var shouldInclude bool
		if filterBySiteId && filterByName {
			// Both filters specified - both must match
			nameMatches := contains(namesMap, nameValue) || contains(namesMap, cast.ToString(helperFields["interfaceName"]))
			siteIdMatches := contains(siteIdsMap, cast.ToString(helperFields["siteId"]))
			shouldInclude = nameMatches && siteIdMatches
		} else if filterByName {
			// Only name filter specified
			shouldInclude = contains(namesMap, nameValue) || contains(namesMap, cast.ToString(helperFields["interfaceName"]))
		} else if filterBySiteId {
			// Only site ID filter specified
			shouldInclude = contains(siteIdsMap, cast.ToString(helperFields["siteId"]))
		} else {
			// No filters specified - include all
			shouldInclude = true
		}

		if shouldInclude {
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					"id":                types.StringValue(item.GetEntity().GetID()),
					"interface_name":    types.StringValue(cast.ToString(helperFields["interfaceName"])),
					"mdns_reflector":    types.BoolValue(cast.ToBool(helperFields["mdnsReflector"])),
					"microsegmentation": types.BoolValue(cast.ToBool(helperFields["microsegmentation"])),
					"site_id":           siteId,
					"site_name":         types.StringValue(cast.ToString(helperFields["siteName"])),
					"subnet":            types.StringValue(cast.ToString(helperFields["subnet"])),
					"name":              types.StringValue(nameValue),
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

	networkRangesDataSource.Items = list
	if diags := resp.State.Set(ctx, &networkRangesDataSource); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
