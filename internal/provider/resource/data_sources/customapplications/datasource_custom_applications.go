package customapplications

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
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

const pageLimit = 1000

func DataSource() datasource.DataSource {
	return &customApplicationsDataSource{}
}

type customApplicationsDataSource struct {
	client *catoclient.Service
}

func (d *customApplicationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_applications"
}

func (d *customApplicationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_custom_applications` data source fetches information about custom applications " +
			"with optional filters. Returns a list of custom applications matching the filter criteria.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of custom application IDs to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of custom application names to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of custom applications matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Custom application ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Custom application name",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Custom application description",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *customApplicationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *customApplicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Lookup
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.fetchAllCustomApplications(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API CustomAppDataCustomApplicationList error",
			err.Error(),
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

	// Process custom applications and apply filters
	var appObjects []attr.Value

	for _, item := range items {
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

		appObj, diags := types.ObjectValue(
			ItemAttrTypes,
			map[string]attr.Value{
				"id":          types.StringValue(item.GetID()),
				"name":        types.StringValue(item.GetName()),
				"description": types.StringValue(cast.ToString(item.GetDescription())),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		appObjects = append(appObjects, appObj)
	}

	itemsList, diags := types.ListValue(
		types.ObjectType{AttrTypes: ItemAttrTypes},
		appObjects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	config.Items = itemsList
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// fetchAllCustomApplications pages through CustomAppDataCustomApplicationList to exhaustion.
func (d *customApplicationsDataSource) fetchAllCustomApplications(
	ctx context.Context,
) ([]*cato_go_sdk.CustomAppDataCustomApplicationList_CustomAppData_CustomApplicationList_Items, error) {
	var allItems []*cato_go_sdk.CustomAppDataCustomApplicationList_CustomAppData_CustomApplicationList_Items

	from := int64(0)
	for {
		listInput := cato_models.CustomApplicationListInput{
			Filter: []*cato_models.CustomApplicationFilterInput{},
			Paging: &cato_models.PagingInput{
				Limit: pageLimit,
				From:  from,
			},
		}

		result, err := d.client.Catov2.CustomAppDataCustomApplicationList(ctx, d.client.AccountId, listInput)
		tflog.Debug(ctx, "DataSource.CustomAppDataCustomApplicationList.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			return nil, err
		}

		appList := result.GetCustomAppData().GetCustomApplicationList()
		if appList == nil {
			break
		}

		pageItems := appList.GetItems()
		allItems = append(allItems, pageItems...)

		total := int64(0)
		if paging := appList.GetPaging(); paging != nil {
			total = paging.GetTotal()
		}

		from += int64(len(pageItems))
		if len(pageItems) == 0 || len(pageItems) < pageLimit || int64(len(allItems)) >= total {
			break
		}
	}

	return allItems, nil
}
