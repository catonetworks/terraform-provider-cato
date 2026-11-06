package rpfrule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	refinput "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/refinput"
)

var rpfAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func portStr(p *cato_scalars.Port) string {
	if p == nil {
		return ""
	}
	return string(*p)
}

func rpfPortRange(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	pr := RemotePortFwdPortRangePlan{}
	diags.Append(obj.As(ctx, &pr, rpfAsOpts)...)
	if diags.HasError() {
		return nil
	}
	return &cato_models.PortRangeInput{
		From: cato_scalars.Port(pr.From.ValueString()),
		To:   cato_scalars.Port(pr.To.ValueString()),
	}
}

func rpfRemoteIps(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.RemotePortFwdRemoteIpsInput {
	out := &cato_models.RemotePortFwdRemoteIpsInput{
		GlobalIPRange: []*cato_models.GlobalIPRangeRefInput{},
		IP:            []string{},
		IPRange:       []*cato_models.IPAddressRangeInput{},
		Subnet:        []string{},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	ri := RemotePortFwdRemoteIpsPlan{}
	diags.Append(obj.As(ctx, &ri, rpfAsOpts)...)
	if diags.HasError() {
		return out
	}
	if g := refinput.FromSet(ctx, ri.GlobalIPRange, diags, "global IP range", refinput.GlobalIPRange); g != nil {
		out.GlobalIPRange = g
	}
	if ips := refinput.StringsFromList(ctx, ri.IP, diags); ips != nil {
		out.IP = ips
	}
	if !ri.IPRange.IsNull() && !ri.IPRange.IsUnknown() {
		var rows []RemotePortFwdPortRangePlan
		diags.Append(ri.IPRange.ElementsAs(ctx, &rows, false)...)
		for _, row := range rows {
			out.IPRange = append(out.IPRange, &cato_models.IPAddressRangeInput{
				From: row.From.ValueString(),
				To:   row.To.ValueString(),
			})
		}
	}
	if sn := refinput.StringsFromList(ctx, ri.Subnet, diags); sn != nil {
		out.Subnet = sn
	}
	return out
}

func rpfTracking(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyRuleTrackingAlertInput {
	out := &cato_models.PolicyRuleTrackingAlertInput{
		Enabled:           false,
		Frequency:         cato_models.PolicyRuleTrackingFrequencyEnumHourly,
		MailingList:       []*cato_models.SubscriptionMailingListRefInput{},
		SubscriptionGroup: []*cato_models.SubscriptionGroupRefInput{},
		Webhook:           []*cato_models.SubscriptionWebhookRefInput{},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	tr := RemotePortFwdTrackingPlan{}
	diags.Append(obj.As(ctx, &tr, rpfAsOpts)...)
	if diags.HasError() {
		return out
	}
	out.Enabled = tr.Enabled.ValueBool()
	if !tr.Frequency.IsNull() && !tr.Frequency.IsUnknown() && tr.Frequency.ValueString() != "" {
		out.Frequency = cato_models.PolicyRuleTrackingFrequencyEnum(tr.Frequency.ValueString())
	}
	if m := refinput.FromSet(ctx, tr.MailingList, diags, "mailing list", refinput.SubscriptionMailingList); m != nil {
		out.MailingList = m
	}
	if g := refinput.FromSet(ctx, tr.SubscriptionGroup, diags, "subscription group", refinput.SubscriptionGroup); g != nil {
		out.SubscriptionGroup = g
	}
	if w := refinput.FromSet(ctx, tr.Webhook, diags, "webhook", refinput.SubscriptionWebhook); w != nil {
		out.Webhook = w
	}
	return out
}

func hydrateRemotePortFwdAddRuleInput(
	ctx context.Context,
	plan RemotePortFwdRule,
) (cato_models.RemotePortFwdAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.RemotePortFwdAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, rpfAsOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := RemotePortFwdRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, rpfAsOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.RemotePortFwdAddRuleDataInput{
		Name:              rule.Name.ValueString(),
		Description:       rule.Description.ValueString(),
		Enabled:           rule.Enabled.ValueBool(),
		InternalIP:        rule.InternalIP.ValueString(),
		ForwardICMP:       rule.ForwardICMP.ValueBool(),
		RestrictionType:   cato_models.RemotePortFwdRestrictionType(rule.RestrictionType.ValueString()),
		ExternalPortRange: rpfPortRange(ctx, rule.ExternalPortRange, &diags),
		InternalPortRange: rpfPortRange(ctx, rule.InternalPortRange, &diags),
		RemoteIPs:         rpfRemoteIps(ctx, rule.RemoteIPs, &diags),
		Tracking:          rpfTracking(ctx, rule.Tracking, &diags),
	}
	if eip, ok := refinput.FromObject(ctx, rule.ExternalIP, &diags, "allocated IP", refinput.AllocatedIP); ok {
		data.ExternalIP = eip
	}

	out.Rule = data
	return out, diags
}

func hydrateRemotePortFwdUpdateRuleInput(
	ctx context.Context,
	plan RemotePortFwdRule,
) (cato_models.RemotePortFwdUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := RemotePortFwdRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, rpfAsOpts)...)
	if diags.HasError() {
		return cato_models.RemotePortFwdUpdateRuleInput{}, diags
	}

	ext := rpfPortRange(ctx, rule.ExternalPortRange, &diags)
	intr := rpfPortRange(ctx, rule.InternalPortRange, &diags)
	ri := rpfRemoteIps(ctx, rule.RemoteIPs, &diags)
	tr := rpfTracking(ctx, rule.Tracking, &diags)

	upd := cato_models.RemotePortFwdUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.RemotePortFwdUpdateRuleDataInput{
			Name:            rule.Name.ValueStringPointer(),
			Description:     rule.Description.ValueStringPointer(),
			Enabled:         rule.Enabled.ValueBoolPointer(),
			InternalIP:      rule.InternalIP.ValueStringPointer(),
			ForwardICMP:     rule.ForwardICMP.ValueBoolPointer(),
			RestrictionType: (*cato_models.RemotePortFwdRestrictionType)(rule.RestrictionType.ValueStringPointer()),
			RemoteIPs: &cato_models.RemotePortFwdRemoteIpsUpdateInput{
				GlobalIPRange: ri.GlobalIPRange,
				IP:            ri.IP,
				IPRange:       ri.IPRange,
				Subnet:        ri.Subnet,
			},
			Tracking: &cato_models.PolicyRuleTrackingAlertUpdateInput{
				Enabled:           &tr.Enabled,
				Frequency:         &tr.Frequency,
				MailingList:       tr.MailingList,
				SubscriptionGroup: tr.SubscriptionGroup,
				Webhook:           tr.Webhook,
			},
		},
	}
	if ext != nil {
		upd.Rule.ExternalPortRange = &cato_models.PortRangeUpdateInput{From: &ext.From, To: &ext.To}
	}
	if intr != nil {
		upd.Rule.InternalPortRange = &cato_models.PortRangeUpdateInput{From: &intr.From, To: &intr.To}
	}
	if eip, ok := refinput.FromObject(ctx, rule.ExternalIP, &diags, "allocated IP", refinput.AllocatedIP); ok {
		upd.Rule.ExternalIP = eip
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *RemotePortFwdRuleRulePlan,
	_ RemotePortFwdRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, RemotePortFwdRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *RemotePortFwdRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	_ RemotePortFwdRuleRulePlan,
	priorHydrated *RemotePortFwdRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, RemotePortFwdRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *RemotePortFwdRule,
	cur *cato_go_sdk.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule,
	ruleObj basetypes.ObjectValue,
	diagnostics *diag.Diagnostics,
) {
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		diagnostics.Append(d...)
		state.At = atObj
	}

	state.ID = types.StringValue(cur.GetID())
	state.Rule = ruleObj
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan RemotePortFwdRule,
	diagnostics *diag.Diagnostics,
) (
	resultMove cato_models.PolicyMoveRuleInput,
	resultAtAsOpts basetypes.ObjectAsOptions,
) {
	resultMove = cato_models.PolicyMoveRuleInput{}
	resultAtAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	if !plan.At.IsNull() {
		pos := policyrule.Position{}
		diagnostics.Append(plan.At.As(ctx, &pos, resultAtAsOpts)...)
		resultMove.To = &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer()),
			Ref:      pos.Ref.ValueStringPointer(),
		}
	}
	return resultMove, resultAtAsOpts
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	_ RemotePortFwdRuleRulePlan,
	priorRulePlan *RemotePortFwdRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, RemotePortFwdRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertUpdateInput3 converts between the Terraform model and SDK values.
