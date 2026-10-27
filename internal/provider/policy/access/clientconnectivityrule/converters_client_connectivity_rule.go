package clientconnectivityrule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.ClientConnectivityAddRuleDataInput {
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
		knownValue{"rule.action", model.Action, true},
		knownValue{"rule.confidence_level", model.ConfidenceLevel, true},
		knownValue{"rule.connection_origin", model.ConnectionOrigin, true},
		knownValue{"rule.country", model.Country, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.device", model.Device, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.platform", model.Platform, true},
		knownValue{"rule.source", model.Source, true},
		knownValue{"rule.source_range", model.SourceRange, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.ClientConnectivityAddRuleDataInput{}
	input.Action = cato_models.ClientConnectivityActionEnum(model.Action.ValueString())
	input.ConfidenceLevel = cato_models.ClientConnectivityConfidenceLevelEnum(model.ConfidenceLevel.ValueString())
	input.ConnectionOrigin = convert.PrepareStrings[cato_models.ClientConnectivityOriginEnum](ctx, model.ConnectionOrigin, diags)
	if input.ConnectionOrigin == nil {
		input.ConnectionOrigin = []cato_models.ClientConnectivityOriginEnum{}
	}
	input.Country = idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, model.Country, diags)
	input.Description = model.Description.ValueString()
	input.Device = idname.PrepareIDNameSet[cato_models.DeviceProfileRefInput](ctx, model.Device, diags)
	input.Enabled = model.Enabled.ValueBool()
	input.Name = model.Name.ValueString()
	input.Platform = convert.PrepareStrings[cato_models.OperatingSystem](ctx, model.Platform, diags)
	if input.Platform == nil {
		input.Platform = []cato_models.OperatingSystem{}
	}
	input.Source = buildCreateObjectSource(ctx, model.Source, diags)
	input.SourceRange = make([]*cato_models.ClientConnectivitySourceRangeInput, 0, len(model.SourceRange.Elements()))
	for _, item := range model.SourceRange.Elements() {
		input.SourceRange = append(input.SourceRange, buildCreateObjectSourceRange(ctx, item.(types.Object), diags))
	}
	return input
}

func buildCreateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.ClientConnectivitySourceInput {
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
		knownValue{"rule.source.users_group", model.UsersGroup, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.ClientConnectivitySourceInput{}
	input.User = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, model.User, diags)
	input.UsersGroup = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, model.UsersGroup, diags)
	return input
}

func buildCreateObjectSourceRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.ClientConnectivitySourceRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source_range must be known before applying.")
		return nil
	}
	var model RuleSourceRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source_range.global_ip_range", model.GlobalIPRange, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.ClientConnectivitySourceRangeInput{}
	input.GlobalIPRange = idname.PrepareIDName[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.ClientConnectivityUpdateRuleDataInput {
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
		knownValue{"rule.action", model.Action, true},
		knownValue{"rule.confidence_level", model.ConfidenceLevel, true},
		knownValue{"rule.connection_origin", model.ConnectionOrigin, true},
		knownValue{"rule.country", model.Country, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.device", model.Device, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.platform", model.Platform, true},
		knownValue{"rule.source", model.Source, true},
		knownValue{"rule.source_range", model.SourceRange, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.ClientConnectivityUpdateRuleDataInput{}
	vAction := cato_models.ClientConnectivityActionEnum(model.Action.ValueString())
	input.Action = &vAction
	vConfidenceLevel := cato_models.ClientConnectivityConfidenceLevelEnum(model.ConfidenceLevel.ValueString())
	input.ConfidenceLevel = &vConfidenceLevel
	input.ConnectionOrigin = convert.PrepareStrings[cato_models.ClientConnectivityOriginEnum](ctx, model.ConnectionOrigin, diags)
	if input.ConnectionOrigin == nil {
		input.ConnectionOrigin = []cato_models.ClientConnectivityOriginEnum{}
	}
	input.Country = idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, model.Country, diags)
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	input.Device = idname.PrepareIDNameSet[cato_models.DeviceProfileRefInput](ctx, model.Device, diags)
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	vName := model.Name.ValueString()
	input.Name = &vName
	input.Platform = convert.PrepareStrings[cato_models.OperatingSystem](ctx, model.Platform, diags)
	if input.Platform == nil {
		input.Platform = []cato_models.OperatingSystem{}
	}
	input.Source = buildUpdateObjectSource(ctx, model.Source, diags)
	input.SourceRange = make([]*cato_models.ClientConnectivitySourceRangeInput, 0, len(model.SourceRange.Elements()))
	for _, item := range model.SourceRange.Elements() {
		input.SourceRange = append(input.SourceRange, buildUpdateObjectSourceRange(ctx, item.(types.Object), diags))
	}
	return input
}

func buildUpdateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.ClientConnectivitySourceUpdateInput {
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
		knownValue{"rule.source.users_group", model.UsersGroup, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.ClientConnectivitySourceUpdateInput{}
	input.User = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, model.User, diags)
	input.UsersGroup = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, model.UsersGroup, diags)
	return input
}

func buildUpdateObjectSourceRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.ClientConnectivitySourceRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source_range must be known before applying.")
		return nil
	}
	var model RuleSourceRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source_range.global_ip_range", model.GlobalIPRange, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.ClientConnectivitySourceRangeInput{}
	input.GlobalIPRange = idname.PrepareIDName[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
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
	if !model.Ref.IsNull() {
		vRef := model.Ref.ValueString()
		input.Ref = &vRef
	}
	return input
}

type parseReadResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.ConfidenceLevel = types.StringValue(string(value.ConfidenceLevel))
	model.ConnectionOrigin = convert.StringSet(ctx, append([]cato_models.ClientConnectivityOriginEnum{}, value.ConnectionOrigin...), diags)
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.Device = idname.ParseIDNameSet(ctx, value.Device, diags)
	if model.Device.IsNull() {
		model.Device = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.Source = parseReadSource(ctx, &value.Source, diags)
	valuesSourceRange := make([]attr.Value, 0, len(value.SourceRange))
	for _, item := range value.SourceRange {
		valuesSourceRange = append(valuesSourceRange, parseReadSourceRange(ctx, item, diags))
	}
	var dSourceRange diag.Diagnostics
	model.SourceRange, dSourceRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceRangeAttrTypes}, valuesSourceRange)
	diags.Append(dSourceRange...)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadSectionResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule_Section

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

type parseReadSourceResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule_Source

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

type parseReadSourceRangeResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule_SourceRange

func parseReadSourceRange(ctx context.Context, value *parseReadSourceRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_range object.")
		return types.ObjectNull(ruleSourceRangeAttrTypes)
	}
	model := RuleSourceRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddResponse = cato.PolicyClientConnectivityAddRule_Policy_ClientConnectivity_AddRule_Rule_Rule

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.ConfidenceLevel = types.StringValue(string(value.ConfidenceLevel))
	model.ConnectionOrigin = convert.StringSet(ctx, append([]cato_models.ClientConnectivityOriginEnum{}, value.ConnectionOrigin...), diags)
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.Device = idname.ParseIDNameSet(ctx, value.Device, diags)
	if model.Device.IsNull() {
		model.Device = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.Source = parseAddSource(ctx, &value.Source, diags)
	valuesSourceRange := make([]attr.Value, 0, len(value.SourceRange))
	for _, item := range value.SourceRange {
		valuesSourceRange = append(valuesSourceRange, parseAddSourceRange(ctx, item, diags))
	}
	var dSourceRange diag.Diagnostics
	model.SourceRange, dSourceRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceRangeAttrTypes}, valuesSourceRange)
	diags.Append(dSourceRange...)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddSectionResponse = cato.PolicyClientConnectivityAddRule_Policy_ClientConnectivity_AddRule_Rule_Rule_Section

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

type parseAddSourceResponse = cato.PolicyClientConnectivityAddRule_Policy_ClientConnectivity_AddRule_Rule_Rule_Source

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

