package globalipranges

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func DataSource() datasource.DataSource {
	return &globalIPRangesDataSource{}
}

type globalIPRangesDataSource struct {
	client *catoclient.Service
}

func (d *globalIPRangesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_ip_ranges"
}

func (d *globalIPRangesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_global_ip_ranges` data source fetches information about global IP ranges with " +
			"optional filters. Returns a list of global IP ranges matching the filter criteria.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of global IP range IDs to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of global IP range names to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of global IP ranges matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Global IP range ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Global IP range name",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Global IP range description",
							Computed:    true,
						},
						"ip_range": schema.StringAttribute{
							Description: "The IP range value, e.g. \"10.0.0.1-10.0.0.10\"",
							Computed:    true,
						},
						"subnet": schema.StringAttribute{
							Description: "The subnet value, e.g. \"10.0.0.0/24\"",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *globalIPRangesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *globalIPRangesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Lookup
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GlobalIPRangeListInput has no paging field - the API returns the full account's range set in one call.
	listInput := &cato_models.GlobalIPRangeListInput{
		Filter: &cato_models.GlobalIPRangeListFilterInput{},
	}

	listResult, err := d.client.Catov2.ObjectGlobalIPRangeList(ctx, d.client.AccountId, listInput)
	tflog.Debug(ctx, "DataSource.ObjectGlobalIPRangeList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(listResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API ObjectGlobalIPRangeList error",
			err.Error(),
		)
		return
	}

	rangeList := listResult.GetObject().GetGlobalIPRangeList()
	if rangeList == nil {
		resp.Diagnostics.AddError(
			"Cato API ObjectGlobalIPRangeList error",
			"GlobalIPRangeList is nil",
		)
		return
	}

	idsMap, filterByID := filterSet(config.IDFilter)
	namesMap, filterByName := filterSet(config.NameFilter)

	// Process ranges and apply filters
	var rangeObjects []attr.Value

	for _, item := range rangeList.GetItems() {
		var shouldInclude bool
		switch {
		case filterByID && filterByName:
			idMatches := utils.Contains(idsMap, item.GetID())
			nameMatches := utils.Contains(namesMap, item.GetName())
			shouldInclude = idMatches && nameMatches
		case filterByID:
			shouldInclude = utils.Contains(idsMap, item.GetID())
		case filterByName:
			shouldInclude = utils.Contains(namesMap, item.GetName())
		default:
			shouldInclude = true
		}

		if !shouldInclude {
			continue
		}

		rangeObj, diags := types.ObjectValue(
			ItemAttrTypes,
			map[string]attr.Value{
				"id":          types.StringValue(item.GetID()),
				"name":        types.StringValue(item.GetName()),
				"description": types.StringValue(cast.ToString(item.GetDescription())),
				"ip_range":    types.StringValue(item.GetIPRange()),
				"subnet":      types.StringValue(cast.ToString(item.GetSubnet())),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		rangeObjects = append(rangeObjects, rangeObj)
	}

	itemsList, diags := types.ListValue(
		types.ObjectType{AttrTypes: ItemAttrTypes},
		rangeObjects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	config.Items = itemsList
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func filterSet(value types.List) (map[string]struct{}, bool) {
	active := !value.IsNull() && value.Elements() != nil
	values := make(map[string]struct{})
	if active {
		for _, element := range value.Elements() {
			values[strings.Trim(element.String(), "\"")] = struct{}{}
		}
	}
	return values, active
}
