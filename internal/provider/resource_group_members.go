package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &groupMembersResource{}
	_ resource.ResourceWithConfigure   = &groupMembersResource{}
	_ resource.ResourceWithImportState = &groupMembersResource{}
)

func NewGroupMembersResource() resource.Resource {
	return &groupMembersResource{}
}

type groupMembersResource struct {
	client *catoClientData
}

func (r *groupMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_members"
}

func (r *groupMembersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_group_members` resource contains the configuration parameters necessary to manage a group. Groups can contain various member types including sites, hosts, network ranges, and more. Documentation for the underlying API used in this resource can be found at [mutation.groups.createGroup()](https://api.catonetworks.com/documentation/#mutation-groups.createGroup).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Group ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_name": schema.StringAttribute{
				Description: "Group Name",
				Required:    true,
			},
			"members": schema.SetNestedAttribute{
				Description: "List of group members. Each member has 'name', 'id', and 'type' fields. You can specify either 'name' or 'id' when creating/updating; both will be populated in state. At least one member is required.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Member name (computed if not specified)",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"id": schema.StringAttribute{
							Description: "Member ID (computed if not specified)",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"type": schema.StringAttribute{
							Description: "Member type: FLOATING_SUBNET, GLOBAL_IP_RANGE, HOST, NETWORK_INTERFACE, SITE, SITE_NETWORK_SUBNET", // USER, USERS_GROUP, SYSTEM_GROUP
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"FLOATING_SUBNET",
									"GLOBAL_IP_RANGE",
									"HOST",
									"NETWORK_INTERFACE",
									"SITE",
									"SITE_NETWORK_SUBNET",
									// "USER",
									// "USERS_GROUP",
									// "SYSTEM_GROUP",
								),
							},
						},
					},
				},
			},
		},
	}
}

func (r *groupMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *groupMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan GroupMembers
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the group exists and get its ID
	groupID, err := r.getGroupByName(ctx, plan.GroupName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Group not found",
			fmt.Sprintf("Cannot find group '%s': %s", plan.GroupName.ValueString(), err.Error()),
		)
		return
	}

	// Set the group ID in state
	plan.Id = types.StringValue(groupID)

	// Convert members from Terraform set to SDK input
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []GroupMember
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME based on which field is provided
			var by cato_models.ObjectRefBy
			var input string

			if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
				// Use ID if provided
				by = cato_models.ObjectRefByID
				input = member.Id.ValueString()
			} else if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
				// Fall back to NAME if ID not provided
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return
			}

			members = append(members, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// setting input
	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: plan.Id.ValueString(),
	}

	input := cato_models.UpdateGroupInput{
		Group:        groupRef,
		MembersToAdd: members,
	}

	tflog.Debug(ctx, "groupMembersCreate.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.GroupsUpdateGroup(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "groupMembersCreate.GroupsUpdateGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API GroupsUpdateGroup error",
			err.Error(),
		)
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, plan.Id.ValueString(), plan.GroupName.ValueString(), plan.Members)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating group members state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan GroupMembers
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state GroupMembers
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the group exists and get its ID
	groupID, err := r.getGroupByName(ctx, plan.GroupName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Group not found",
			fmt.Sprintf("Cannot find group '%s': %s", plan.GroupName.ValueString(), err.Error()),
		)
		return
	}

	// Update the group ID in state
	plan.Id = types.StringValue(groupID)

	// Get current state members
	var stateMembersList []GroupMember
	if !state.Members.IsNull() && !state.Members.IsUnknown() {
		diags = state.Members.ElementsAs(ctx, &stateMembersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Get planned members
	var planMembersList []GroupMember
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		diags = plan.Members.ElementsAs(ctx, &planMembersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create maps to track members by unique key (type:id or type:name)
	stateMembersMap := make(map[string]GroupMember)
	for _, member := range stateMembersList {
		key := fmt.Sprintf("%s:%s:%s", member.Type.ValueString(), member.Id.ValueString(), member.Name.ValueString())
		stateMembersMap[key] = member
	}

	planMembersMap := make(map[string]GroupMember)
	for _, member := range planMembersList {
		key := fmt.Sprintf("%s:%s:%s", member.Type.ValueString(), member.Id.ValueString(), member.Name.ValueString())
		planMembersMap[key] = member
	}

	// Determine members to add (in plan but not in state)
	var membersToAdd []*cato_models.GroupMemberRefTypedInput
	for key, member := range planMembersMap {
		if _, exists := stateMembersMap[key]; !exists {
			// Determine whether to use ID or NAME
			var by cato_models.ObjectRefBy
			var input string

			if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
				by = cato_models.ObjectRefByID
				input = member.Id.ValueString()
			} else if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return
			}

			membersToAdd = append(membersToAdd, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// Determine members to remove (in state but not in plan)
	var membersToRemove []*cato_models.GroupMemberRefTypedInput
	for key, member := range stateMembersMap {
		if _, exists := planMembersMap[key]; !exists {
			// Determine whether to use ID or NAME
			var by cato_models.ObjectRefBy
			var input string

			if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
				by = cato_models.ObjectRefByID
				input = member.Id.ValueString()
			} else if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return
			}

			membersToRemove = append(membersToRemove, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// Only update if there are changes
	if len(membersToAdd) == 0 && len(membersToRemove) == 0 {
		// No changes, but still hydrate state to ensure all computed values are populated
		hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, plan.Id.ValueString(), plan.GroupName.ValueString(), plan.Members)
		if hydrateErr != nil {
			resp.Diagnostics.AddError(
				"Error hydrating group members state",
				hydrateErr.Error(),
			)
			return
		}
		diags = resp.State.Set(ctx, &hydratedState)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Setting input with MembersToAdd and MembersToRemove
	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: plan.Id.ValueString(),
	}

	input := cato_models.UpdateGroupInput{
		Group:           groupRef,
		MembersToAdd:    membersToAdd,
		MembersToRemove: membersToRemove,
	}

	tflog.Debug(ctx, "groupMembersUpdate.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.GroupsUpdateGroup(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "GroupsCreateGroup.GroupsUpdateGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API GroupsUpdateGroup error",
			err.Error(),
		)
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, plan.Id.ValueString(), plan.GroupName.ValueString(), plan.Members)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating group members state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state GroupMembers
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, state.Id.ValueString(), state.GroupName.ValueString(), state.Members)
	if hydrateErr != nil {
		// Check if group not found
		if hydrateErr.Error() == "group not found" {
			tflog.Warn(ctx, "group not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group members state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state GroupMembers
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all members from state to remove
	var membersToRemove []*cato_models.GroupMemberRefTypedInput
	if !state.Members.IsNull() && !state.Members.IsUnknown() {
		var membersList []GroupMember
		diags = state.Members.ElementsAs(ctx, &membersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME
			var by cato_models.ObjectRefBy
			var input string

			if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
				by = cato_models.ObjectRefByID
				input = member.Id.ValueString()
			} else if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return
			}

			membersToRemove = append(membersToRemove, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// Remove all members from the group
	if len(membersToRemove) > 0 {
		groupRef := &cato_models.GroupRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.Id.ValueString(),
		}

		input := cato_models.UpdateGroupInput{
			Group:           groupRef,
			MembersToRemove: membersToRemove,
		}

		tflog.Debug(ctx, "groupMembersDelete.GroupsUpdateGroup.request", map[string]interface{}{
			"request": utils.InterfaceToJSONString(input),
		})
		result, err := r.client.catov2.GroupsUpdateGroup(ctx, input, r.client.AccountId)
		tflog.Debug(ctx, "groupMembersDelete.GroupsUpdateGroup.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})

		if err != nil {
			resp.Diagnostics.AddError(
				"Cato API GroupsUpdateGroup error",
				fmt.Sprintf("Error removing group members: %s", err.Error()),
			)
			return
		}
	}
}

// hydrateGroupMembersState fetches the current state of group members from the API
// It only includes members that are in the planMembers set
func (r *groupMembersResource) hydrateGroupMembersState(ctx context.Context, groupID string, groupName string, planMembers types.Set) (GroupMembers, error) {
	var state GroupMembers
	state.Id = types.StringValue(groupID)
	state.GroupName = types.StringValue(groupName)

	// Read group members using GroupsMembers
	groupRef := cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: groupID,
	}

	groupMembersListInput := cato_models.GroupMembersListInput{
		Filter: []*cato_models.GroupMembersListFilterInput{},
		Paging: &cato_models.PagingInput{
			Limit: 1000,
			From:  0,
		},
		Sort: &cato_models.GroupMembersListSortInput{},
	}

	membersResult, err := r.client.catov2.GroupsMembers(ctx, groupRef, groupMembersListInput, r.client.AccountId)
	tflog.Debug(ctx, "hydrateGroupMembersState.GroupsMembers.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(membersResult),
	})
	if err != nil {
		return state, fmt.Errorf("GroupsMembers API error: %w", err)
	}

	if membersResult.Groups == nil || membersResult.Groups.Group == nil || membersResult.Groups.Group.Members.Items == nil || len(membersResult.Groups.Group.Members.Items) == 0 {
		return state, fmt.Errorf("group not found")
	}

	// Get plan members to create a filter
	var planMembersList []GroupMember
	if !planMembers.IsNull() && !planMembers.IsUnknown() {
		diags := planMembers.ElementsAs(ctx, &planMembersList, false)
		if diags.HasError() {
			return state, fmt.Errorf("failed to parse plan members")
		}
	}

	// Create a map of plan members for quick lookup
	// Key format: "type:id" or "type:name" depending on what's available
	planMembersMap := make(map[string]bool)
	for _, member := range planMembersList {
		if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
			key := fmt.Sprintf("%s:id:%s", member.Type.ValueString(), member.Id.ValueString())
			planMembersMap[key] = true
		}
		if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
			key := fmt.Sprintf("%s:name:%s", member.Type.ValueString(), member.Name.ValueString())
			planMembersMap[key] = true
		}
	}

	// Filter API members to only include those in the plan
	filteredMembers := []interface{}{}
	for _, apiMember := range membersResult.Groups.Group.Members.Items {
		// Check if this member matches any in the plan
		memberObj := parseGroupMember(ctx, apiMember, "group_members.members")
		if memberObj.IsNull() {
			continue
		}

		// Extract the type, id, and name from the parsed object
		attrs := memberObj.Attributes()
		memberType := attrs["type"].(types.String).ValueString()
		memberID := attrs["id"].(types.String)
		memberName := attrs["name"].(types.String)

		// Check if this member is in the plan
		matchFound := false
		if !memberID.IsNull() && !memberID.IsUnknown() && memberID.ValueString() != "" {
			key := fmt.Sprintf("%s:id:%s", memberType, memberID.ValueString())
			if planMembersMap[key] {
				matchFound = true
			}
		}
		if !matchFound && !memberName.IsNull() && !memberName.IsUnknown() && memberName.ValueString() != "" {
			key := fmt.Sprintf("%s:name:%s", memberType, memberName.ValueString())
			if planMembersMap[key] {
				matchFound = true
			}
		}

		if matchFound {
			filteredMembers = append(filteredMembers, apiMember)
		}
	}

	// Convert filtered members to Terraform set
	state.Members = parseGroupMemberList(ctx, filteredMembers, "group_members.members")

	return state, nil
}

// getGroupByName fetches a group ID by name
func (r *groupMembersResource) getGroupByName(ctx context.Context, groupName string) (string, error) {
	var groupID string

	// Read group metadata using GroupsList
	groupListInput := &cato_models.GroupListInput{
		Paging: &cato_models.PagingInput{
			Limit: 1000,
			From:  0,
		},
		Sort: &cato_models.GroupListSortInput{},
	}

	groupListResult, err := r.client.catov2.GroupsList(ctx, groupListInput, r.client.AccountId)
	tflog.Debug(ctx, "getGroupByName.GroupsList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(groupListResult),
	})
	if err != nil {
		return groupID, fmt.Errorf("GroupsList API error: %w", err)
	}

	// Check if the group was found
	if groupListResult.Groups.GroupList == nil || len(groupListResult.Groups.GroupList.Items) == 0 {
		return groupID, fmt.Errorf("no groups found in account")
	}

	groupPresent := false

	for _, groupItem := range groupListResult.Groups.GroupList.Items {
		if groupItem.Name == groupName {
			groupPresent = true
			groupID = groupItem.ID
			break
		}
	}

	if !groupPresent {
		return groupID, fmt.Errorf("group not found: %s", groupName)
	}

	return groupID, nil
}