func convertUpdateInput3(pos policyrule.Position) (resultAtObj basetypes.ObjectValue, resultAdiags diag.Diagnostics) {
	resultAtObj, resultAdiags = types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": pos.Position,
		"ref":      pos.Ref,
	})
	return resultAtObj, resultAdiags
}

// convertUpdateInput4 converts between the Terraform model and SDK values.
func convertUpdateInput4(
	rule RemotePortFwdRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt RemotePortFwdRule) {
	resultSt = RemotePortFwdRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule RemotePortFwdRuleRulePlan) (resultRm cato_models.RemotePortFwdRemoveRuleInput) {
	resultRm = cato_models.RemotePortFwdRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateRemotePortFwdRuleStateFromClient maps a policy read rule into Terraform nested objects.
//
//nolint:funlen // Hydrate all rule fields together.
func hydrateRemotePortFwdRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule,
) (RemotePortFwdRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := RemotePortFwdRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.ExternalIP = types.ObjectNull(idname.RefAttrTypes)
	out.ExternalPortRange = types.ObjectNull(RemotePortFwdPortRangeAttrTypes)
	out.InternalPortRange = types.ObjectNull(RemotePortFwdPortRangeAttrTypes)
	out.RemoteIPs = types.ObjectNull(RemotePortFwdRemoteIpsAttrTypes)
	out.Tracking = types.ObjectNull(RemotePortFwdTrackingAttrTypes)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	out.InternalIP = types.StringValue(r.GetInternalIP())
	out.ForwardICMP = types.BoolValue(r.GetForwardICMP())
	if r.GetRestrictionType() != nil {
		out.RestrictionType = types.StringValue(string(*r.GetRestrictionType()))
	}

	if eip := r.GetExternalIP(); eip != nil && (eip.GetID() != "" || eip.GetName() != "") {
		o, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(eip.GetID()),
			"name": types.StringValue(eip.GetName()),
		})
		diags.Append(d...)
		out.ExternalIP = o
	}

	if pr := r.GetExternalPortRange(); pr != nil {
		o, d := types.ObjectValue(RemotePortFwdPortRangeAttrTypes, map[string]attr.Value{
			"from": types.StringValue(portStr(pr.GetFrom())),
			"to":   types.StringValue(portStr(pr.GetTo())),
		})
		diags.Append(d...)
		out.ExternalPortRange = o
	}
	if pr := r.GetInternalPortRange(); pr != nil {
		o, d := types.ObjectValue(RemotePortFwdPortRangeAttrTypes, map[string]attr.Value{
			"from": types.StringValue(portStr(pr.GetFrom())),
			"to":   types.StringValue(portStr(pr.GetTo())),
		})
		diags.Append(d...)
		out.InternalPortRange = o
	}

	if ri := r.GetRemoteIPs(); ri != nil {
		ranges := []attr.Value{}
		for _, rr := range ri.GetIPRange() {
			if rr == nil {
				continue
			}
			o, d := types.ObjectValue(RemotePortFwdPortlessRangeAttrTypes, map[string]attr.Value{
				"from": types.StringValue(rr.GetFrom()),
				"to":   types.StringValue(rr.GetTo()),
			})
			diags.Append(d...)
			ranges = append(ranges, o)
		}
		rangeList := types.ListNull(types.ObjectType{AttrTypes: RemotePortFwdPortlessRangeAttrTypes})
		if len(ranges) > 0 {
			var d diag.Diagnostics
			rangeList, d = types.ListValue(types.ObjectType{AttrTypes: RemotePortFwdPortlessRangeAttrTypes}, ranges)
			diags.Append(d...)
		}
		o, d := types.ObjectValue(RemotePortFwdRemoteIpsAttrTypes, map[string]attr.Value{
			"global_ip_range": oldidname.ParseNameIDList(ctx, ri.GetGlobalIPRange(), "rule.remote_ips.global_ip_range"),
			"ip":              convert.ParseList(ctx, types.StringType, ri.GetIP(), "rule.remote_ips.ip"),
			"ip_range":        rangeList,
			"subnet":          convert.ParseList(ctx, types.StringType, ri.GetSubnet(), "rule.remote_ips.subnet"),
		})
		diags.Append(d...)
		out.RemoteIPs = o
	}

	if tr := r.GetTracking(); tr != nil {
		freq := ""
		if tr.GetFrequency() != nil {
			freq = string(*tr.GetFrequency())
		}
		o, d := types.ObjectValue(RemotePortFwdTrackingAttrTypes, map[string]attr.Value{
			"enabled":            types.BoolValue(tr.GetEnabled()),
			"frequency":          types.StringValue(freq),
			"mailing_list":       oldidname.ParseNameIDList(ctx, tr.GetMailingList(), "rule.tracking.mailing_list"),
			"subscription_group": oldidname.ParseNameIDList(ctx, tr.GetSubscriptionGroup(), "rule.tracking.subscription_group"),
			"webhook":            oldidname.ParseNameIDList(ctx, tr.GetWebhook(), "rule.tracking.webhook"),
		})
		diags.Append(d...)
		out.Tracking = o
	}

	return out, diags
}