type parseAddSourceRangeResponse = cato.PolicyClientConnectivityAddRule_Policy_ClientConnectivity_AddRule_Rule_Rule_SourceRange

func parseAddSourceRange(ctx context.Context, value *parseAddSourceRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_range object.")
		return types.ObjectNull(ruleSourceRangeAttrTypes)
	}
	model := RuleSourceRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateResponse = cato.PolicyClientConnectivityUpdateRule_Policy_ClientConnectivity_UpdateRule_Rule_Rule

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.ConfidenceLevel = types.StringValue(string(value.ConfidenceLevel))
	model.ConnectionOrigin = convert.StringSet(ctx, append([]cato_models.ClientConnectivityOriginEnum{}, value.ConnectionOrigin...), diags)
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.Device = idname.ParseIDNameSet(ctx, value.Device, diags)
	if model.Device.IsNull() {
		model.Device = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.Source = parseUpdateSource(ctx, &value.Source, diags)
	valuesSourceRange := make([]attr.Value, 0, len(value.SourceRange))
	for _, item := range value.SourceRange {
		valuesSourceRange = append(valuesSourceRange, parseUpdateSourceRange(ctx, item, diags))
	}
	var dSourceRange diag.Diagnostics
	model.SourceRange, dSourceRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceRangeAttrTypes}, valuesSourceRange)
	diags.Append(dSourceRange...)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateSectionResponse = cato.PolicyClientConnectivityUpdateRule_Policy_ClientConnectivity_UpdateRule_Rule_Rule_Section

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

type parseUpdateSourceResponse = cato.PolicyClientConnectivityUpdateRule_Policy_ClientConnectivity_UpdateRule_Rule_Rule_Source

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

type parseUpdateSourceRangeResponse = cato.PolicyClientConnectivityUpdateRule_Policy_ClientConnectivity_UpdateRule_Rule_Rule_SourceRange

func parseUpdateSourceRange(ctx context.Context, value *parseUpdateSourceRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_range object.")
		return types.ObjectNull(ruleSourceRangeAttrTypes)
	}
	model := RuleSourceRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveResponse = cato.PolicyClientConnectivityMoveRule_Policy_ClientConnectivity_MoveRule_Rule_Rule

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.ConfidenceLevel = types.StringValue(string(value.ConfidenceLevel))
	model.ConnectionOrigin = convert.StringSet(ctx, append([]cato_models.ClientConnectivityOriginEnum{}, value.ConnectionOrigin...), diags)
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Description = types.StringValue(value.Description)
	model.Device = idname.ParseIDNameSet(ctx, value.Device, diags)
	if model.Device.IsNull() {
		model.Device = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.Source = parseMoveSource(ctx, &value.Source, diags)
	valuesSourceRange := make([]attr.Value, 0, len(value.SourceRange))
	for _, item := range value.SourceRange {
		valuesSourceRange = append(valuesSourceRange, parseMoveSourceRange(ctx, item, diags))
	}
	var dSourceRange diag.Diagnostics
	model.SourceRange, dSourceRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceRangeAttrTypes}, valuesSourceRange)
	diags.Append(dSourceRange...)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveSectionResponse = cato.PolicyClientConnectivityMoveRule_Policy_ClientConnectivity_MoveRule_Rule_Rule_Section

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

type parseMoveSourceResponse = cato.PolicyClientConnectivityMoveRule_Policy_ClientConnectivity_MoveRule_Rule_Rule_Source

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

type parseMoveSourceRangeResponse = cato.PolicyClientConnectivityMoveRule_Policy_ClientConnectivity_MoveRule_Rule_Rule_SourceRange

func parseMoveSourceRange(ctx context.Context, value *parseMoveSourceRangeResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_range object.")
		return types.ObjectNull(ruleSourceRangeAttrTypes)
	}
	model := RuleSourceRangeModel{}
	model.GlobalIPRange = idname.ParseIDName(ctx, value.GlobalIPRange, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceRangeAttrTypes, model)
	diags.Append(d...)
	return result
}
