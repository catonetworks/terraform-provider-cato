package wfsection

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	plan WanFirewallSection,
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
		sectionInput := idname.Ref{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)
		resultInput.Section.Name = sectionInput.Name.ValueString()
	}
	return resultInput
}

// populateWanSectionID sets the created WAN section ID from the mutation response.
func populateWanSectionID(plan *WanFirewallSection, policyChange *cato_go_sdk.PolicyWanFirewallAddSection) {
	plan.ID = types.StringValue(policyChange.GetPolicy().GetWanFirewall().GetAddSection().Section.GetSection().ID)
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	ctx context.Context,
	state WanFirewallSection,
	priorDiags *diag.Diagnostics,
) (resultSection idname.Ref) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultSection = idname.Ref{}
	diags = state.Section.As(ctx, &resultSection, basetypes.ObjectAsOptions{})
	return resultSection
}

// convertReadState3 converts between the Terraform model and SDK values.
func convertReadState3(
	ctx context.Context,
	state *WanFirewallSection,
	body *cato_go_sdk.Policy,
	section idname.Ref,
	diagnostics *diag.Diagnostics,
) (resultSectionExist bool, resultDiagstmp diag.Diagnostics) {
	sectionList := body.GetPolicy().WanFirewall.Policy.GetSections()
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
			return resultSectionExist, resultDiagstmp
		}
		state.Section = curSectionObj
	}

	// The API does not return section position. Preserve configuration state to
	// avoid drift, defaulting imports without position state to LAST_IN_POLICY.
	curAtObj, resultDiagstmp := convert.DefaultSectionPosition(ctx, state.At)
	state.At = curAtObj
	return resultSectionExist, resultDiagstmp
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan WanFirewallSection,
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
		sectionInput := idname.Ref{}
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
	state WanFirewallSection,
	priorDiags *diag.Diagnostics,
) (resultSection idname.Ref) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultSection = idname.Ref{}
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
