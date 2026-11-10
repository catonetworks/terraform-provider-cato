package subscriptiongroup

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
	return &subscriptionGroupsDataSource{}
}

type subscriptionGroupsDataSource struct {
	client *catoclient.Service
}

func (d *subscriptionGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_groups"
}

func (d *subscriptionGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_subscription_groups` data source fetches information about account notification " +
			"subscription groups with optional filters. Returns a list of subscription groups matching the filter " +
			"criteria, each with its referenced mailing lists.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of subscription group IDs to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of subscription group names to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of subscription groups matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Subscription group ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Subscription group name",
							Computed:    true,
						},
						"mailing_list": schema.ListNestedAttribute{
							Description: "Mailing lists referenced by this subscription group",
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
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *subscriptionGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

//nolint:gocyclo,funlen
func (d *subscriptionGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Lookup
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.Catov2.NotificationSubscriptionGroupList(ctx, d.client.AccountId)
	tflog.Debug(ctx, "DataSource.NotificationSubscriptionGroupList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API NotificationSubscriptionGroupList error",
			err.Error(),
		)
		return
	}

	if result.GetNotification() == nil {
		resp.Diagnostics.AddError(
			"Cato API NotificationSubscriptionGroupList error",
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

	var groupObjects []attr.Value

	for _, item := range result.GetNotification().GetSubscriptionGroupList().GetItems() {
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

		// Build mailing list references
		var mailingListValues []attr.Value
		for _, ml := range item.GetMailingList() {
			mlObj, diags := types.ObjectValue(
				MailingListAttrTypes,
				map[string]attr.Value{
					"id":   types.StringValue(ml.GetID()),
					"name": types.StringValue(ml.GetName()),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			mailingListValues = append(mailingListValues, mlObj)
		}
		mailingListList, diags := types.ListValue(MailingListObjectType, mailingListValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		groupObj, diags := types.ObjectValue(
			ItemAttrTypes,
			map[string]attr.Value{
				"id":           types.StringValue(item.GetID()),
				"name":         types.StringValue(item.GetName()),
				"mailing_list": mailingListList,
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		groupObjects = append(groupObjects, groupObj)
	}

	itemsList, diags := types.ListValue(
		types.ObjectType{AttrTypes: ItemAttrTypes},
		groupObjects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	config.Items = itemsList
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
