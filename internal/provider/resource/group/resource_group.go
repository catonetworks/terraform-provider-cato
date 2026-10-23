package group

import (
	"context"
	"errors"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	groupmembers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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

	apiClient ResourceClient
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

func (r *groupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
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
	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.GroupsCreateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.resourceAPI().GroupsCreateGroup(ctx, input, r.client.AccountId)
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
	hydratedState := convertCreateInput2(
		ctx,
		r,
		&plan,
		result,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

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
	input := convertUpdateInput(ctx, plan, &diags, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.resourceAPI().GroupsUpdateGroup(ctx, input, r.client.AccountId)
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
	groupRef := convertDeleteInput(state)

	result, err := r.resourceAPI().GroupsDeleteGroup(ctx, groupRef, r.client.AccountId)
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

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	r *groupResource,
	priorPlan *Group,
	result *cato_go_sdk.GroupsCreateGroup,
	diagnostics *diag.Diagnostics,
) (resultHydratedState Group) {
	plan := *priorPlan
	defer func() { *priorPlan = plan }()
	plan.ID = types.StringValue(result.Groups.CreateGroup.Group.ID)

	// Hydrate state from API
	resultHydratedState, hydrateErr := r.hydrateGroupState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return resultHydratedState
	}
	return resultHydratedState
}

// hydrateGroupState fetches the current state of a group from the API
// It takes a plan parameter to match config members with API members correctly
func (r *groupResource) hydrateGroupState(ctx context.Context, groupID string, plan Group) (Group, error) {
	state, groupListInput := convertHydrateGroupStateValues(&groupID)

	groupListResult, err := r.resourceAPI().GroupsList(ctx, groupListInput, r.client.AccountId)
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
	groupRef, groupMembersListInput := convertHydrateGroupStateValues2(groupID, &state, groupListResult)

	membersResult, err := r.resourceAPI().GroupsMembers(ctx, groupRef, groupMembersListInput, r.client.AccountId)
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
