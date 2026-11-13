package ipaddressrangecontainer

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

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func DataSource() datasource.DataSource {
	return &ipAddressRangeContainerDataSource{}
}

type ipAddressRangeContainerDataSource struct {
	client *catoclient.Service
}

func (d *ipAddressRangeContainerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range_container"
}

func (d *ipAddressRangeContainerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_ip_address_range_container` data source fetches information about account IPv4 " +
			"range containers with optional filters. Returns a list of containers matching the filter criteria. " +
			"Member values are not included; use the container's ID with the provider's import/download tooling to " +
			"inspect members.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of container IDs to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of container names to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of IP address range containers matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Container ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Container name",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Container description",
							Computed:    true,
						},
						"size": schema.Int64Attribute{
							Description: "Number of members reported by the API",
							Computed:    true,
						},
						"mode": schema.StringAttribute{
							Description: "INLINE or URL, derived from whether the container has synchronization settings",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ipAddressRangeContainerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

//nolint:gocyclo,funlen
func (d *ipAddressRangeContainerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Lookup
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	searchInput := cato_models.ContainerSearchInput{
		Types: []cato_models.ContainerType{cato_models.ContainerTypeIPRange},
	}

	result, err := d.client.Catov2.ContainerList(ctx, d.client.AccountId, searchInput)
	tflog.Debug(ctx, "DataSource.ContainerList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API ContainerList error",
			err.Error(),
		)
		return
	}

	if result.GetContainer() == nil || result.GetContainer().GetList() == nil {
		resp.Diagnostics.AddError(
			"Cato API ContainerList error",
			"container list response is nil",
		)
		return
	}

	// Build filter maps
	filterByID := !config.IDFilter.IsNull() && config.IDFilter.Elements() != nil
	idsMap := make(map[string]struct{})
	if filterByID {
		for _, value := range config.IDFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"")
			idsMap[valueStr] = struct{}{}
		}
	}

	filterByName := !config.NameFilter.IsNull() && config.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range config.NameFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"")
			namesMap[valueStr] = struct{}{}
		}
	}

	var containerObjects []attr.Value

	for _, item := range result.GetContainer().GetList().GetContainers() {
		// Apply filters
		var shouldInclude bool
		switch {
		case filterByID && filterByName:
			shouldInclude = utils.Contains(idsMap, item.GetID()) && utils.Contains(namesMap, item.GetName())
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

		description := types.StringValue("")
		if item.GetDescription() != nil {
			description = types.StringValue(*item.GetDescription())
		}

		mode := "INLINE"
		if item.GetSyncData() != nil {
			mode = "URL"
		}

		containerObj, diags := types.ObjectValue(
			ItemAttrTypes,
			map[string]attr.Value{
				"id":          types.StringValue(item.GetID()),
				"name":        types.StringValue(item.GetName()),
				"description": description,
				"size":        types.Int64Value(item.GetSize()),
				"mode":        types.StringValue(mode),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		containerObjects = append(containerObjects, containerObj)
	}

	itemsList, diags := types.ListValue(
		types.ObjectType{AttrTypes: ItemAttrTypes},
		containerObjects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	config.Items = itemsList
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
