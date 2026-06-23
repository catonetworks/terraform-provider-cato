package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
		pos := PolicyRulePositionInput{}
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
		app := PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost{}
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

	data.Schedule = &cato_models.PolicyScheduleInput{ActiveOn: cato_models.PolicyActiveOnEnum(policyActiveOnAlways)}
	if !rule.Schedule.IsNull() && !rule.Schedule.IsUnknown() {
		scheduleAsOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		sch := PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(rule.Schedule.As(ctx, &sch, scheduleAsOpts)...)
		if !diags.HasError() {
			if !sch.ActiveOn.IsNull() && !sch.ActiveOn.IsUnknown() && sch.ActiveOn.ValueString() != "" {
				data.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(sch.ActiveOn.ValueString())
			}
			active := data.Schedule.ActiveOn
			if active == cato_models.PolicyActiveOnEnumCustomTimeframe &&
				!sch.CustomTimeframe.IsNull() && !sch.CustomTimeframe.IsUnknown() {
				ctf := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
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
				cr := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
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
		acAdd, _, sdiags := applicationControlSourcePairFromTerraformObject(ctx, rule.Source)
		diags.Append(sdiags...)
		if !diags.HasError() && acAdd != nil {
			data.Source = cloneAppTenantRestrictionSource(acAdd)
		}
	}

	out.Rule = data
	return out, diags
}

//nolint:gocyclo
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
		app := PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost{}
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

	defaultActive := cato_models.PolicyActiveOnEnum(policyActiveOnAlways)
	upd.Rule.Schedule = &cato_models.PolicyScheduleUpdateInput{ActiveOn: &defaultActive}
	if !rule.Schedule.IsNull() && !rule.Schedule.IsUnknown() {
		scheduleAsOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		sch := PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(rule.Schedule.As(ctx, &sch, scheduleAsOpts)...)
		if !diags.HasError() {
			if !sch.ActiveOn.IsNull() && !sch.ActiveOn.IsUnknown() && sch.ActiveOn.ValueString() != "" {
				upd.Rule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(sch.ActiveOn.ValueStringPointer())
			}
			active := upd.Rule.Schedule.ActiveOn
			if active != nil && *active == cato_models.PolicyActiveOnEnumCustomTimeframe &&
				!sch.CustomTimeframe.IsNull() && !sch.CustomTimeframe.IsUnknown() {
				ctf := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
				diags.Append(sch.CustomTimeframe.As(ctx, &ctf, scheduleAsOpts)...)
				from := ctf.From.ValueStringPointer()
				to := ctf.To.ValueStringPointer()
				if (from != nil && *from != "") || (to != nil && *to != "") {
					upd.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{
						From: from,
						To:   to,
					}
				}
			}
			if active != nil && *active == cato_models.PolicyActiveOnEnumCustomRecurring &&
				!sch.CustomRecurring.IsNull() && !sch.CustomRecurring.IsUnknown() {
				cr := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
				diags.Append(sch.CustomRecurring.As(ctx, &cr, scheduleAsOpts)...)
				from := cr.From.ValueString()
				to := cr.To.ValueString()
				if from != "" && to != "" {
					upd.Rule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{
						From: (*cato_scalars.Time)(cr.From.ValueStringPointer()),
						To:   (*cato_scalars.Time)(cr.To.ValueStringPointer()),
					}
					if !cr.Days.IsNull() && !cr.Days.IsUnknown() {
						diags.Append(cr.Days.ElementsAs(ctx, &upd.Rule.Schedule.CustomRecurring.Days, false)...)
					}
				}
			}
		}
	}

	upd.Rule.Source = &cato_models.AppTenantRestrictionSourceUpdateInput{}
	if !rule.Source.IsNull() && !rule.Source.IsUnknown() {
		_, acUpd, sdiags := applicationControlSourcePairFromTerraformObject(ctx, rule.Source)
		diags.Append(sdiags...)
		if !diags.HasError() && acUpd != nil {
			upd.Rule.Source = cloneAppTenantRestrictionSourceUpdate(acUpd)
		}
	}

	return upd, diags
}
