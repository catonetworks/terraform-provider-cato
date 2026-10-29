package splittunnelrule

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
) *cato_models.SplitTunnelAddRuleDataInput {
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
		knownValue{"rule.country", model.Country, true},
		knownValue{"rule.coverage", model.Coverage, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.destination", model.Destination, true},
		knownValue{"rule.dns_exclusion", model.DNSExclusion, false},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.platform", model.Platform, true},
		knownValue{"rule.routing_priority", model.RoutingPriority, true},
		knownValue{"rule.source", model.Source, true},
		knownValue{"rule.source_network", model.SourceNetwork, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelAddRuleDataInput{}
	input.Action = cato_models.SplitTunnelActionEnum(model.Action.ValueString())
	input.Country = idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, model.Country, diags)
	vCoverage := cato_models.SplitTunnelCoverageEnum(model.Coverage.ValueString())
	input.Coverage = &vCoverage
	input.Description = model.Description.ValueString()
	input.Destination = buildCreateObjectDestination(ctx, model.Destination, diags)
	input.DNSExclusion = buildCreateObjectDNSExclusion(ctx, model.DNSExclusion, diags)
	input.Enabled = model.Enabled.ValueBool()
	input.Name = model.Name.ValueString()
	input.Platform = convert.PrepareStrings[cato_models.OperatingSystem](ctx, model.Platform, diags)
	if input.Platform == nil {
		input.Platform = []cato_models.OperatingSystem{}
	}
	input.RoutingPriority = cato_models.SplitTunnelRoutingPriorityEnum(model.RoutingPriority.ValueString())
	input.Source = buildCreateObjectSource(ctx, model.Source, diags)
	input.SourceNetwork = buildCreateObjectSourceNetwork(ctx, model.SourceNetwork, diags)
	return input
}

func buildCreateObjectDestination(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelDestinationInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.destination must be known before applying.")
		return nil
	}
	var model RuleDestinationModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.destination.application", model.Application, true},
		knownValue{"rule.destination.domain", model.Domain, true},
		knownValue{"rule.destination.fqdn", model.Fqdn, true},
		knownValue{"rule.destination.global_ip_range", model.GlobalIPRange, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelDestinationInput{}
	input.Application = idname.PrepareIDNameSet[cato_models.ApplicationRefInput](ctx, model.Application, diags)
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	input.Fqdn = convert.PrepareStrings[string](ctx, model.Fqdn, diags)
	if input.Fqdn == nil {
		input.Fqdn = []string{}
	}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	return input
}

func buildCreateObjectDNSExclusion(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelDNSExclusionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.dns_exclusion must be known before applying.")
		return nil
	}
	var model RuleDNSExclusionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.dns_exclusion.domain", model.Domain, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelDNSExclusionInput{}
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	return input
}

func buildCreateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelSourceInput {
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
	input := &cato_models.SplitTunnelSourceInput{}
	input.User = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, model.User, diags)
	input.UsersGroup = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, model.UsersGroup, diags)
	return input
}

