package group

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	groupmembers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	plan Group,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.CreateGroupInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []groupmembers.Member
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultInput
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
				diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return resultInput
			}

			members = append(members, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// setting input
	resultInput = cato_models.CreateGroupInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Members:     members,
	}
	return resultInput
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan Group,
	priorDiags *diag.Diagnostics,
	state Group,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.UpdateGroupInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []groupmembers.Member
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultInput
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
				diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return resultInput
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

	resultInput = cato_models.UpdateGroupInput{
		Group:       groupRef,
		Name:        nameToSend,
		Description: plan.Description.ValueStringPointer(),
		Members:     members,
	}
	return resultInput
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state Group) (resultGroupRef cato_models.GroupRefInput) {
	resultGroupRef = cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.ID.ValueString(),
	}
	return resultGroupRef
}

// convertHydrateGroupStateValues converts between the Terraform model and SDK values.
func convertHydrateGroupStateValues(
	originalGroupID *string,
) (resultState Group, resultGroupListInput *cato_models.GroupListInput) {
	groupID := *originalGroupID
	defer func() { *originalGroupID = groupID }()

	resultState.ID = types.StringValue(groupID)

	// Read group metadata using GroupsList
	resultGroupListInput = &cato_models.GroupListInput{
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
	return resultState, resultGroupListInput
}

// convertHydrateGroupStateValues2 converts between the Terraform model and SDK values.
func convertHydrateGroupStateValues2(
	groupID string,
	state *Group,
	groupListResult *cato_go_sdk.GroupsList,
) (
	resultGroupRef cato_models.GroupRefInput,
	resultGroupMembersListInput cato_models.GroupMembersListInput,
) {
	groupItem := groupListResult.Groups.GroupList.Items[0]
	state.Name = types.StringValue(groupItem.Name)
	if groupItem.Description != nil {
		state.Description = types.StringValue(*groupItem.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Read group members using GroupsMembers
	resultGroupRef = cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: groupID,
	}

	resultGroupMembersListInput = cato_models.GroupMembersListInput{
		Filter: []*cato_models.GroupMembersListFilterInput{},
		Paging: &cato_models.PagingInput{
			Limit: 1000,
			From:  0,
		},
		Sort: &cato_models.GroupMembersListSortInput{},
	}
	return resultGroupRef, resultGroupMembersListInput
}
