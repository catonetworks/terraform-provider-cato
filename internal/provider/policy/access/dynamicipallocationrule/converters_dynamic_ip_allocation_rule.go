package dynamicipallocationrule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationAddRuleDataInput {
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
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.country",
			Value:    model.Country,
			Required: true},
		policyshared.KnownValue{Name: "rule.description",
			Value:    model.Description,
			Required: true},
		policyshared.KnownValue{Name: "rule.device_posture_profile",
			Value:    model.DevicePostureProfile,
			Required: true},
		policyshared.KnownValue{Name: "rule.enabled",
			Value:    model.Enabled,
			Required: true},
		policyshared.KnownValue{Name: "rule.name",
			Value:    model.Name,
			Required: true},
		policyshared.KnownValue{Name: "rule.platform",
			Value:    model.Platform,
			Required: true},
		policyshared.KnownValue{Name: "rule.range",
			Value:    model.Range,
			Required: true},
		policyshared.KnownValue{Name: "rule.source",
			Value:    model.Source,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.DynamicIPAllocationAddRuleDataInput{}
	input.Country = idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, model.Country, diags)
	input.Description = model.Description.ValueString()
	input.DevicePostureProfile = idname.PrepareIDNameSet[cato_models.DeviceProfileRefInput](ctx, model.DevicePostureProfile, diags)
	input.Enabled = model.Enabled.ValueBool()
	input.Name = model.Name.ValueString()
	input.Platform = convert.PrepareStrings[cato_models.OperatingSystem](ctx, model.Platform, diags)
	if input.Platform == nil {
		input.Platform = []cato_models.OperatingSystem{}
	}
	input.Range = buildCreateObjectRange(ctx, model.Range, diags)
	input.Source = buildCreateObjectSource(ctx, model.Source, diags)
	return input
}

func buildCreateObjectRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.range must be known before applying.")
		return nil
	}
	var model RuleRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.range.global_ip_range",
			Value:    model.GlobalIPRange,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.DynamicIPAllocationRangeInput{}
	input.GlobalIPRange = idname.PrepareIDName[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	return input
}

func buildCreateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationSourceInput {
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
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.source.user",
			Value:    model.User,
			Required: true},
		policyshared.KnownValue{Name: "rule.source.users_group",
			Value:    model.UsersGroup,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.DynamicIPAllocationSourceInput{}
	input.User = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, model.User, diags)
	input.UsersGroup = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, model.UsersGroup, diags)
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationUpdateRuleDataInput {
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
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.country",
			Value:    model.Country,
			Required: true},
		policyshared.KnownValue{Name: "rule.description",
			Value:    model.Description,
			Required: true},
		policyshared.KnownValue{Name: "rule.device_posture_profile",
			Value:    model.DevicePostureProfile,
			Required: true},
		policyshared.KnownValue{Name: "rule.enabled",
			Value:    model.Enabled,
			Required: true},
		policyshared.KnownValue{Name: "rule.name",
			Value:    model.Name,
			Required: true},
		policyshared.KnownValue{Name: "rule.platform",
			Value:    model.Platform,
			Required: true},
		policyshared.KnownValue{Name: "rule.range",
			Value:    model.Range,
			Required: true},
		policyshared.KnownValue{Name: "rule.source",
			Value:    model.Source,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.DynamicIPAllocationUpdateRuleDataInput{}
	input.Country = idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, model.Country, diags)
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	input.DevicePostureProfile = idname.PrepareIDNameSet[cato_models.DeviceProfileRefInput](ctx, model.DevicePostureProfile, diags)
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	vName := model.Name.ValueString()
	input.Name = &vName
	input.Platform = convert.PrepareStrings[cato_models.OperatingSystem](ctx, model.Platform, diags)
	if input.Platform == nil {
		input.Platform = []cato_models.OperatingSystem{}
	}
	input.Range = buildUpdateObjectRange(ctx, model.Range, diags)
	input.Source = buildUpdateObjectSource(ctx, model.Source, diags)
	return input
}

func buildUpdateObjectRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationRangeUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.range must be known before applying.")
		return nil
	}
	var model RuleRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.range.global_ip_range",
			Value:    model.GlobalIPRange,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.DynamicIPAllocationRangeUpdateInput{}
	input.GlobalIPRange = idname.PrepareIDName[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	return input
}

func buildUpdateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationSourceUpdateInput {
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
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.source.user",
			Value:    model.User,
			Required: true},
		policyshared.KnownValue{Name: "rule.source.users_group",
			Value:    model.UsersGroup,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.DynamicIPAllocationSourceUpdateInput{}
	input.User = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, model.User, diags)
	input.UsersGroup = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, model.UsersGroup, diags)
	return input
}

type parseReadResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.DevicePostureProfile = idname.ParseIDNameSet(ctx, value.DevicePostureProfile, diags)
	if model.DevicePostureProfile.IsNull() {
		model.DevicePostureProfile = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Range = parseReadRange(ctx, &value.Range, diags)
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.Source = parseReadSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadRangeResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule_Range

func parseReadRange(ctx context.Context, value *parseReadRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.range object.")
		return types.ObjectNull(ruleRangeAttrTypes)
	}
	model := RuleRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadSectionResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule_Section

func parseReadSection(ctx context.Context, value *parseReadSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseReadSourceResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule_Source

func parseReadSource(ctx context.Context, value *parseReadSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.User = idname.ParseIDNameSet(ctx, value.User, diags)
	if model.User.IsNull() {
		model.User = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.UsersGroup = idname.ParseIDNameSet(ctx, value.UsersGroup, diags)
	if model.UsersGroup.IsNull() {
		model.UsersGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddResponse = cato.PolicyDynamicIpAllocationAddRule_Policy_DynamicIPAllocation_AddRule_Rule_Rule

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.DevicePostureProfile = idname.ParseIDNameSet(ctx, value.DevicePostureProfile, diags)
	if model.DevicePostureProfile.IsNull() {
		model.DevicePostureProfile = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Range = parseAddRange(ctx, &value.Range, diags)
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.Source = parseAddSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddRangeResponse = cato.PolicyDynamicIpAllocationAddRule_Policy_DynamicIPAllocation_AddRule_Rule_Rule_Range

func parseAddRange(ctx context.Context, value *parseAddRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.range object.")
		return types.ObjectNull(ruleRangeAttrTypes)
	}
	model := RuleRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddSectionResponse = cato.PolicyDynamicIpAllocationAddRule_Policy_DynamicIPAllocation_AddRule_Rule_Rule_Section

func parseAddSection(ctx context.Context, value *parseAddSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseAddSourceResponse = cato.PolicyDynamicIpAllocationAddRule_Policy_DynamicIPAllocation_AddRule_Rule_Rule_Source

func parseAddSource(ctx context.Context, value *parseAddSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.User = idname.ParseIDNameSet(ctx, value.User, diags)
	if model.User.IsNull() {
		model.User = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.UsersGroup = idname.ParseIDNameSet(ctx, value.UsersGroup, diags)
	if model.UsersGroup.IsNull() {
		model.UsersGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateResponse = cato.PolicyDynamicIpAllocationUpdateRule_Policy_DynamicIPAllocation_UpdateRule_Rule_Rule

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.DevicePostureProfile = idname.ParseIDNameSet(ctx, value.DevicePostureProfile, diags)
	if model.DevicePostureProfile.IsNull() {
		model.DevicePostureProfile = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Range = parseUpdateRange(ctx, &value.Range, diags)
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.Source = parseUpdateSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateRangeResponse = cato.PolicyDynamicIpAllocationUpdateRule_Policy_DynamicIPAllocation_UpdateRule_Rule_Rule_Range

func parseUpdateRange(ctx context.Context, value *parseUpdateRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.range object.")
		return types.ObjectNull(ruleRangeAttrTypes)
	}
	model := RuleRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateSectionResponse = cato.PolicyDynamicIpAllocationUpdateRule_Policy_DynamicIPAllocation_UpdateRule_Rule_Rule_Section

func parseUpdateSection(ctx context.Context, value *parseUpdateSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseUpdateSourceResponse = cato.PolicyDynamicIpAllocationUpdateRule_Policy_DynamicIPAllocation_UpdateRule_Rule_Rule_Source

func parseUpdateSource(ctx context.Context, value *parseUpdateSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.User = idname.ParseIDNameSet(ctx, value.User, diags)
	if model.User.IsNull() {
		model.User = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.UsersGroup = idname.ParseIDNameSet(ctx, value.UsersGroup, diags)
	if model.UsersGroup.IsNull() {
		model.UsersGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveResponse = cato.PolicyDynamicIpAllocationMoveRule_Policy_DynamicIPAllocation_MoveRule_Rule_Rule

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.DevicePostureProfile = idname.ParseIDNameSet(ctx, value.DevicePostureProfile, diags)
	if model.DevicePostureProfile.IsNull() {
		model.DevicePostureProfile = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Range = parseMoveRange(ctx, &value.Range, diags)
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.Source = parseMoveSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveRangeResponse = cato.PolicyDynamicIpAllocationMoveRule_Policy_DynamicIPAllocation_MoveRule_Rule_Rule_Range

func parseMoveRange(ctx context.Context, value *parseMoveRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.range object.")
		return types.ObjectNull(ruleRangeAttrTypes)
	}
	model := RuleRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveSectionResponse = cato.PolicyDynamicIpAllocationMoveRule_Policy_DynamicIPAllocation_MoveRule_Rule_Rule_Section

func parseMoveSection(ctx context.Context, value *parseMoveSectionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseMoveSourceResponse = cato.PolicyDynamicIpAllocationMoveRule_Policy_DynamicIPAllocation_MoveRule_Rule_Rule_Source

func parseMoveSource(ctx context.Context, value *parseMoveSourceResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.User = idname.ParseIDNameSet(ctx, value.User, diags)
	if model.User.IsNull() {
		model.User = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.UsersGroup = idname.ParseIDNameSet(ctx, value.UsersGroup, diags)
	if model.UsersGroup.IsNull() {
		model.UsersGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}