func buildCreateObjectSourceNetwork(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelSourceNetworkInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source_network must be known before applying.")
		return nil
	}
	var model RuleSourceNetworkModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source_network.source_network_type", model.SourceNetworkType, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelSourceNetworkInput{}
	input.SourceNetworkType = cato_models.SplitTunnelSourceNetworkTypeEnum(model.SourceNetworkType.ValueString())
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelUpdateRuleDataInput {
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
		knownValue{"rule.country", model.Country, true},
		knownValue{"rule.coverage", model.Coverage, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.destination", model.Destination, true},
		knownValue{"rule.dns_exclusion", model.DNSExclusion, false},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.platform", model.Platform, true},
		knownValue{"rule.routing_priority", model.RoutingPriority, true},
		knownValue{"rule.source", model.Source, true},
		knownValue{"rule.source_network", model.SourceNetwork, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelUpdateRuleDataInput{}
	vAction := cato_models.SplitTunnelActionEnum(model.Action.ValueString())
	input.Action = &vAction
	input.Country = idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, model.Country, diags)
	vCoverage := cato_models.SplitTunnelCoverageEnum(model.Coverage.ValueString())
	input.Coverage = &vCoverage
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	input.Destination = buildUpdateObjectDestination(ctx, model.Destination, diags)
	input.DNSExclusion = buildUpdateObjectDNSExclusion(ctx, model.DNSExclusion, diags)
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	vName := model.Name.ValueString()
	input.Name = &vName
	input.Platform = convert.PrepareStrings[cato_models.OperatingSystem](ctx, model.Platform, diags)
	if input.Platform == nil {
		input.Platform = []cato_models.OperatingSystem{}
	}
	vRoutingPriority := cato_models.SplitTunnelRoutingPriorityEnum(model.RoutingPriority.ValueString())
	input.RoutingPriority = &vRoutingPriority
	input.Source = buildUpdateObjectSource(ctx, model.Source, diags)
	input.SourceNetwork = buildUpdateObjectSourceNetwork(ctx, model.SourceNetwork, diags)
	if model.DNSExclusion.IsNull() {
		input.DNSExclusion = &cato_models.SplitTunnelDNSExclusionUpdateInput{Domain: []string{}}
	}
	return input
}

func buildUpdateObjectDestination(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelDestinationUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.destination must be known before applying.")
		return nil
	}
	var model RuleDestinationModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.destination.application", model.Application, true},
		knownValue{"rule.destination.domain", model.Domain, true},
		knownValue{"rule.destination.fqdn", model.Fqdn, true},
		knownValue{"rule.destination.global_ip_range", model.GlobalIPRange, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelDestinationUpdateInput{}
	input.Application = idname.PrepareIDNameSet[cato_models.ApplicationRefInput](ctx, model.Application, diags)
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	input.Fqdn = convert.PrepareStrings[string](ctx, model.Fqdn, diags)
	if input.Fqdn == nil {
		input.Fqdn = []string{}
	}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	return input
}

func buildUpdateObjectDNSExclusion(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelDNSExclusionUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.dns_exclusion must be known before applying.")
		return nil
	}
	var model RuleDNSExclusionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.dns_exclusion.domain", model.Domain, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelDNSExclusionUpdateInput{}
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	return input
}

func buildUpdateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelSourceUpdateInput {
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
	input := &cato_models.SplitTunnelSourceUpdateInput{}
	input.User = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, model.User, diags)
	input.UsersGroup = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, model.UsersGroup, diags)
	return input
}

func buildUpdateObjectSourceNetwork(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SplitTunnelSourceNetworkUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source_network must be known before applying.")
		return nil
	}
	var model RuleSourceNetworkModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source_network.source_network_type", model.SourceNetworkType, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SplitTunnelSourceNetworkUpdateInput{}
	vSourceNetworkType := cato_models.SplitTunnelSourceNetworkTypeEnum(model.SourceNetworkType.ValueString())
	input.SourceNetworkType = &vSourceNetworkType
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

type parseReadResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	if value.Coverage == nil {
		diags.AddError("Invalid API response", "Missing rule.coverage")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.Coverage = types.StringValue(string(*value.Coverage))
	model.Description = types.StringValue(value.Description)
	model.Destination = parseReadDestination(ctx, &value.Destination, diags)
	if value.DNSExclusion == nil {
		model.DNSExclusion = types.ObjectNull(ruleDNSExclusionAttrTypes)
	} else {
		model.DNSExclusion = parseReadDNSExclusion(ctx, value.DNSExclusion, diags)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.RoutingPriority = types.StringValue(string(value.RoutingPriority))
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.Source = parseReadSource(ctx, &value.Source, diags)
	model.SourceNetwork = parseReadSourceNetwork(ctx, &value.SourceNetwork, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadDestinationResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule_Destination

func parseReadDestination(ctx context.Context, value *parseReadDestinationResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadDNSExclusionResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule_DNSExclusion

func parseReadDNSExclusion(ctx context.Context, value *parseReadDNSExclusionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.dns_exclusion object.")
		return types.ObjectNull(ruleDNSExclusionAttrTypes)
	}
	model := RuleDNSExclusionModel{}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	result, d := types.ObjectValueFrom(ctx, ruleDNSExclusionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseReadSectionResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule_Section

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

type parseReadSourceResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule_Source

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

type parseReadSourceNetworkResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule_SourceNetwork

func parseReadSourceNetwork(ctx context.Context, value *parseReadSourceNetworkResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_network object.")
		return types.ObjectNull(ruleSourceNetworkAttrTypes)
	}
	model := RuleSourceNetworkModel{}
	model.SourceNetworkType = types.StringValue(string(value.SourceNetworkType))
	result, d := types.ObjectValueFrom(ctx, ruleSourceNetworkAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddResponse = cato.PolicySplitTunnelAddRule_Policy_SplitTunnel_AddRule_Rule_Rule

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	if value.Coverage == nil {
		diags.AddError("Invalid API response", "Missing rule.coverage")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.Coverage = types.StringValue(string(*value.Coverage))
	model.Description = types.StringValue(value.Description)
	model.Destination = parseAddDestination(ctx, &value.Destination, diags)
	if value.DNSExclusion == nil {
		model.DNSExclusion = types.ObjectNull(ruleDNSExclusionAttrTypes)
	} else {
		model.DNSExclusion = parseAddDNSExclusion(ctx, value.DNSExclusion, diags)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.RoutingPriority = types.StringValue(string(value.RoutingPriority))
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.Source = parseAddSource(ctx, &value.Source, diags)
	model.SourceNetwork = parseAddSourceNetwork(ctx, &value.SourceNetwork, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddDestinationResponse = cato.PolicySplitTunnelAddRule_Policy_SplitTunnel_AddRule_Rule_Rule_Destination

func parseAddDestination(ctx context.Context, value *parseAddDestinationResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddDNSExclusionResponse = cato.PolicySplitTunnelAddRule_Policy_SplitTunnel_AddRule_Rule_Rule_DNSExclusion

func parseAddDNSExclusion(ctx context.Context, value *parseAddDNSExclusionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.dns_exclusion object.")
		return types.ObjectNull(ruleDNSExclusionAttrTypes)
	}
	model := RuleDNSExclusionModel{}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	result, d := types.ObjectValueFrom(ctx, ruleDNSExclusionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseAddSectionResponse = cato.PolicySplitTunnelAddRule_Policy_SplitTunnel_AddRule_Rule_Rule_Section

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

type parseAddSourceResponse = cato.PolicySplitTunnelAddRule_Policy_SplitTunnel_AddRule_Rule_Rule_Source

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

type parseAddSourceNetworkResponse = cato.PolicySplitTunnelAddRule_Policy_SplitTunnel_AddRule_Rule_Rule_SourceNetwork

func parseAddSourceNetwork(ctx context.Context, value *parseAddSourceNetworkResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_network object.")
		return types.ObjectNull(ruleSourceNetworkAttrTypes)
	}
	model := RuleSourceNetworkModel{}
	model.SourceNetworkType = types.StringValue(string(value.SourceNetworkType))
	result, d := types.ObjectValueFrom(ctx, ruleSourceNetworkAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateResponse = cato.PolicySplitTunnelUpdateRule_Policy_SplitTunnel_UpdateRule_Rule_Rule

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	if value.Coverage == nil {
		diags.AddError("Invalid API response", "Missing rule.coverage")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.Coverage = types.StringValue(string(*value.Coverage))
	model.Description = types.StringValue(value.Description)
	model.Destination = parseUpdateDestination(ctx, &value.Destination, diags)
	if value.DNSExclusion == nil {
		model.DNSExclusion = types.ObjectNull(ruleDNSExclusionAttrTypes)
	} else {
		model.DNSExclusion = parseUpdateDNSExclusion(ctx, value.DNSExclusion, diags)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.RoutingPriority = types.StringValue(string(value.RoutingPriority))
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.Source = parseUpdateSource(ctx, &value.Source, diags)
	model.SourceNetwork = parseUpdateSourceNetwork(ctx, &value.SourceNetwork, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateDestinationResponse = cato.PolicySplitTunnelUpdateRule_Policy_SplitTunnel_UpdateRule_Rule_Rule_Destination

func parseUpdateDestination(ctx context.Context, value *parseUpdateDestinationResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateDNSExclusionResponse = cato.PolicySplitTunnelUpdateRule_Policy_SplitTunnel_UpdateRule_Rule_Rule_DNSExclusion

func parseUpdateDNSExclusion(ctx context.Context, value *parseUpdateDNSExclusionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.dns_exclusion object.")
		return types.ObjectNull(ruleDNSExclusionAttrTypes)
	}
	model := RuleDNSExclusionModel{}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	result, d := types.ObjectValueFrom(ctx, ruleDNSExclusionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseUpdateSectionResponse = cato.PolicySplitTunnelUpdateRule_Policy_SplitTunnel_UpdateRule_Rule_Rule_Section

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

type parseUpdateSourceResponse = cato.PolicySplitTunnelUpdateRule_Policy_SplitTunnel_UpdateRule_Rule_Rule_Source

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

type parseUpdateSourceNetworkResponse = cato.PolicySplitTunnelUpdateRule_Policy_SplitTunnel_UpdateRule_Rule_Rule_SourceNetwork

func parseUpdateSourceNetwork(ctx context.Context, value *parseUpdateSourceNetworkResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_network object.")
		return types.ObjectNull(ruleSourceNetworkAttrTypes)
	}
	model := RuleSourceNetworkModel{}
	model.SourceNetworkType = types.StringValue(string(value.SourceNetworkType))
	result, d := types.ObjectValueFrom(ctx, ruleSourceNetworkAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveResponse = cato.PolicySplitTunnelMoveRule_Policy_SplitTunnel_MoveRule_Rule_Rule

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = types.StringValue(string(value.Action))
	model.Country = idname.ParseIDNameSet(ctx, value.Country, diags)
	if model.Country.IsNull() {
		model.Country = types.SetValueMust(idname.RefObjectType, nil)
	}
	if value.Coverage == nil {
		diags.AddError("Invalid API response", "Missing rule.coverage")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.Coverage = types.StringValue(string(*value.Coverage))
	model.Description = types.StringValue(value.Description)
	model.Destination = parseMoveDestination(ctx, &value.Destination, diags)
	if value.DNSExclusion == nil {
		model.DNSExclusion = types.ObjectNull(ruleDNSExclusionAttrTypes)
	} else {
		model.DNSExclusion = parseMoveDNSExclusion(ctx, value.DNSExclusion, diags)
	}
	model.Enabled = types.BoolValue(value.Enabled)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Platform = convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, value.Platform...), diags)
	model.RoutingPriority = types.StringValue(string(value.RoutingPriority))
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.Source = parseMoveSource(ctx, &value.Source, diags)
	model.SourceNetwork = parseMoveSourceNetwork(ctx, &value.SourceNetwork, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveDestinationResponse = cato.PolicySplitTunnelMoveRule_Policy_SplitTunnel_MoveRule_Rule_Rule_Destination

func parseMoveDestination(ctx context.Context, value *parseMoveDestinationResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveDNSExclusionResponse = cato.PolicySplitTunnelMoveRule_Policy_SplitTunnel_MoveRule_Rule_Rule_DNSExclusion

func parseMoveDNSExclusion(ctx context.Context, value *parseMoveDNSExclusionResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.dns_exclusion object.")
		return types.ObjectNull(ruleDNSExclusionAttrTypes)
	}
	model := RuleDNSExclusionModel{}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	result, d := types.ObjectValueFrom(ctx, ruleDNSExclusionAttrTypes, model)
	diags.Append(d...)
	return result
}

type parseMoveSectionResponse = cato.PolicySplitTunnelMoveRule_Policy_SplitTunnel_MoveRule_Rule_Rule_Section

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

type parseMoveSourceResponse = cato.PolicySplitTunnelMoveRule_Policy_SplitTunnel_MoveRule_Rule_Rule_Source

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

type parseMoveSourceNetworkResponse = cato.PolicySplitTunnelMoveRule_Policy_SplitTunnel_MoveRule_Rule_Rule_SourceNetwork

func parseMoveSourceNetwork(ctx context.Context, value *parseMoveSourceNetworkResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source_network object.")
		return types.ObjectNull(ruleSourceNetworkAttrTypes)
	}
	model := RuleSourceNetworkModel{}
	model.SourceNetworkType = types.StringValue(string(value.SourceNetworkType))
	result, d := types.ObjectValueFrom(ctx, ruleSourceNetworkAttrTypes, model)
	diags.Append(d...)
	return result
}

// preserveOptionalObjects normalizes an empty API default only when configuration
// omitted DNS exclusions. Nonempty remote values must remain visible for drift.
func preserveOptionalObjects(reference, value types.Object) types.Object {
	if reference.IsNull() || reference.IsUnknown() || value.IsNull() || value.IsUnknown() {
		return value
	}
	if !reference.Attributes()["dns_exclusion"].IsNull() {
		return value
	}
	fields := value.Attributes()
	exclusion := fields["dns_exclusion"].(types.Object)
	if exclusion.IsNull() || exclusion.IsUnknown() {
		return value
	}
	domains := exclusion.Attributes()["domain"].(types.Set)
	if domains.IsNull() || domains.IsUnknown() || len(domains.Elements()) != 0 {
		return value
	}
	fields["dns_exclusion"] = types.ObjectNull(ruleDNSExclusionAttrTypes)
	return types.ObjectValueMust(ruleAttrTypes, fields)
}
