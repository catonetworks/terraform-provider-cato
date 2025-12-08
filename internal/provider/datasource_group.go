package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func GroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	client *catoClientData
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_group` data source fetches information about a specific group by ID or name. Groups can contain various member types including sites, hosts, network ranges, and more.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Group ID (computed when using name lookup)",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"name": schema.StringAttribute{
				Description: "Group name (use instead of ID for name-based lookup)",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Group description",
				Computed:    true,
			},
			"members": schema.SetNestedAttribute{
				Description: "List of group members with name, id, and type",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Member name",
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "Member ID",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Member type: SITE, HOST, NETWORK_INTERFACE, GLOBAL_IP_RANGE, FLOATING_SUBNET, SITE_NETWORK_SUBNET, USER, USERS_GROUP, SYSTEM_GROUP",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Group
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var groupID string

	// Determine if we're looking up by ID or name
	if !config.Id.IsNull() && !config.Id.IsUnknown() {
		groupID = config.Id.ValueString()
	} else if !config.Name.IsNull() && !config.Name.IsUnknown() {
		// Lookup group by name using GroupsList
		groupName := config.Name.ValueString()
		groupListInput := &cato_models.GroupListInput{
			Filter: []*cato_models.GroupListFilterInput{
				{
					Name: []*cato_models.AdvancedStringFilterInput{
						{
							Eq: &groupName,
						},
					},
				},
			},
			Paging: &cato_models.PagingInput{
				Limit: 1,
				From:  0,
			},
			Sort: &cato_models.GroupListSortInput{},
		}

		groupListResult, err := d.client.catov2.GroupsList(ctx, groupListInput, d.client.AccountId)
		tflog.Debug(ctx, "DataSource.GroupsList.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(groupListResult),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Cato API GroupsList error",
				err.Error(),
			)
			return
		}

		// Check if the group was found
		if groupListResult.Groups.GroupList == nil || len(groupListResult.Groups.GroupList.Items) == 0 {
			resp.Diagnostics.AddError(
				"Group not found",
				"No group found with name: "+groupName,
			)
			return
		}

		groupID = groupListResult.Groups.GroupList.Items[0].ID
	} else {
		resp.Diagnostics.AddError(
			"Missing required attribute",
			"Either 'id' or 'name' must be specified",
		)
		return
	}

	// Create a temporary resource to use its hydration function
	// For datasource, we pass an empty plan since we're just reading
	tempResource := &groupResource{client: d.client}
	emptyPlan := Group{Id: types.StringValue(groupID)}
	state, err := tempResource.hydrateGroupState(ctx, groupID, emptyPlan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching group data",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
