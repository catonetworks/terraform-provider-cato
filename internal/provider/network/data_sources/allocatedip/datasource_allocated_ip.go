package allocatedip

import (
	"context"
	"fmt"
	"strings"

	cato "github.com/catonetworks/cato-go-sdk"
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

func Source() datasource.DataSource {
	return &allocatedIPDataSource{}
}

type allocatedIPDataSource struct {
	client *catoclient.Service
}

func (d *allocatedIPDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocatedIp"
}

func (d *allocatedIPDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves account allocated IPs. Filters compose (AND) and every filter is " +
			"optional: name_filter matches the allocated IP address (back-compat); " +
			"description_filter + pop_location_filter select by the stable cross-tenant " +
			"identity (the display name embeds the Cato-assigned IP, so names differ per tenant).",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of allocated IP addresses to filter (matches the `name` item field)",
				Required:    false,
				Optional:    true,
			},
			"description_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of allocation descriptions to filter (stable cross-tenant identity, combined with pop_location_filter)",
				Required:    false,
				Optional:    true,
			},
			"pop_location_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of PoP location names to filter (matches the `location` item field)",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of allocatedIps",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Allocated IP address",
							Computed:    true,
						},
						"location": schema.StringAttribute{
							Description: "Pop Location",
							Computed:    true,
						},
						"cc": schema.StringAttribute{
							Description: "Country code",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Allocation description (stable across tenants)",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "CMA display name (\"<pop> - <ip>\") -- the form policy rules reference",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *allocatedIPDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

// listFilterSet collects a Terraform list filter into a lookup set; returns
// (set, active). An absent/null filter is inactive and matches everything.
func listFilterSet(l types.List) (map[string]struct{}, bool) {
	if l.IsNull() || l.Elements() == nil {
		return nil, false
	}
	set := make(map[string]struct{})
	for _, value := range l.Elements() {
		set[strings.Trim(value.String(), "\"")] = struct{}{}
	}
	return set, true
}

func (d *allocatedIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var allocatedIPLookup Lookup
	if diags := req.Config.Get(ctx, &allocatedIPLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	zeroInt64 := int64(0)
	result, err := d.client.Catov2.EntityLookup(
		ctx, d.client.AccountId, cato_models.EntityTypeAllocatedIP, &zeroInt64, nil, nil, nil, nil, nil, nil, nil,
	)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	// EntityLookup has no description field; join PopLocationsAllocatedIPList
	// (keyed by ip) to enrich items with the stable allocation description.
	listResult, err := d.client.Catov2.PopLocationsAllocatedIPList(ctx, d.client.AccountId,
		&cato_models.PopLocationAllocatedIPInput{})
	tflog.Debug(ctx, "Read.PopLocationsAllocatedIPList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(listResult),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PopLocationsAllocatedIPList error", err.Error())
		return
	}
	descByIP := allocationDescriptions(listResult)

	namesSet, filterByName := listFilterSet(allocatedIPLookup.NameFilter)
	descSet, filterByDesc := listFilterSet(allocatedIPLookup.DescriptionFilter)
	popSet, filterByPop := listFilterSet(allocatedIPLookup.PopLocationFilter)

	attrTypes := map[string]attr.Type{
		"id":           types.StringType,
		"name":         types.StringType,
		"location":     types.StringType,
		"cc":           types.StringType,
		"description":  types.StringType,
		"display_name": types.StringType,
	}
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		ip := cast.ToString(helperFields["allocatedIp"])
		location := cast.ToString(helperFields["popLocation"])
		description := descByIP[ip]
		if filterByName && !utils.Contains(namesSet, ip) {
			continue
		}
		if filterByDesc && !utils.Contains(descSet, description) {
			continue
		}
		if filterByPop && !utils.Contains(popSet, location) {
			continue
		}
		cc := cast.ToString(helperFields["countryCode"])
		obj, diags := types.ObjectValue(
			attrTypes,
			map[string]attr.Value{
				"id":           types.StringValue(item.GetEntity().GetID()),
				"name":         types.StringValue(ip),
				"location":     types.StringValue(location),
				"cc":           types.StringValue(cc),
				"description":  types.StringValue(description),
				"display_name": types.StringValue(fmt.Sprintf("%s - %s", location, ip)),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		objects = append(objects, obj)
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

	allocatedIPLookup.Items = list
	if diags := resp.State.Set(ctx, &allocatedIPLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func allocationDescriptions(result *cato.PopLocationsAllocatedIPList) map[string]string {
	descByIP := make(map[string]string)
	if list := result.GetPopLocations().GetAllocatedIPList(); list != nil {
		for _, item := range list.AllocatedIP {
			if item == nil {
				continue
			}
			if desc := item.GetDescription(); desc != nil {
				descByIP[item.GetIP()] = *desc
			}
		}
	}

	return descByIP
}
