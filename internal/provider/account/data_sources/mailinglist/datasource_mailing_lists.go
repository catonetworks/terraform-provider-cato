package mailinglist

import (
	"context"
	"strings"

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

func NewDataSource() datasource.DataSource {
	return &mailingListsDataSource{}
}

type mailingListsDataSource struct {
	client *catoclient.Service
}

func (d *mailingListsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailing_lists"
}

func (d *mailingListsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_mailing_lists` data source fetches information about account mailing lists with " +
			"optional filters. Returns a list of mailing lists matching the filter criteria.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of mailing list IDs to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of mailing list names to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of mailing lists matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Mailing list ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Mailing list name",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Mailing list address type, e.g. SPECIFIC_EMAILS",
							Computed:    true,
						},
						"address": schema.ListAttribute{
							ElementType: types.StringType,
							Description: "Email addresses belonging to the mailing list",
							Computed:    true,
						},
						"admin": schema.ListNestedAttribute{
							Description: "Administrators of the mailing list",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "Admin ID",
										Computed:    true,
									},
									"name": schema.StringAttribute{
										Description: "Admin name",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *mailingListsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

//nolint:gocyclo,funlen
func (d *mailingListsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Lookup
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.Catov2.NotificationMailingListList(ctx, d.client.AccountId)
	tflog.Debug(ctx, "DataSource.NotificationMailingListList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API NotificationMailingListList error",
			err.Error(),
		)
		return
	}

	if result.GetNotification() == nil {
		resp.Diagnostics.AddError(
			"Cato API NotificationMailingListList error",
			"notification response is nil",
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

	var listObjects []attr.Value

	for _, item := range result.GetNotification().GetMailingListList().GetItems() {
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

		// Build address list
		var addressValues []attr.Value
		for _, address := range item.GetAddress() {
			addressValues = append(addressValues, types.StringValue(address))
		}
		addressList, diags := types.ListValue(types.StringType, addressValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// Build admin list
		var adminValues []attr.Value
		for _, admin := range item.GetAdmin() {
			adminObj, diags := types.ObjectValue(
				AdminAttrTypes,
				map[string]attr.Value{
					"id":   types.StringValue(admin.GetID()),
					"name": types.StringValue(admin.GetName()),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			adminValues = append(adminValues, adminObj)
		}
		adminList, diags := types.ListValue(AdminObjectType, adminValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		addrType := types.StringNull()
		if item.GetType() != nil {
			addrType = types.StringValue(string(*item.GetType()))
		}

		listObj, diags := types.ObjectValue(
			ItemAttrTypes,
			map[string]attr.Value{
				"id":      types.StringValue(item.GetID()),
				"name":    types.StringValue(item.GetName()),
				"type":    addrType,
				"address": addressList,
				"admin":   adminList,
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		listObjects = append(listObjects, listObj)
	}

	itemsList, diags := types.ListValue(
		types.ObjectType{AttrTypes: ItemAttrTypes},
		listObjects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	config.Items = itemsList
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
