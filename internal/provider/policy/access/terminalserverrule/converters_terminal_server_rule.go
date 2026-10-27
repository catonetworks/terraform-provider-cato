package terminalserverrule

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
) *cato_models.TerminalServerAddRuleDataInput {
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
		knownValue{"rule.allowed_host_ip", model.AllowedHostIP, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.exclude_traffic", model.ExcludeTraffic, true},
		knownValue{"rule.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.TerminalServerAddRuleDataInput{}
	input.AllowedHostIP = idname.PrepareIDName[cato_models.GlobalIPRangeRefInput](ctx, model.AllowedHostIP, diags)
	input.Description = model.Description.ValueString()
	input.Enabled = model.Enabled.ValueBool()
	input.ExcludeTraffic = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.ExcludeTraffic, diags)
	input.Name = model.Name.ValueString()
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.TerminalServerUpdateRuleDataInput {
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
		knownValue{"rule.allowed_host_ip", model.AllowedHostIP, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.exclude_traffic", model.ExcludeTraffic, true},
		knownValue{"rule.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.TerminalServerUpdateRuleDataInput{}
	input.AllowedHostIP = idname.PrepareIDName[cato_models.GlobalIPRangeRefInput](ctx, model.AllowedHostIP, diags)
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	input.ExcludeTraffic = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.ExcludeTraffic, diags)
	vName := model.Name.ValueString()
	input.Name = &vName
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

type parseReadResponse = cato.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Rules_Rule

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.AllowedHostIP = idname.ParseIDName(ctx, value.AllowedHostIP, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExcludeTraffic = idname.ParseIDNameSet(ctx, value.ExcludeTraffic, diags)
	if model.ExcludeTraffic.IsNull() {
		model.ExcludeTraffic = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseReadSection(ctx, &value.Section, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadSectionResponse = cato.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Rules_Rule_Section

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

type parseAddResponse = cato.PolicyTerminalServerAddRule_Policy_TerminalServer_AddRule_Rule_Rule

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.AllowedHostIP = idname.ParseIDName(ctx, value.AllowedHostIP, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExcludeTraffic = idname.ParseIDNameSet(ctx, value.ExcludeTraffic, diags)
	if model.ExcludeTraffic.IsNull() {
		model.ExcludeTraffic = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseAddSection(ctx, &value.Section, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddSectionResponse = cato.PolicyTerminalServerAddRule_Policy_TerminalServer_AddRule_Rule_Rule_Section

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

type parseUpdateResponse = cato.PolicyTerminalServerUpdateRule_Policy_TerminalServer_UpdateRule_Rule_Rule

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.AllowedHostIP = idname.ParseIDName(ctx, value.AllowedHostIP, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExcludeTraffic = idname.ParseIDNameSet(ctx, value.ExcludeTraffic, diags)
	if model.ExcludeTraffic.IsNull() {
		model.ExcludeTraffic = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateSectionResponse = cato.PolicyTerminalServerUpdateRule_Policy_TerminalServer_UpdateRule_Rule_Rule_Section

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

type parseMoveResponse = cato.PolicyTerminalServerMoveRule_Policy_TerminalServer_MoveRule_Rule_Rule

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.AllowedHostIP = idname.ParseIDName(ctx, value.AllowedHostIP, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExcludeTraffic = idname.ParseIDNameSet(ctx, value.ExcludeTraffic, diags)
	if model.ExcludeTraffic.IsNull() {
		model.ExcludeTraffic = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveSectionResponse = cato.PolicyTerminalServerMoveRule_Policy_TerminalServer_MoveRule_Rule_Rule_Section

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
