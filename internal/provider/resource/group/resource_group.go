package group

import (
	"context"
	"errors"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

const groupNotFoundMessage = "group not found"

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *catoclient.Service
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Group
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert members from Terraform set to SDK input
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []groupmembers.Member
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME based on which field is provided
			var by cato_models.ObjectRefBy
			var input string

			switch {
			case !member.ID.IsNull() && !member.ID.IsUnknown() && member.ID.ValueString() != "":
				// Use ID if provided
				by = cato_models.ObjectRefByID
				input = member.ID.ValueString()
			case !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "":
				// Fall back to NAME if ID not provided
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			default:
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
	input := cato_models.CreateGroupInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Members:     members,
	}

	tflog.Debug(ctx, "Create.GroupsCreateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.Catov2.GroupsCreateGroup(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.GroupsCreateGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API GroupsCreateGroup error",
			err.Error(),
		)
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(result.Groups.CreateGroup.Group.ID)

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating group state",
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

//nolint:gocyclo
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Group
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Group
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert members from Terraform set to SDK input
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []groupmembers.Member
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME based on which field is provided
			var by cato_models.ObjectRefBy
			var input string

			switch {
			case !member.ID.IsNull() && !member.ID.IsUnknown() && member.ID.ValueString() != "":
				// Use ID if provided
				by = cato_models.ObjectRefByID
				input = member.ID.ValueString()
			case !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "":
				// Fall back to NAME if ID not provided
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			default:
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
	} else {
		members = []*cato_models.GroupMemberRefTypedInput{}
	}

	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: plan.ID.ValueString(),
	}

	// Only send name if it has changed from state
	var nameToSend *string
	if plan.Name.ValueString() != state.Name.ValueString() {
		nameToSend = plan.Name.ValueStringPointer()
	}

	input := cato_models.UpdateGroupInput{
		Group:       groupRef,
		Name:        nameToSend,
		Description: plan.Description.ValueStringPointer(),
		Members:     members,
	}

	tflog.Debug(ctx, "Update.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.Catov2.GroupsUpdateGroup(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Update.GroupsUpdateGroup.response", map[string]interface{}{
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
	hydratedState, hydrateErr := r.hydrateGroupState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating group state",
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

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Group
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API using current state as plan reference
	hydratedState, hydrateErr := r.hydrateGroupState(ctx, state.ID.ValueString(), state)
	if hydrateErr != nil {
		// Check if group not found
		if hydrateErr.Error() == groupNotFoundMessage {
			tflog.Warn(ctx, "group not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group state",
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

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Group
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the group
	groupRef := cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.ID.ValueString(),
	}

	result, err := r.client.Catov2.GroupsDeleteGroup(ctx, groupRef, r.client.AccountId)
	tflog.Debug(ctx, "Delete.GroupsDeleteGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		// Check if the error is because the group is already deleted
		resp.Diagnostics.AddError(
			"Cato API GroupsDeleteGroup error",
			fmt.Sprintf("Error deleting group: %s", err.Error()),
		)
		return
	}
}

// hydrateGroupState fetches the current state of a group from the API
// It takes a plan parameter to match config members with API members correctly
func (r *groupResource) hydrateGroupState(ctx context.Context, groupID string, plan Group) (Group, error) {
	var state Group
	state.ID = types.StringValue(groupID)

	// Read group metadata using GroupsList
	groupListInput := &cato_models.GroupListInput{
		Filter: []*cato_models.GroupListFilterInput{
			{
				ID: []*cato_models.IDFilterInput{
					{
						Eq: &groupID,
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

	groupListResult, err := r.client.Catov2.GroupsList(ctx, groupListInput, r.client.AccountId)
	tflog.Debug(ctx, "hydrateGroupState.GroupsList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(groupListResult),
	})
	if err != nil {
		return state, fmt.Errorf("GroupsList API error: %w", err)
	}

	// Check if the group was found
	if groupListResult.Groups.GroupList == nil || len(groupListResult.Groups.GroupList.Items) == 0 {
		return state, errors.New(groupNotFoundMessage)
	}

	// Update state with group details from GroupsList
	groupItem := groupListResult.Groups.GroupList.Items[0]
	state.Name = types.StringValue(groupItem.Name)
	if groupItem.Description != nil {
		state.Description = types.StringValue(*groupItem.Description)
	} else {
		state.Description = types.StringNull()
	}

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

	membersResult, err := r.client.Catov2.GroupsMembers(ctx, groupRef, groupMembersListInput, r.client.AccountId)
	tflog.Debug(ctx, "hydrateGroupState.GroupsMembers.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(membersResult),
	})
	if err != nil {
		return state, fmt.Errorf("GroupsMembers API error: %w", err)
	}

	// Convert members to Terraform set, matching with plan/config to preserve user's choice of ID vs name
	state.Members = groupmembers.ParseGroupMemberListWithConfig(ctx, membersResult.Groups.Group.Members.Items, plan.Members, "group.members")

	return state, nil
}
