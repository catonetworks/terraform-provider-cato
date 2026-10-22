package lansection

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	plan SocketLanSection,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.PolicyAddSectionInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultInput = cato_models.PolicyAddSectionInput{}

	// setting at
	if !plan.At.IsNull() {
		resultInput.At = &cato_models.PolicySectionPositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)

		resultInput.At.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		resultInput.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// setting section
	if !plan.Section.IsNull() {
		resultInput.Section = &cato_models.PolicyAddSectionInfoInput{}
		var sectionInput idname.Ref
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)

		resultInput.Section.Name = sectionInput.Name.ValueString()
	}
	return resultInput
}

// convertCreateInput3 converts between the Terraform model and SDK values.
func convertCreateInput3(plan *SocketLanSection, policyChange *cato_go_sdk.PolicySocketLanAddSection) {
	plan.ID = types.StringValue(policyChange.GetPolicy().GetSocketLan().GetAddSection().Section.GetSection().ID)
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(ctx context.Context, state SocketLanSection, priorDiags *diag.Diagnostics) (resultSection idname.Ref) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = state.Section.As(ctx, &resultSection, basetypes.ObjectAsOptions{})
	return resultSection
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *SocketLanSection,
	body *cato_go_sdk.PolicySocketLanPolicy,
	section idname.Ref,
	diagnostics *diag.Diagnostics,
) (resultSectionExist bool) {
	sectionList := body.GetPolicy().SocketLan.Policy.GetSections()
	resultSectionExist = false
	for _, sectionListItem := range sectionList {
		if sectionListItem.GetSection().ID != section.ID.ValueString() {
			continue
		}

		resultSectionExist = true
		state.ID = types.StringValue(sectionListItem.GetSection().ID)
		curSectionObj, diagstmp := types.ObjectValue(
			idname.RefAttrTypes,
			map[string]attr.Value{
				"id":   types.StringValue(sectionListItem.GetSection().ID),
				"name": types.StringValue(sectionListItem.GetSection().GetName()),
			},
		)
		diagnostics.Append(diagstmp...)
		if diagnostics.HasError() {
			return resultSectionExist
		}
		state.Section = curSectionObj
	}

	// Hard coding LAST_IN_POLICY position as the API does not return any value
	curAtObj := state.At
	if !utils.HasValue(state.At) {
		var diagstmp diag.Diagnostics
		curAtObj, diagstmp = types.ObjectValue(
			policyrule.PositionAttrTypes,
			map[string]attr.Value{
				"position": types.StringValue("LAST_IN_POLICY"),
				"ref":      types.StringNull(),
			},
		)
		diagnostics.Append(diagstmp...)
	}
	state.At = curAtObj
	return resultSectionExist
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan SocketLanSection,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (
	resultInputUpdateSection cato_models.PolicyUpdateSectionInput,
	resultInputMoveSection cato_models.PolicyMoveSectionInput,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultInputUpdateSection = cato_models.PolicyUpdateSectionInput{}
	resultInputMoveSection = cato_models.PolicyMoveSectionInput{}

	// setting section
	if !plan.Section.IsNull() {
		resultInputUpdateSection.Section = &cato_models.PolicyUpdateSectionInfoInput{}
		var sectionInput idname.Ref
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)

		resultInputUpdateSection.Section.Name = sectionInput.Name.ValueStringPointer()
		resultInputUpdateSection.ID = sectionInput.ID.ValueString()
	}

	// setting at
	if !plan.At.IsNull() {
		resultInputMoveSection.To = &cato_models.PolicySectionPositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)

		resultInputMoveSection.To.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		resultInputMoveSection.To.Ref = positionInput.Ref.ValueStringPointer()
		resultInputMoveSection.ID = resultInputUpdateSection.ID
	}
	return resultInputUpdateSection, resultInputMoveSection
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state SocketLanSection,
	priorDiags *diag.Diagnostics,
) (resultSection idname.Ref) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = state.Section.As(ctx, &resultSection, basetypes.ObjectAsOptions{})
	return resultSection
}

// convertDeleteInput2 converts between the Terraform model and SDK values.
func convertDeleteInput2(section idname.Ref) (resultRemoveSection cato_models.PolicyRemoveSectionInput) {
	resultRemoveSection = cato_models.PolicyRemoveSectionInput{
		ID: section.ID.ValueString(),
	}
	return resultRemoveSection
}
