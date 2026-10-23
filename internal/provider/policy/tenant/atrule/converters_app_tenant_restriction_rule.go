package atrule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	ifrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
	wfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
	applicationcontrol "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/applicationcontrol"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

//nolint:gocyclo,funlen
func hydrateAppTenantRestrictionAddRuleInput(
	ctx context.Context,
	plan AppTenantRestrictionRule,
) (cato_models.AppTenantRestrictionAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.AppTenantRestrictionAddRuleInput{}

	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, asOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := AppTenantRestrictionRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, asOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.AppTenantRestrictionAddRuleDataInput{
		Name:        rule.Name.ValueString(),
		Description: rule.Description.ValueString(),
		Enabled:     rule.Enabled.ValueBool(),
		Action:      cato_models.AppTenantRestrictionActionEnum(rule.Action.ValueString()),
		Severity:    cato_models.AppTenantRestrictionSeverityEnum(rule.Severity.ValueString()),
	}

	if !rule.Application.IsNull() && !rule.Application.IsUnknown() {
		app := ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost{}
		diags.Append(rule.Application.As(ctx, &app, asOpts)...)
		if !diags.HasError() {
			ref, err := utils.TransformObjectRefInput(app)
			if err != nil {
				diags.AddError("Invalid application reference", err.Error())
				return out, diags
			}
			data.Application = &cato_models.ApplicationRefInput{
				By:    cato_models.ObjectRefBy(ref.By),
				Input: ref.Input,
			}
		}
	}

	if !rule.Headers.IsNull() && !rule.Headers.IsUnknown() {
		var headers []AppTenantRestrictionHeaderPlan
		diags.Append(rule.Headers.ElementsAs(ctx, &headers, false)...)
		for _, h := range headers {
			data.Headers = append(data.Headers, &cato_models.AppTenantRestrictionHeaderValueInput{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
	}

	data.Schedule = &cato_models.PolicyScheduleInput{ActiveOn: cato_models.PolicyActiveOnEnumAlways}
	if !rule.Schedule.IsNull() && !rule.Schedule.IsUnknown() {
		scheduleAsOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		sch := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(rule.Schedule.As(ctx, &sch, scheduleAsOpts)...)
		if !diags.HasError() {
			if !sch.ActiveOn.IsNull() && !sch.ActiveOn.IsUnknown() && sch.ActiveOn.ValueString() != "" {
				data.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(sch.ActiveOn.ValueString())
			}
			active := data.Schedule.ActiveOn
			if active == cato_models.PolicyActiveOnEnumCustomTimeframe &&
				!sch.CustomTimeframe.IsNull() && !sch.CustomTimeframe.IsUnknown() {
				ctf := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
				diags.Append(sch.CustomTimeframe.As(ctx, &ctf, scheduleAsOpts)...)
				from := ctf.From.ValueString()
				to := ctf.To.ValueString()
				if from != "" || to != "" {
					data.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{
						From: from,
						To:   to,
					}
				}
			}
			if active == cato_models.PolicyActiveOnEnumCustomRecurring &&
				!sch.CustomRecurring.IsNull() && !sch.CustomRecurring.IsUnknown() {
				cr := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
				diags.Append(sch.CustomRecurring.As(ctx, &cr, scheduleAsOpts)...)
				from := cr.From.ValueString()
				to := cr.To.ValueString()
				if from != "" && to != "" {
					data.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{
						From: cato_scalars.Time(from),
						To:   cato_scalars.Time(to),
					}
					if !cr.Days.IsNull() && !cr.Days.IsUnknown() {
						diags.Append(cr.Days.ElementsAs(ctx, &data.Schedule.CustomRecurring.Days, false)...)
					}
				}
			}
		}
	}

	data.Source = &cato_models.AppTenantRestrictionSourceInput{}
	if !rule.Source.IsNull() && !rule.Source.IsUnknown() {
		acAdd, _, sdiags := applicationcontrol.ApplicationControlSourcePairFromTerraformObject(ctx, rule.Source)
		diags.Append(sdiags...)
		if !diags.HasError() && acAdd != nil {
			data.Source = cloneAppTenantRestrictionSource(acAdd)
		}
	}

	out.Rule = data
	return out, diags
}

// atrUpdateScheduleFromPlan builds the PolicyScheduleUpdateInput for an update call.
//
//nolint:gocyclo
func atrUpdateScheduleFromPlan(
	ctx context.Context,
	rule AppTenantRestrictionRuleRulePlan,
	diags *diag.Diagnostics,
) *cato_models.PolicyScheduleUpdateInput {
	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	defaultActive := cato_models.PolicyActiveOnEnumAlways
	sched := &cato_models.PolicyScheduleUpdateInput{ActiveOn: &defaultActive}
	if rule.Schedule.IsNull() || rule.Schedule.IsUnknown() {
		return sched
	}
	sch := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
	diags.Append(rule.Schedule.As(ctx, &sch, asOpts)...)
	if diags.HasError() {
		return sched
	}
	if !sch.ActiveOn.IsNull() && !sch.ActiveOn.IsUnknown() && sch.ActiveOn.ValueString() != "" {
		sched.ActiveOn = (*cato_models.PolicyActiveOnEnum)(sch.ActiveOn.ValueStringPointer())
	}
	active := sched.ActiveOn
	if active != nil && *active == cato_models.PolicyActiveOnEnumCustomTimeframe &&
		!sch.CustomTimeframe.IsNull() && !sch.CustomTimeframe.IsUnknown() {
		ctf := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
		diags.Append(sch.CustomTimeframe.As(ctx, &ctf, asOpts)...)
		from, to := ctf.From.ValueStringPointer(), ctf.To.ValueStringPointer()
		if (from != nil && *from != "") || (to != nil && *to != "") {
			sched.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{From: from, To: to}
		}
	}
	if active != nil && *active == cato_models.PolicyActiveOnEnumCustomRecurring &&
		!sch.CustomRecurring.IsNull() && !sch.CustomRecurring.IsUnknown() {
		cr := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
		diags.Append(sch.CustomRecurring.As(ctx, &cr, asOpts)...)
		from, to := cr.From.ValueString(), cr.To.ValueString()
		if from != "" && to != "" {
			sched.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{
				From: (*cato_scalars.Time)(cr.From.ValueStringPointer()),
				To:   (*cato_scalars.Time)(cr.To.ValueStringPointer()),
			}
			if !cr.Days.IsNull() && !cr.Days.IsUnknown() {
				diags.Append(cr.Days.ElementsAs(ctx, &sched.CustomRecurring.Days, false)...)
			}
		}
	}
	return sched
}

func hydrateAppTenantRestrictionUpdateRuleInput(
	ctx context.Context,
	plan AppTenantRestrictionRule,
) (cato_models.AppTenantRestrictionUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	rule := AppTenantRestrictionRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, asOpts)...)
	if diags.HasError() {
		return cato_models.AppTenantRestrictionUpdateRuleInput{}, diags
	}

	upd := cato_models.AppTenantRestrictionUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.AppTenantRestrictionUpdateRuleDataInput{
			Name:        rule.Name.ValueStringPointer(),
			Description: rule.Description.ValueStringPointer(),
			Enabled:     rule.Enabled.ValueBoolPointer(),
			Action:      (*cato_models.AppTenantRestrictionActionEnum)(rule.Action.ValueStringPointer()),
			Severity:    (*cato_models.AppTenantRestrictionSeverityEnum)(rule.Severity.ValueStringPointer()),
		},
	}

	if !rule.Application.IsNull() && !rule.Application.IsUnknown() {
		app := ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost{}
		diags.Append(rule.Application.As(ctx, &app, asOpts)...)
		if !diags.HasError() {
			ref, err := utils.TransformObjectRefInput(app)
			if err != nil {
				diags.AddError("Invalid application reference", err.Error())
				return upd, diags
			}
			upd.Rule.Application = &cato_models.ApplicationRefInput{
				By:    cato_models.ObjectRefBy(ref.By),
				Input: ref.Input,
			}
		}
	}

	if !rule.Headers.IsNull() && !rule.Headers.IsUnknown() {
		var headers []AppTenantRestrictionHeaderPlan
		diags.Append(rule.Headers.ElementsAs(ctx, &headers, false)...)
		for _, h := range headers {
			upd.Rule.Headers = append(upd.Rule.Headers, &cato_models.AppTenantRestrictionHeaderValueInput{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
	}

	upd.Rule.Schedule = atrUpdateScheduleFromPlan(ctx, rule, &diags)

	upd.Rule.Source = &cato_models.AppTenantRestrictionSourceUpdateInput{}
	if !rule.Source.IsNull() && !rule.Source.IsUnknown() {
		_, acUpd, sdiags := applicationcontrol.ApplicationControlSourcePairFromTerraformObject(ctx, rule.Source)
		diags.Append(sdiags...)
		if !diags.HasError() && acUpd != nil {
			upd.Rule.Source = cloneAppTenantRestrictionSourceUpdate(acUpd)
		}
	}

	return upd, diags
}

func cloneAppTenantRestrictionSourceUpdate(
	in *cato_models.ApplicationControlSourceUpdateInput,
) *cato_models.AppTenantRestrictionSourceUpdateInput {
	if in == nil {
		return nil
	}
	return &cato_models.AppTenantRestrictionSourceUpdateInput{
		Country:           in.Country,
		FloatingSubnet:    in.FloatingSubnet,
		GlobalIPRange:     in.GlobalIPRange,
		Group:             in.Group,
		Host:              in.Host,
		IP:                in.IP,
		IPRange:           in.IPRange,
		NetworkInterface:  in.NetworkInterface,
		Site:              in.Site,
		SiteNetworkSubnet: in.SiteNetworkSubnet,
		Subnet:            in.Subnet,
		SystemGroup:       in.SystemGroup,
		User:              in.User,
		UsersGroup:        in.UsersGroup,
	}
}

func cloneAppTenantRestrictionSource(
	in *cato_models.ApplicationControlSourceInput,
) *cato_models.AppTenantRestrictionSourceInput {
	if in == nil {
		return nil
	}
	return &cato_models.AppTenantRestrictionSourceInput{
		Country:           in.Country,
		FloatingSubnet:    in.FloatingSubnet,
		GlobalIPRange:     in.GlobalIPRange,
		Group:             in.Group,
		Host:              in.Host,
		IP:                in.IP,
		IPRange:           in.IPRange,
		NetworkInterface:  in.NetworkInterface,
		Site:              in.Site,
		SiteNetworkSubnet: in.SiteNetworkSubnet,
		Subnet:            in.Subnet,
		SystemGroup:       in.SystemGroup,
		User:              in.User,
		UsersGroup:        in.UsersGroup,
	}
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *AppTenantRestrictionRuleRulePlan,
	planRule AppTenantRestrictionRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	if !planRule.Headers.IsNull() && !planRule.Headers.IsUnknown() {
		rulePlan.Headers = planRule.Headers
	}

	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *AppTenantRestrictionRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	rulePlan AppTenantRestrictionRuleRulePlan,
	priorHydrated *AppTenantRestrictionRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	if !rulePlan.Headers.IsNull() && !rulePlan.Headers.IsUnknown() {
		hydrated.Headers = rulePlan.Headers
	}

	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *AppTenantRestrictionRule,
	cur *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule,
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
	plan AppTenantRestrictionRule,
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
	rule AppTenantRestrictionRuleRulePlan,
	priorRulePlan *AppTenantRestrictionRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	if !rule.Headers.IsNull() && !rule.Headers.IsUnknown() {
		rulePlan.Headers = rule.Headers
	}

	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, rulePlan)
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
	rule AppTenantRestrictionRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt AppTenantRestrictionRule) {
	resultSt = AppTenantRestrictionRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule AppTenantRestrictionRuleRulePlan) (resultRm cato_models.AppTenantRestrictionRemoveRuleInput) {
	resultRm = cato_models.AppTenantRestrictionRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateAppTenantRestrictionRuleStateFromClient maps a policy read rule into Terraform nested objects.
//
//nolint:gocyclo,funlen
func hydrateAppTenantRestrictionRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule,
) (AppTenantRestrictionRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := AppTenantRestrictionRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.Application = types.ObjectNull(idname.RefAttrTypes)
	out.Headers = types.ListNull(types.ObjectType{AttrTypes: AppTenantRestrictionHeaderAttrTypes})
	out.Source = types.ObjectNull(applicationcontrol.ApplicationControlSourceAttrTypes)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	if r.GetAction() != nil {
		out.Action = types.StringValue(string(*r.GetAction()))
	}
	if r.GetSeverity() != nil {
		out.Severity = types.StringValue(string(*r.GetSeverity()))
	}

	if app := r.GetApplication(); app != nil && (app.GetID() != "" || app.GetName() != "") {
		appObj, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(app.GetID()),
			"name": types.StringValue(app.GetName()),
		})
		diags.Append(d...)
		out.Application = appObj
	}

	if hdrs := r.GetHeaders(); len(hdrs) > 0 {
		elems := make([]attr.Value, 0, len(hdrs))
		for _, h := range hdrs {
			if h == nil {
				continue
			}
			o, d := types.ObjectValue(AppTenantRestrictionHeaderAttrTypes, map[string]attr.Value{
				"name":  types.StringValue(h.GetName()),
				"value": types.StringValue(h.GetValue()),
			})
			diags.Append(d...)
			elems = append(elems, o)
		}
		lst, d := types.ListValue(types.ObjectType{AttrTypes: AppTenantRestrictionHeaderAttrTypes}, elems)
		diags.Append(d...)
		out.Headers = lst
	}

	sch := r.GetSchedule()
	if sch != nil {
		var ctfObj types.Object
		if ct := sch.GetCustomTimeframeAppTenantRestriction(); ct != nil {
			var dtf diag.Diagnostics
			ctfObj, dtf = types.ObjectValue(policyschedule.CustomTimeframeAttrTypes, map[string]attr.Value{
				"from": types.StringValue(utils.NormalizePolicyTimestampString(ct.GetFrom())),
				"to":   types.StringValue(utils.NormalizePolicyTimestampString(ct.GetTo())),
			})
			diags.Append(dtf...)
		} else {
			ctfObj = types.ObjectNull(policyschedule.CustomTimeframeAttrTypes)
		}

		var crObj types.Object
		if cr := sch.GetCustomRecurringAppTenantRestriction(); cr != nil {
			fromS := ""
			toS := ""
			if cr.GetFrom() != nil {
				fromS = utils.NormalizePolicyRecurringClock(string(*cr.GetFrom()))
			}
			if cr.GetTo() != nil {
				toS = utils.NormalizePolicyRecurringClock(string(*cr.GetTo()))
			}
			days := convert.ParseList(ctx, types.StringType, cr.GetDays(), "rule.schedule.custom_recurring.days")
			var dcr diag.Diagnostics
			crObj, dcr = types.ObjectValue(policyschedule.FromToDaysAttrTypes, map[string]attr.Value{
				"from": types.StringValue(fromS),
				"to":   types.StringValue(toS),
				"days": days,
			})
			diags.Append(dcr...)
		} else {
			crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
		}

		active := ""
		if sch.GetActiveOn() != nil {
			active = string(*sch.GetActiveOn())
		}
		schObj, d := types.ObjectValue(policyschedule.ScheduleAttrTypes, map[string]attr.Value{
			"active_on":        types.StringValue(active),
			"custom_timeframe": ctfObj,
			"custom_recurring": crObj,
		})
		diags.Append(d...)
		out.Schedule = schObj
	}

	src := r.GetSource()
	if src != nil {
		srcAttrs := map[string]attr.Value{
			"country":             oldidname.ParseNameIDList(ctx, src.GetCountry(), "rule.source.country"),
			"ip":                  convert.ParseList(ctx, types.StringType, src.GetIP(), "rule.source.ip"),
			"host":                oldidname.ParseNameIDList(ctx, src.GetHost(), "rule.source.host"),
			"site":                oldidname.ParseNameIDList(ctx, src.GetSite(), "rule.source.site"),
			"subnet":              convert.ParseList(ctx, types.StringType, src.GetSubnet(), "rule.source.subnet"),
			"ip_range":            convert.ParseFromToList(ctx, src.GetIPRange(), "rule.source.ip_range"),
			"global_ip_range":     oldidname.ParseNameIDList(ctx, src.GetGlobalIPRange(), "rule.source.global_ip_range"),
			"network_interface":   oldidname.ParseNameIDList(ctx, src.GetNetworkInterface(), "rule.source.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, src.GetSiteNetworkSubnet(), "rule.source.site_network_subnet"),
			"floating_subnet":     oldidname.ParseNameIDList(ctx, src.GetFloatingSubnet(), "rule.source.floating_subnet"),
			"user":                oldidname.ParseNameIDList(ctx, src.GetUser(), "rule.source.user"),
			"users_group":         oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "rule.source.users_group"),
			"group":               oldidname.ParseNameIDList(ctx, src.GetGroup(), "rule.source.group"),
			"system_group":        oldidname.ParseNameIDList(ctx, src.GetSystemGroup(), "rule.source.system_group"),
		}
		o, d := types.ObjectValue(applicationcontrol.ApplicationControlSourceAttrTypes, srcAttrs)
		diags.Append(d...)
		out.Source = o
	}

	return out, diags
}
