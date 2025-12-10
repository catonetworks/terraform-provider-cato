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
		Description: "The `cato_group` data source fetches information about groups with optional filters. Returns a list of groups matching the filter criteria, each with their members.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of group IDs to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of group names to filter by",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of groups matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Group ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Group name",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Group description",
							Computed:    true,
						},
						"members": schema.SetNestedAttribute{
							Description: "List of group members",
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
	var config GroupsLookup
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch all groups using GroupsList with high limit
	groupListInput := &cato_models.GroupListInput{
		Filter: []*cato_models.GroupListFilterInput{},
		Paging: &cato_models.PagingInput{
			Limit: 1000,
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

	if groupListResult.Groups.GroupList == nil {
		resp.Diagnostics.AddError(
			"Cato API GroupsList error",
			"GroupList is nil",
		)
		return
	}

	// Build filter maps
	filterById := !config.IdFilter.IsNull() && config.IdFilter.Elements() != nil
	idsMap := make(map[string]struct{})
	if filterById {
		for _, value := range config.IdFilter.Elements() {
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

	// Process groups and apply filters
	var groupObjects []attr.Value

	for _, groupItem := range groupListResult.Groups.GroupList.Items {
		// Apply filters
		var shouldInclude bool
		if filterById && filterByName {
			// Both filters - both must match
			idMatches := contains(idsMap, groupItem.ID)
			nameMatches := contains(namesMap, groupItem.Name)
			shouldInclude = idMatches && nameMatches
		} else if filterById {
			shouldInclude = contains(idsMap, groupItem.ID)
		} else if filterByName {
			shouldInclude = contains(namesMap, groupItem.Name)
		} else {
			// No filters - include all
			shouldInclude = true
		}

		if !shouldInclude {
			continue
		}

		// Fetch members for this group
		groupRef := cato_models.GroupRefInput{
			By:    cato_models.ObjectRefByID,
			Input: groupItem.ID,
		}

		groupMembersListInput := cato_models.GroupMembersListInput{
			Filter: []*cato_models.GroupMembersListFilterInput{},
			Paging: &cato_models.PagingInput{
				Limit: 1000,
				From:  0,
			},
			Sort: &cato_models.GroupMembersListSortInput{},
		}

		membersResult, err := d.client.catov2.GroupsMembers(ctx, groupRef, groupMembersListInput, d.client.AccountId)
		if err != nil {
			tflog.Warn(ctx, "Failed to fetch members for group", map[string]interface{}{
				"group_id": groupItem.ID,
				"error":    err.Error(),
			})
			// Continue with empty members
		}

		// Build members set
		var memberValues []attr.Value
		if membersResult != nil && membersResult.Groups.Group != nil {
			for _, member := range membersResult.Groups.Group.Members.Items {
				memberObj, diags := types.ObjectValue(
					GroupMemberAttrTypes,
					map[string]attr.Value{
						"id":   types.StringValue(member.ID),
						"name": types.StringValue(member.Name),
						"type": types.StringValue(string(member.TypeGroupMemberRefTyped)),
					},
				)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				memberValues = append(memberValues, memberObj)
			}
		}

		membersSet, diags := types.SetValue(GroupMemberObjectType, memberValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// Build description
		description := types.StringNull()
		if groupItem.Description != nil {
			description = types.StringValue(*groupItem.Description)
		}

		// Build group object
		groupObj, diags := types.ObjectValue(
			GroupItemAttrTypes,
			map[string]attr.Value{
				"id":          types.StringValue(groupItem.ID),
				"name":        types.StringValue(groupItem.Name),
				"description": description,
				"members":     membersSet,
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		groupObjects = append(groupObjects, groupObj)
	}

	// Build items list
	itemsList, diags := types.ListValue(
		types.ObjectType{AttrTypes: GroupItemAttrTypes},
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
