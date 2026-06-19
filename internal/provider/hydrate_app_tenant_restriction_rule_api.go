package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

//nolint:gocyclo
func hydrateAppTenantRestrictionAddRuleInput(
	ctx context.Context,
	plan AppTenantRestrictionRule,
) (cato_models.AppTenantRestrictionAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.AppTenantRestrictionAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := PolicyRulePositionInput{}
		diags.Append(plan.At.As(ctx, &pos, basetypes.ObjectAsOptions{})...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := AppTenantRestrictionRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
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
		diags.Append(rule.Application.As(ctx, &app, basetypes.ObjectAsOptions{})...)
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
		objs := make([]types.Object, 0, len(rule.Headers.Elements()))
		diags.Append(rule.Headers.ElementsAs(ctx, &objs, false)...)
		for _, o := range objs {
			var h AppTenantRestrictionHeaderPlan
			diags.Append(o.As(ctx, &h, basetypes.ObjectAsOptions{})...)
			data.Headers = append(data.Headers, &cato_models.AppTenantRestrictionHeaderValueInput{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
	}

	if !rule.Schedule.IsNull() && !rule.Schedule.IsUnknown() {
		sch := PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(rule.Schedule.As(ctx, &sch, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			data.Schedule = &cato_models.PolicyScheduleInput{
				ActiveOn: cato_models.PolicyActiveOnEnum(sch.ActiveOn.ValueString()),
			}
			if !sch.CustomTimeframe.IsNull() {
				ctf := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
				diags.Append(sch.CustomTimeframe.As(ctx, &ctf, basetypes.ObjectAsOptions{})...)
				data.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{
					From: ctf.From.ValueString(),
					To:   ctf.To.ValueString(),
				}
			}
			if !sch.CustomRecurring.IsNull() {
				cr := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
				diags.Append(sch.CustomRecurring.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
				data.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{
					From: cato_scalars.Time(cr.From.ValueString()),
					To:   cato_scalars.Time(cr.To.ValueString()),
				}
				diags.Append(cr.Days.ElementsAs(ctx, &data.Schedule.CustomRecurring.Days, false)...)
			}
		}
	}

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
	rule := AppTenantRestrictionRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
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
		diags.Append(rule.Application.As(ctx, &app, basetypes.ObjectAsOptions{})...)
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
		objs := make([]types.Object, 0, len(rule.Headers.Elements()))
		diags.Append(rule.Headers.ElementsAs(ctx, &objs, false)...)
		for _, o := range objs {
			var h AppTenantRestrictionHeaderPlan
			diags.Append(o.As(ctx, &h, basetypes.ObjectAsOptions{})...)
			upd.Rule.Headers = append(upd.Rule.Headers, &cato_models.AppTenantRestrictionHeaderValueInput{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
	}

	if !rule.Schedule.IsNull() && !rule.Schedule.IsUnknown() {
		sch := PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(rule.Schedule.As(ctx, &sch, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			upd.Rule.Schedule = &cato_models.PolicyScheduleUpdateInput{
				ActiveOn: (*cato_models.PolicyActiveOnEnum)(sch.ActiveOn.ValueStringPointer()),
			}
			if !sch.CustomTimeframe.IsNull() {
				ctf := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
				diags.Append(sch.CustomTimeframe.As(ctx, &ctf, basetypes.ObjectAsOptions{})...)
				upd.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{
					From: ctf.From.ValueStringPointer(),
					To:   ctf.To.ValueStringPointer(),
				}
			} else {
				upd.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}
			}
			if !sch.CustomRecurring.IsNull() {
				cr := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
				diags.Append(sch.CustomRecurring.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
				upd.Rule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{
					From: (*cato_scalars.Time)(cr.From.ValueStringPointer()),
					To:   (*cato_scalars.Time)(cr.To.ValueStringPointer()),
				}
				diags.Append(cr.Days.ElementsAs(ctx, &upd.Rule.Schedule.CustomRecurring.Days, false)...)
			} else {
				upd.Rule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}
			}
		}
	}

	if !rule.Source.IsNull() && !rule.Source.IsUnknown() {
		_, acUpd, sdiags := applicationControlSourcePairFromTerraformObject(ctx, rule.Source)
		diags.Append(sdiags...)
		if !diags.HasError() && acUpd != nil {
			upd.Rule.Source = cloneAppTenantRestrictionSourceUpdate(acUpd)
		}
	}

	return upd, diags
}
