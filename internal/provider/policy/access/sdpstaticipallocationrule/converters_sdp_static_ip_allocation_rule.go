package sdpstaticipallocationrule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SdpStaticIPAllocationAddRuleDataInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule must be known before applying.")
		return nil
	}
	var model RuleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.ip", model.IP, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.source", model.Source, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SdpStaticIPAllocationAddRuleDataInput{}
	input.Description = model.Description.ValueString()
	input.Enabled = model.Enabled.ValueBool()
	input.IP = model.IP.ValueString()
	input.Name = model.Name.ValueString()
	input.Source = buildCreateObjectSource(ctx, model.Source, diags)
	return input
}

func buildCreateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SdpStaticIPAllocationSourceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source must be known before applying.")
		return nil
	}
	var model RuleSourceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source.user", model.User, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SdpStaticIPAllocationSourceInput{}
	input.User = idname.PrepareIDName[cato_models.UserRefInput](ctx, model.User, diags)
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SdpStaticIPAllocationUpdateRuleDataInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule must be known before applying.")
		return nil
	}
	var model RuleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.ip", model.IP, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.source", model.Source, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SdpStaticIPAllocationUpdateRuleDataInput{}
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	vIP := model.IP.ValueString()
	input.IP = &vIP
	vName := model.Name.ValueString()
	input.Name = &vName
	input.Source = buildUpdateObjectSource(ctx, model.Source, diags)
	return input
}

func buildUpdateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SdpStaticIPAllocationSourceUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source must be known before applying.")
		return nil
	}
	var model RuleSourceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source.user", model.User, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SdpStaticIPAllocationSourceUpdateInput{}
	input.User = idname.PrepareIDName[cato_models.UserRefInput](ctx, model.User, diags)
	return input
}

func buildPosition(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRulePositionInput {
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
	input := &cato_models.PolicyRulePositionInput{}
	vPosition := cato_models.PolicyRulePositionEnum(model.Position.ValueString())
	input.Position = &vPosition
	input.Ref = convert.StringPointerForOptionalInput(model.Ref)
	return input
}

type parseReadResponse = cato.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Rules_Rule

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.IP = types.StringValue(value.IP)
	model.Name = types.StringValue(value.Name)
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.Source = parseReadSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadSectionResponse = cato.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Rules_Rule_Section

func parseReadSection(ctx context.Context, value *parseReadSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	model := RuleSectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, ruleSectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadSourceResponse = cato.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Rules_Rule_Source

func parseReadSource(ctx context.Context, value *parseReadSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	if value.User == nil {
		diags.AddError("Invalid API response", "Missing reference rule.source.user")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model.User = idname.ParseIDName(ctx, *value.User, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddResponse = cato.PolicySdpStaticIpAllocationAddRule_Policy_SdpStaticIPAllocation_AddRule_Rule_Rule

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.IP = types.StringValue(value.IP)
	model.Name = types.StringValue(value.Name)
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.Source = parseAddSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddSectionResponse = cato.PolicySdpStaticIpAllocationAddRule_Policy_SdpStaticIPAllocation_AddRule_Rule_Rule_Section

func parseAddSection(ctx context.Context, value *parseAddSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	model := RuleSectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, ruleSectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddSourceResponse = cato.PolicySdpStaticIpAllocationAddRule_Policy_SdpStaticIPAllocation_AddRule_Rule_Rule_Source

func parseAddSource(ctx context.Context, value *parseAddSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	if value.User == nil {
		diags.AddError("Invalid API response", "Missing reference rule.source.user")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model.User = idname.ParseIDName(ctx, *value.User, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateResponse = cato.PolicySdpStaticIpAllocationUpdateRule_Policy_SdpStaticIPAllocation_UpdateRule_Rule_Rule

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.IP = types.StringValue(value.IP)
	model.Name = types.StringValue(value.Name)
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.Source = parseUpdateSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateSectionResponse = cato.PolicySdpStaticIpAllocationUpdateRule_Policy_SdpStaticIPAllocation_UpdateRule_Rule_Rule_Section

func parseUpdateSection(ctx context.Context, value *parseUpdateSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	model := RuleSectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, ruleSectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateSourceResponse = cato.PolicySdpStaticIpAllocationUpdateRule_Policy_SdpStaticIPAllocation_UpdateRule_Rule_Rule_Source

func parseUpdateSource(ctx context.Context, value *parseUpdateSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	if value.User == nil {
		diags.AddError("Invalid API response", "Missing reference rule.source.user")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model.User = idname.ParseIDName(ctx, *value.User, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveResponse = cato.PolicySdpStaticIpAllocationMoveRule_Policy_SdpStaticIPAllocation_MoveRule_Rule_Rule

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.IP = types.StringValue(value.IP)
	model.Name = types.StringValue(value.Name)
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.Source = parseMoveSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveSectionResponse = cato.PolicySdpStaticIpAllocationMoveRule_Policy_SdpStaticIPAllocation_MoveRule_Rule_Rule_Section

func parseMoveSection(ctx context.Context, value *parseMoveSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	model := RuleSectionModel{}
	model.ID = types.StringValue(value.ID)
	model.Name = types.StringValue(value.Name)
	result, d := types.ObjectValueFrom(ctx, ruleSectionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveSourceResponse = cato.PolicySdpStaticIpAllocationMoveRule_Policy_SdpStaticIPAllocation_MoveRule_Rule_Rule_Source

func parseMoveSource(ctx context.Context, value *parseMoveSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	if value.User == nil {
		diags.AddError("Invalid API response", "Missing reference rule.source.user")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model.User = idname.ParseIDName(ctx, *value.User, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}
