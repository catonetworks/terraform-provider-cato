package groupmembers

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	groupmembers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	plan *Members,
	priorDiags *diag.Diagnostics,
	groupID string,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.UpdateGroupInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	plan.ID = types.StringValue(groupID)

	// Convert members from Terraform set to SDK input
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
	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: plan.ID.ValueString(),
	}

	resultInput = cato_models.UpdateGroupInput{
		Group:        groupRef,
		MembersToAdd: members,
	}
	return resultInput
}

// convertUpdateInput converts between the Terraform model and SDK values.
//
//nolint:gocyclo,funlen
func convertUpdateInput(
	ctx context.Context,
	plan *Members,
	priorDiags *diag.Diagnostics,
	state Members,
	groupID string,
	diagnostics *diag.Diagnostics,
) (
	resultMembersToAdd []*cato_models.GroupMemberRefTypedInput,
	resultMembersToRemove []*cato_models.GroupMemberRefTypedInput,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	plan.ID = types.StringValue(groupID)

	// Get current state members
	var stateMembersList []groupmembers.Member
	if !state.Members.IsNull() && !state.Members.IsUnknown() {
		diags = state.Members.ElementsAs(ctx, &stateMembersList, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultMembersToAdd, resultMembersToRemove
		}
	}

	// Get planned members
	var planMembersList []groupmembers.Member
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		diags = plan.Members.ElementsAs(ctx, &planMembersList, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultMembersToAdd, resultMembersToRemove
		}
	}

	// Create maps to track members by unique key (type:id or type:name)
	stateMembersMap := make(map[string]groupmembers.Member)
	for _, member := range stateMembersList {
		key := fmt.Sprintf("%s:%s:%s", member.Type.ValueString(), member.ID.ValueString(), member.Name.ValueString())
		stateMembersMap[key] = member
	}

	planMembersMap := make(map[string]groupmembers.Member)
	for _, member := range planMembersList {
		key := fmt.Sprintf("%s:%s:%s", member.Type.ValueString(), member.ID.ValueString(), member.Name.ValueString())
		planMembersMap[key] = member
	}

	// Determine members to add (in plan but not in state)

	for key, member := range planMembersMap {
		if _, exists := stateMembersMap[key]; !exists {
			// Determine whether to use ID or NAME
			var by cato_models.ObjectRefBy
			var input string

			switch {
			case !member.ID.IsNull() && !member.ID.IsUnknown() && member.ID.ValueString() != "":
				by = cato_models.ObjectRefByID
				input = member.ID.ValueString()
			case !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "":
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			default:
				diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return resultMembersToAdd, resultMembersToRemove
			}

			resultMembersToAdd = append(resultMembersToAdd, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// Determine members to remove (in state but not in plan)

	for key, member := range stateMembersMap {
		if _, exists := planMembersMap[key]; !exists {
			// Determine whether to use ID or NAME
			var by cato_models.ObjectRefBy
			var input string

			switch {
			case !member.ID.IsNull() && !member.ID.IsUnknown() && member.ID.ValueString() != "":
				by = cato_models.ObjectRefByID
				input = member.ID.ValueString()
			case !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "":
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			default:
				diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return resultMembersToAdd, resultMembersToRemove
			}

			resultMembersToRemove = append(resultMembersToRemove, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}
	return resultMembersToAdd, resultMembersToRemove
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	plan Members,
	membersToAdd []*cato_models.GroupMemberRefTypedInput,
	membersToRemove []*cato_models.GroupMemberRefTypedInput,
) (resultInput cato_models.UpdateGroupInput) {
	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: plan.ID.ValueString(),
	}

	resultInput = cato_models.UpdateGroupInput{
		Group:           groupRef,
		MembersToAdd:    membersToAdd,
		MembersToRemove: membersToRemove,
	}
	return resultInput
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state Members,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultMembersToRemove []*cato_models.GroupMemberRefTypedInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	if !state.Members.IsNull() && !state.Members.IsUnknown() {
		var membersList []groupmembers.Member
		diags = state.Members.ElementsAs(ctx, &membersList, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultMembersToRemove
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME
			var by cato_models.ObjectRefBy
			var input string

			switch {
			case !member.ID.IsNull() && !member.ID.IsUnknown() && member.ID.ValueString() != "":
				by = cato_models.ObjectRefByID
				input = member.ID.ValueString()
			case !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "":
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			default:
				diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return resultMembersToRemove
			}

			resultMembersToRemove = append(resultMembersToRemove, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}
	return resultMembersToRemove
}

// convertDeleteValues converts between the Terraform model and SDK values.
func convertDeleteValues(
	state Members,
	membersToRemove []*cato_models.GroupMemberRefTypedInput,
) (resultInput cato_models.UpdateGroupInput) {
	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.ID.ValueString(),
	}

	resultInput = cato_models.UpdateGroupInput{
		Group:           groupRef,
		MembersToRemove: membersToRemove,
	}
	return resultInput
}

// convertHydrateGroupMembersStateValues converts between the Terraform model and SDK values.
func convertHydrateGroupMembersStateValues(
	groupID string,
	groupName string,
) (
	resultState Members,
	resultGroupRef cato_models.GroupRefInput,
	resultGroupMembersListInput cato_models.GroupMembersListInput,
) {
	resultState.ID = types.StringValue(groupID)
	resultState.GroupName = types.StringValue(groupName)

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
	return resultState, resultGroupRef, resultGroupMembersListInput
}

// convertHydrateGroupMembersStateValues2 converts between the Terraform model and SDK values.
func convertHydrateGroupMembersStateValues2(
	ctx context.Context,
	planMembers types.Set,
	originalPlanMembersList *[]groupmembers.Member,
) (resultDiags diag.Diagnostics) {
	planMembersList := *originalPlanMembersList
	defer func() { *originalPlanMembersList = planMembersList }()

	resultDiags = planMembers.ElementsAs(ctx, &planMembersList, false)
	return resultDiags
}
