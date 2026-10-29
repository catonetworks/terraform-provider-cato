package rpfrule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.RemotePortFwdAddRuleDataInput {
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
		knownValue{"rule.external_ip", model.ExternalIP, true},
		knownValue{"rule.external_port_range", model.ExternalPortRange, true},
		knownValue{"rule.forward_icmp", model.ForwardIcmp, true},
		knownValue{"rule.internal_ip", model.InternalIP, true},
		knownValue{"rule.internal_port_range", model.InternalPortRange, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.remote_ips", model.RemoteIps, true},
		knownValue{"rule.restriction_type", model.RestrictionType, true},
		knownValue{"rule.tracking", model.Tracking, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.RemotePortFwdAddRuleDataInput{}
	input.Description = model.Description.ValueString()
	input.Enabled = model.Enabled.ValueBool()
	input.ExternalIP = idname.PrepareIDName[cato_models.AllocatedIPRefInput](ctx, model.ExternalIP, diags)
	input.ExternalPortRange = buildCreateObjectExternalPortRange(ctx, model.ExternalPortRange, diags)
	input.ForwardICMP = model.ForwardIcmp.ValueBool()
	input.InternalIP = model.InternalIP.ValueString()
	input.InternalPortRange = buildCreateObjectInternalPortRange(ctx, model.InternalPortRange, diags)
	input.Name = model.Name.ValueString()
	input.RemoteIPs = buildCreateObjectRemoteIps(ctx, model.RemoteIps, diags)
	input.RestrictionType = cato_models.RemotePortFwdRestrictionType(model.RestrictionType.ValueString())
	input.Tracking = buildCreateObjectTracking(ctx, model.Tracking, diags)
	return input
}

func buildCreateObjectExternalPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.external_port_range must be known before applying.")
		return nil
	}
	var model RuleExternalPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.external_port_range.from", model.From, true},
		knownValue{"rule.external_port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeInput{}
	input.From = scalars.Port(model.From.ValueString())
	input.To = scalars.Port(model.To.ValueString())
	return input
}

func buildCreateObjectInternalPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.internal_port_range must be known before applying.")
		return nil
	}
	var model RuleInternalPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.internal_port_range.from", model.From, true},
		knownValue{"rule.internal_port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeInput{}
	input.From = scalars.Port(model.From.ValueString())
	input.To = scalars.Port(model.To.ValueString())
	return input
}

func buildCreateObjectRemoteIps(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.RemotePortFwdRemoteIpsInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.remote_ips must be known before applying.")
		return nil
	}
	var model RuleRemoteIpsModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.remote_ips.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.remote_ips.ip", model.IP, true},
		knownValue{"rule.remote_ips.ip_range", model.IPRange, true},
		knownValue{"rule.remote_ips.subnet", model.Subnet, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.RemotePortFwdRemoteIpsInput{}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildCreateObjectRemoteIpsIPRange(ctx, item.(types.Object), diags))
	}
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	return input
}

func buildCreateObjectRemoteIpsIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.remote_ips.ip_range must be known before applying.")
		return nil
	}
	var model RuleRemoteIpsIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.remote_ips.ip_range.from", model.From, true},
		knownValue{"rule.remote_ips.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildCreateObjectTracking(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRuleTrackingAlertInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.tracking must be known before applying.")
		return nil
	}
	var model RuleTrackingModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.tracking.enabled", model.Enabled, true},
		knownValue{"rule.tracking.frequency", model.Frequency, true},
		knownValue{"rule.tracking.mailing_list", model.MailingList, true},
		knownValue{"rule.tracking.subscription_group", model.SubscriptionGroup, true},
		knownValue{"rule.tracking.webhook", model.Webhook, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyRuleTrackingAlertInput{}
	input.Enabled = model.Enabled.ValueBool()
	input.Frequency = cato_models.PolicyRuleTrackingFrequencyEnum(model.Frequency.ValueString())
	input.MailingList = idname.PrepareIDNameSet[cato_models.SubscriptionMailingListRefInput](ctx, model.MailingList, diags)
	input.SubscriptionGroup = idname.PrepareIDNameSet[cato_models.SubscriptionGroupRefInput](ctx, model.SubscriptionGroup, diags)
	input.Webhook = idname.PrepareIDNameSet[cato_models.SubscriptionWebhookRefInput](ctx, model.Webhook, diags)
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.RemotePortFwdUpdateRuleDataInput {
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
		knownValue{"rule.external_ip", model.ExternalIP, true},
		knownValue{"rule.external_port_range", model.ExternalPortRange, true},
		knownValue{"rule.forward_icmp", model.ForwardIcmp, true},
		knownValue{"rule.internal_ip", model.InternalIP, true},
		knownValue{"rule.internal_port_range", model.InternalPortRange, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.remote_ips", model.RemoteIps, true},
		knownValue{"rule.restriction_type", model.RestrictionType, true},
		knownValue{"rule.tracking", model.Tracking, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.RemotePortFwdUpdateRuleDataInput{}
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	input.ExternalIP = idname.PrepareIDName[cato_models.AllocatedIPRefInput](ctx, model.ExternalIP, diags)
	input.ExternalPortRange = buildUpdateObjectExternalPortRange(ctx, model.ExternalPortRange, diags)
	vForwardICMP := model.ForwardIcmp.ValueBool()
	input.ForwardICMP = &vForwardICMP
	vInternalIP := model.InternalIP.ValueString()
	input.InternalIP = &vInternalIP
	input.InternalPortRange = buildUpdateObjectInternalPortRange(ctx, model.InternalPortRange, diags)
	vName := model.Name.ValueString()
	input.Name = &vName
	input.RemoteIPs = buildUpdateObjectRemoteIps(ctx, model.RemoteIps, diags)
	vRestrictionType := cato_models.RemotePortFwdRestrictionType(model.RestrictionType.ValueString())
	input.RestrictionType = &vRestrictionType
	input.Tracking = buildUpdateObjectTracking(ctx, model.Tracking, diags)
	return input
}

func buildUpdateObjectExternalPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.external_port_range must be known before applying.")
		return nil
	}
	var model RuleExternalPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.external_port_range.from", model.From, true},
		knownValue{"rule.external_port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeUpdateInput{}
	vFrom := scalars.Port(model.From.ValueString())
	input.From = &vFrom
	vTo := scalars.Port(model.To.ValueString())
	input.To = &vTo
	return input
}

func buildUpdateObjectInternalPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.internal_port_range must be known before applying.")
		return nil
	}
	var model RuleInternalPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.internal_port_range.from", model.From, true},
		knownValue{"rule.internal_port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeUpdateInput{}
	vFrom := scalars.Port(model.From.ValueString())
	input.From = &vFrom
	vTo := scalars.Port(model.To.ValueString())
	input.To = &vTo
	return input
}

func buildUpdateObjectRemoteIps(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.RemotePortFwdRemoteIpsUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.remote_ips must be known before applying.")
		return nil
	}
	var model RuleRemoteIpsModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.remote_ips.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.remote_ips.ip", model.IP, true},
		knownValue{"rule.remote_ips.ip_range", model.IPRange, true},
		knownValue{"rule.remote_ips.subnet", model.Subnet, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.RemotePortFwdRemoteIpsUpdateInput{}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildUpdateObjectRemoteIpsIPRange(ctx, item.(types.Object), diags))
	}
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	return input
}

func buildUpdateObjectRemoteIpsIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.remote_ips.ip_range must be known before applying.")
		return nil
	}
	var model RuleRemoteIpsIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.remote_ips.ip_range.from", model.From, true},
		knownValue{"rule.remote_ips.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildUpdateObjectTracking(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRuleTrackingAlertUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.tracking must be known before applying.")
		return nil
	}
	var model RuleTrackingModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.tracking.enabled", model.Enabled, true},
		knownValue{"rule.tracking.frequency", model.Frequency, true},
		knownValue{"rule.tracking.mailing_list", model.MailingList, true},
		knownValue{"rule.tracking.subscription_group", model.SubscriptionGroup, true},
		knownValue{"rule.tracking.webhook", model.Webhook, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyRuleTrackingAlertUpdateInput{}
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	vFrequency := cato_models.PolicyRuleTrackingFrequencyEnum(model.Frequency.ValueString())
	input.Frequency = &vFrequency
	input.MailingList = idname.PrepareIDNameSet[cato_models.SubscriptionMailingListRefInput](ctx, model.MailingList, diags)
	input.SubscriptionGroup = idname.PrepareIDNameSet[cato_models.SubscriptionGroupRefInput](ctx, model.SubscriptionGroup, diags)
	input.Webhook = idname.PrepareIDNameSet[cato_models.SubscriptionWebhookRefInput](ctx, model.Webhook, diags)
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

type read = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule

func parseRead(
	ctx context.Context, value *read, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExternalIP = idname.ParseIDName(ctx, value.ExternalIP, diags)
	model.ExternalPortRange = parseReadExternalPortRange(ctx, &value.ExternalPortRange, diags)
	model.ForwardIcmp = types.BoolValue(value.ForwardICMP)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.InternalIP = types.StringValue(value.InternalIP)
	model.InternalPortRange = parseReadInternalPortRange(ctx, &value.InternalPortRange, diags)
	model.Name = types.StringValue(value.Name)
	model.RemoteIps = parseReadRemoteIps(ctx, &value.RemoteIPs, diags)
	model.RestrictionType = types.StringValue(string(value.RestrictionType))
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.Tracking = parseReadTracking(ctx, &value.Tracking, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExtPortRange = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule_ExternalPortRange

func parseReadExternalPortRange(
	ctx context.Context, value *readExtPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.external_port_range object.")
		return types.ObjectNull(ruleExternalPortRangeAttrTypes)
	}
	model := RuleExternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readIntPortRange = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule_InternalPortRange

func parseReadInternalPortRange(
	ctx context.Context, value *readIntPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.internal_port_range object.")
		return types.ObjectNull(ruleInternalPortRangeAttrTypes)
	}
	model := RuleInternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleInternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readRemoteIps = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule_RemoteIPs

func parseReadRemoteIps(
	ctx context.Context, value *readRemoteIps, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips object.")
		return types.ObjectNull(ruleRemoteIpsAttrTypes)
	}
	model := RuleRemoteIpsModel{}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseReadRemoteIpsIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleRemoteIpsIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsAttrTypes, model)
	diags.Append(d...)
	return result
}

type readRemoteIpsIPRange = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule_RemoteIPs_IPRange

func parseReadRemoteIpsIPRange(
	ctx context.Context, value *readRemoteIpsIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips.ip_range object.")
		return types.ObjectNull(ruleRemoteIpsIPRangeAttrTypes)
	}
	model := RuleRemoteIpsIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSection = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule_Section

func parseReadSection(
	ctx context.Context, value *readSection, diags *diag.Diagnostics,
) types.Object {
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

type readTrack = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule_Tracking

func parseReadTracking(
	ctx context.Context, value *readTrack, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.tracking object.")
		return types.ObjectNull(ruleTrackingAttrTypes)
	}
	model := RuleTrackingModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	model.Frequency = types.StringValue(string(value.Frequency))
	model.MailingList = idname.ParseIDNameSet(ctx, value.MailingList, diags)
	if model.MailingList.IsNull() {
		model.MailingList = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SubscriptionGroup = idname.ParseIDNameSet(ctx, value.SubscriptionGroup, diags)
	if model.SubscriptionGroup.IsNull() {
		model.SubscriptionGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Webhook = idname.ParseIDNameSet(ctx, value.Webhook, diags)
	if model.Webhook.IsNull() {
		model.Webhook = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleTrackingAttrTypes, model)
	diags.Append(d...)
	return result
}

type add = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule

func parseAdd(
	ctx context.Context, value *add, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExternalIP = idname.ParseIDName(ctx, value.ExternalIP, diags)
	model.ExternalPortRange = parseAddExternalPortRange(ctx, &value.ExternalPortRange, diags)
	model.ForwardIcmp = types.BoolValue(value.ForwardICMP)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.InternalIP = types.StringValue(value.InternalIP)
	model.InternalPortRange = parseAddInternalPortRange(ctx, &value.InternalPortRange, diags)
	model.Name = types.StringValue(value.Name)
	model.RemoteIps = parseAddRemoteIps(ctx, &value.RemoteIPs, diags)
	model.RestrictionType = types.StringValue(string(value.RestrictionType))
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.Tracking = parseAddTracking(ctx, &value.Tracking, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExtPortRange = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule_ExternalPortRange

func parseAddExternalPortRange(
	ctx context.Context, value *addExtPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.external_port_range object.")
		return types.ObjectNull(ruleExternalPortRangeAttrTypes)
	}
	model := RuleExternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addIntPortRange = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule_InternalPortRange

func parseAddInternalPortRange(
	ctx context.Context, value *addIntPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.internal_port_range object.")
		return types.ObjectNull(ruleInternalPortRangeAttrTypes)
	}
	model := RuleInternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleInternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addRemoteIps = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule_RemoteIPs

func parseAddRemoteIps(
	ctx context.Context, value *addRemoteIps, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips object.")
		return types.ObjectNull(ruleRemoteIpsAttrTypes)
	}
	model := RuleRemoteIpsModel{}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseAddRemoteIpsIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleRemoteIpsIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsAttrTypes, model)
	diags.Append(d...)
	return result
}

type addRemoteIpsIPRange = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule_RemoteIPs_IPRange

func parseAddRemoteIpsIPRange(
	ctx context.Context, value *addRemoteIpsIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips.ip_range object.")
		return types.ObjectNull(ruleRemoteIpsIPRangeAttrTypes)
	}
	model := RuleRemoteIpsIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSection = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule_Section

func parseAddSection(
	ctx context.Context, value *addSection, diags *diag.Diagnostics,
) types.Object {
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

type addTrack = cato.PolicyRemotePortFwdAddRule_Policy_RemotePortFwd_AddRule_Rule_Rule_Tracking

func parseAddTracking(
	ctx context.Context, value *addTrack, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.tracking object.")
		return types.ObjectNull(ruleTrackingAttrTypes)
	}
	model := RuleTrackingModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	model.Frequency = types.StringValue(string(value.Frequency))
	model.MailingList = idname.ParseIDNameSet(ctx, value.MailingList, diags)
	if model.MailingList.IsNull() {
		model.MailingList = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SubscriptionGroup = idname.ParseIDNameSet(ctx, value.SubscriptionGroup, diags)
	if model.SubscriptionGroup.IsNull() {
		model.SubscriptionGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Webhook = idname.ParseIDNameSet(ctx, value.Webhook, diags)
	if model.Webhook.IsNull() {
		model.Webhook = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleTrackingAttrTypes, model)
	diags.Append(d...)
	return result
}

type update = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule

func parseUpdate(
	ctx context.Context, value *update, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExternalIP = idname.ParseIDName(ctx, value.ExternalIP, diags)
	model.ExternalPortRange = parseUpdateExternalPortRange(ctx, &value.ExternalPortRange, diags)
	model.ForwardIcmp = types.BoolValue(value.ForwardICMP)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.InternalIP = types.StringValue(value.InternalIP)
	model.InternalPortRange = parseUpdateInternalPortRange(ctx, &value.InternalPortRange, diags)
	model.Name = types.StringValue(value.Name)
	model.RemoteIps = parseUpdateRemoteIps(ctx, &value.RemoteIPs, diags)
	model.RestrictionType = types.StringValue(string(value.RestrictionType))
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.Tracking = parseUpdateTracking(ctx, &value.Tracking, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExtPortRange = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule_ExternalPortRange

func parseUpdateExternalPortRange(
	ctx context.Context, value *updateExtPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.external_port_range object.")
		return types.ObjectNull(ruleExternalPortRangeAttrTypes)
	}
	model := RuleExternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateIntPortRange = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule_InternalPortRange

func parseUpdateInternalPortRange(
	ctx context.Context, value *updateIntPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.internal_port_range object.")
		return types.ObjectNull(ruleInternalPortRangeAttrTypes)
	}
	model := RuleInternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleInternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateRemoteIps = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule_RemoteIPs

func parseUpdateRemoteIps(
	ctx context.Context, value *updateRemoteIps, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips object.")
		return types.ObjectNull(ruleRemoteIpsAttrTypes)
	}
	model := RuleRemoteIpsModel{}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseUpdateRemoteIpsIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleRemoteIpsIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateRemoteIpsIPRange = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule_RemoteIPs_IPRange

func parseUpdateRemoteIpsIPRange(
	ctx context.Context, value *updateRemoteIpsIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips.ip_range object.")
		return types.ObjectNull(ruleRemoteIpsIPRangeAttrTypes)
	}
	model := RuleRemoteIpsIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSection = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule_Section

func parseUpdateSection(
	ctx context.Context, value *updateSection, diags *diag.Diagnostics,
) types.Object {
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

type updateTrack = cato.PolicyRemotePortFwdUpdateRule_Policy_RemotePortFwd_UpdateRule_Rule_Rule_Tracking

func parseUpdateTracking(
	ctx context.Context, value *updateTrack, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.tracking object.")
		return types.ObjectNull(ruleTrackingAttrTypes)
	}
	model := RuleTrackingModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	model.Frequency = types.StringValue(string(value.FrequencyPolicyRuleTrackingAlert))
	model.MailingList = idname.ParseIDNameSet(ctx, value.MailingList, diags)
	if model.MailingList.IsNull() {
		model.MailingList = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SubscriptionGroup = idname.ParseIDNameSet(ctx, value.SubscriptionGroup, diags)
	if model.SubscriptionGroup.IsNull() {
		model.SubscriptionGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Webhook = idname.ParseIDNameSet(ctx, value.Webhook, diags)
	if model.Webhook.IsNull() {
		model.Webhook = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleTrackingAttrTypes, model)
	diags.Append(d...)
	return result
}

type move = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule

func parseMove(
	ctx context.Context, value *move, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.ExternalIP = idname.ParseIDName(ctx, value.ExternalIP, diags)
	model.ExternalPortRange = parseMoveExternalPortRange(ctx, &value.ExternalPortRange, diags)
	model.ForwardIcmp = types.BoolValue(value.ForwardICMP)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.InternalIP = types.StringValue(value.InternalIP)
	model.InternalPortRange = parseMoveInternalPortRange(ctx, &value.InternalPortRange, diags)
	model.Name = types.StringValue(value.Name)
	model.RemoteIps = parseMoveRemoteIps(ctx, &value.RemoteIPs, diags)
	model.RestrictionType = types.StringValue(string(value.RestrictionType))
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.Tracking = parseMoveTracking(ctx, &value.Tracking, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExtPortRange = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule_ExternalPortRange

func parseMoveExternalPortRange(
	ctx context.Context, value *moveExtPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.external_port_range object.")
		return types.ObjectNull(ruleExternalPortRangeAttrTypes)
	}
	model := RuleExternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveIntPortRange = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule_InternalPortRange

func parseMoveInternalPortRange(
	ctx context.Context, value *moveIntPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.internal_port_range object.")
		return types.ObjectNull(ruleInternalPortRangeAttrTypes)
	}
	model := RuleInternalPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleInternalPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveRemoteIps = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule_RemoteIPs

func parseMoveRemoteIps(
	ctx context.Context, value *moveRemoteIps, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips object.")
		return types.ObjectNull(ruleRemoteIpsAttrTypes)
	}
	model := RuleRemoteIpsModel{}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseMoveRemoteIpsIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleRemoteIpsIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveRemoteIpsIPRange = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule_RemoteIPs_IPRange

func parseMoveRemoteIpsIPRange(
	ctx context.Context, value *moveRemoteIpsIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.remote_ips.ip_range object.")
		return types.ObjectNull(ruleRemoteIpsIPRangeAttrTypes)
	}
	model := RuleRemoteIpsIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleRemoteIpsIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSection = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule_Section

func parseMoveSection(
	ctx context.Context, value *moveSection, diags *diag.Diagnostics,
) types.Object {
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

type moveTrack = cato.PolicyRemotePortFwdMoveRule_Policy_RemotePortFwd_MoveRule_Rule_Rule_Tracking

func parseMoveTracking(
	ctx context.Context, value *moveTrack, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.tracking object.")
		return types.ObjectNull(ruleTrackingAttrTypes)
	}
	model := RuleTrackingModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	model.Frequency = types.StringValue(string(value.Frequency))
	model.MailingList = idname.ParseIDNameSet(ctx, value.MailingList, diags)
	if model.MailingList.IsNull() {
		model.MailingList = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SubscriptionGroup = idname.ParseIDNameSet(ctx, value.SubscriptionGroup, diags)
	if model.SubscriptionGroup.IsNull() {
		model.SubscriptionGroup = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Webhook = idname.ParseIDNameSet(ctx, value.Webhook, diags)
	if model.Webhook.IsNull() {
		model.Webhook = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleTrackingAttrTypes, model)
	diags.Append(d...)
	return result
}
