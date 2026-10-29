package pasection

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyAddSectionInfoInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "section must be known before applying.")
		return nil
	}
	var model SectionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"section.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyAddSectionInfoInput{}
	input.Name = model.Name.ValueString()
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyUpdateSectionInfoInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "section must be known before applying.")
		return nil
	}
	var model SectionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"section.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyUpdateSectionInfoInput{}
	vName := model.Name.ValueString()
	input.Name = &vName
	return input
}

func buildPosition(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicySectionPositionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "at must be known before applying.")
		return nil
	}
	var model AtModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"at.position", model.Position, true},
		knownValue{"at.ref", model.Ref, false},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicySectionPositionInput{}
	input.Position = cato_models.PolicySectionPositionEnum(model.Position.ValueString())
	input.Ref = convert.StringPointerForOptionalInput(model.Ref)
	return input
}

type parseReadResponse = cato.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Sections_Section

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	model := SectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, sectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddResponse = cato.PolicyPrivateAccessAddSection_Policy_PrivateAccess_AddSection_Section_Section

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	model := SectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, sectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateResponse = cato.PolicyPrivateAccessUpdateSection_Policy_PrivateAccess_UpdateSection_Section_Section

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	model := SectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, sectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveResponse = cato.PolicyPrivateAccessMoveSection_Policy_PrivateAccess_MoveSection_Section_Section

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	model := SectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, sectionAttrTypes, model)
	diags.Append(d...)
	return result
}
