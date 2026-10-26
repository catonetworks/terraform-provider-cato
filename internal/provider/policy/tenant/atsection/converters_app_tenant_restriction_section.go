package atsection

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
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	plan AppTenantRestrictionSection,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.PolicyAddSectionInput) {
	resultInput = cato_models.PolicyAddSectionInput{}
	if !plan.At.IsNull() {
		resultInput.At = &cato_models.PolicySectionPositionInput{}
		positionInput := policyrule.Position{}
		diagnostics.Append(plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)
		resultInput.At.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		resultInput.At.Ref = positionInput.Ref.ValueStringPointer()
	}
	if !plan.Section.IsNull() {
		resultInput.Section = &cato_models.PolicyAddSectionInfoInput{}
		sectionInput := idname.Ref{}
		diagnostics.Append(plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})...)
		resultInput.Section.Name = sectionInput.Name.ValueString()
	}
	return resultInput
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	plan *AppTenantRestrictionSection,
	add *cato_go_sdk.PolicyAppTenantRestrictionAddSection_Policy_AppTenantRestriction_AddSection,
) (resultSid string) {
	resultSid = add.GetSection().GetSection().GetID()
	plan.ID = types.StringValue(resultSid)
	return resultSid
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	state *AppTenantRestrictionSection,
	section idname.Ref,
	pol *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy,
	diagnostics *diag.Diagnostics,
) (resultFound bool) {
	resultFound = false
	for _, item := range pol.GetSections() {
		if item == nil || item.GetSection() == nil {
			continue
		}
		if item.GetSection().GetID() != section.ID.ValueString() {
			continue
		}
		resultFound = true
		state.ID = types.StringValue(item.GetSection().GetID())
		cur, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(item.GetSection().GetID()),
			"name": types.StringValue(item.GetSection().GetName()),
		})
		diagnostics.Append(d...)
		state.Section = cur
		break
	}

	// Preserve the at block from existing state; the API does not return positioning
	// info on read. Only default to LAST_IN_POLICY when state has nothing (e.g. import).
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		diagnostics.Append(d...)
		state.At = atObj
	}
	return resultFound
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan AppTenantRestrictionSection,
	diagnostics *diag.Diagnostics,
) (
	resultInputUpdate cato_models.PolicyUpdateSectionInput,
	resultInputMove cato_models.PolicyMoveSectionInput,
) {
	resultInputUpdate = cato_models.PolicyUpdateSectionInput{}
	resultInputMove = cato_models.PolicyMoveSectionInput{}
	if !plan.Section.IsNull() {
		sectionInput := idname.Ref{}
		diagnostics.Append(plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})...)
		resultInputUpdate.ID = sectionInput.ID.ValueString()
		resultInputUpdate.Section = &cato_models.PolicyUpdateSectionInfoInput{
			Name: sectionInput.Name.ValueStringPointer(),
		}
	}
	if !plan.At.IsNull() {
		positionInput := policyrule.Position{}
		diagnostics.Append(plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)
		resultInputMove.To = &cato_models.PolicySectionPositionInput{
			Position: cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueStringPointer(),
		}
		resultInputMove.ID = resultInputUpdate.ID
	}
	return resultInputUpdate, resultInputMove
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(section idname.Ref) (resultRemove cato_models.PolicyRemoveSectionInput) {
	resultRemove = cato_models.PolicyRemoveSectionInput{ID: section.ID.ValueString()}
	return resultRemove
}
