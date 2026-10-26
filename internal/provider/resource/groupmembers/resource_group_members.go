package groupmembers

import (
	"context"
	"fmt"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	groupmembers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *groupMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_members"
}

func (r *groupMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *groupMembersResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Members
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
	input := convertCreateInput(ctx, &plan, &diags, groupID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "groupMembersCreate.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.resourceAPI().GroupsUpdateGroup(ctx, input, r.client.AccountId)
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
	hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, plan.ID.ValueString(), plan.GroupName.ValueString(), plan.Members)
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
	var plan Members
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Members
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
	membersToAdd, membersToRemove := convertUpdateInput(ctx, &plan, &diags, state, groupID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only update if there are changes
	if len(membersToAdd) == 0 && len(membersToRemove) == 0 {
		// No changes, but still hydrate state to ensure all computed values are populated
		hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, plan.ID.ValueString(), plan.GroupName.ValueString(), plan.Members)
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
	input := convertUpdateInput2(plan, membersToAdd, membersToRemove)

	tflog.Debug(ctx, "groupMembersUpdate.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.resourceAPI().GroupsUpdateGroup(ctx, input, r.client.AccountId)
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
	hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, plan.ID.ValueString(), plan.GroupName.ValueString(), plan.Members)
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
	var state Members
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupMembersState(ctx, state.ID.ValueString(), state.GroupName.ValueString(), state.Members)
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
	var state Members
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all members from state to remove
	membersToRemove := convertDeleteInput(ctx, state, &diags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove all members from the group
	if len(membersToRemove) > 0 {
		input := convertDeleteValues(state, membersToRemove)

		tflog.Debug(ctx, "groupMembersDelete.GroupsUpdateGroup.request", map[string]interface{}{
			"request": utils.InterfaceToJSONString(input),
		})
		result, err := r.resourceAPI().GroupsUpdateGroup(ctx, input, r.client.AccountId)
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
//
//nolint:gocyclo
func (r *groupMembersResource) hydrateGroupMembersState(
	ctx context.Context,
	groupID string,
	groupName string,
	planMembers types.Set,
) (Members, error) {
	state, groupRef, groupMembersListInput := convertHydrateGroupMembersStateValues(groupID, groupName)

	membersResult, err := r.resourceAPI().GroupsMembers(ctx, groupRef, groupMembersListInput, r.client.AccountId)
	tflog.Debug(ctx, "hydrateGroupMembersState.GroupsMembers.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(membersResult),
	})
	if err != nil {
		return state, fmt.Errorf("GroupsMembers API error: %w", err)
	}

	if membersResult.Groups == nil ||
		membersResult.Groups.Group == nil ||
		membersResult.Groups.Group.Members.Items == nil ||
		len(membersResult.Groups.Group.Members.Items) == 0 {
		return state, fmt.Errorf("group not found")
	}

	// Get plan members to create a filter
	var planMembersList []groupmembers.Member
	if !planMembers.IsNull() && !planMembers.IsUnknown() {
		diags := convertHydrateGroupMembersStateValues2(ctx, planMembers, &planMembersList)
		if diags.HasError() {
			return state, fmt.Errorf("failed to parse plan members")
		}
	}

	// Create a map of plan members for quick lookup
	// Key format: "type:id" or "type:name" depending on what's available
	planMembersMap := make(map[string]bool)
	for _, member := range planMembersList {
		if !member.ID.IsNull() && !member.ID.IsUnknown() && member.ID.ValueString() != "" {
			key := fmt.Sprintf("%s:id:%s", member.Type.ValueString(), member.ID.ValueString())
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
		memberObj := groupmembers.ParseGroupMember(ctx, apiMember, "group_members.members")
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
	state.Members = groupmembers.ParseGroupMemberList(ctx, filteredMembers, "group_members.members")

	return state, nil
}
