package acrule

import (
	"context"
	"fmt"

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
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func acEmptyApplicationControlActionConfigInput() *cato_models.ApplicationControlActionConfigInput {
	return &cato_models.ApplicationControlActionConfigInput{
		UserNotification: []*cato_models.UserNotificationTemplateRefInput{},
	}
}

// acEmptyApplicationContextInput returns a non-nil context object required by the
// Application Control GraphQL mutations (applicationContext cannot be null).
func acEmptyApplicationContextInput() *cato_models.ApplicationControlContextInput {
	return &cato_models.ApplicationControlContextInput{
		ApplicationTenant: []*cato_models.ApplicationControlTenantInput{},
	}
}

// acEmptyApplicationCriteriaInput returns a non-nil criteria object required by
// Application Control GraphQL for rule types that support application criteria.
func acEmptyApplicationCriteriaInput() *cato_models.ApplicationControlCriteriaInput {
	anyV := cato_models.ApplicationControlAttributeValueAny
	return &cato_models.ApplicationControlCriteriaInput{
		Attributes: &cato_models.ApplicationControlAttributesInput{
			ComplianceAttributes: &cato_models.ApplicationControlComplianceAttributesInput{
				Hippa: anyV, Isae3402: anyV, Iso27001: anyV, PciDss: anyV,
				Soc1: anyV, Soc2: anyV, Soc3: anyV, Sox: anyV,
			},
			SecurityAttributes: &cato_models.ApplicationControlSecurityAttributesInput{
				AuditTrail: anyV, EncryptionAtRest: anyV, HTTPSecurityHeaders: anyV,
				Mfa: anyV, Rbac: anyV, RememberPassword: anyV, Sso: anyV,
				TLSEnforcement: anyV, TrustedCertificate: anyV,
			},
		},
		OriginCountry: []*cato_models.CountryRefInput{},
		Risk:          []*cato_models.ApplicationControlRiskCriteriaInput{},
	}
}

func hydrateApplicationControlAddRuleInput(
	ctx context.Context,
	plan ApplicationControlRule,
) (cato_models.ApplicationControlAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.ApplicationControlAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := ApplicationControlRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.ApplicationControlAddRuleDataInput{
		Name:        rule.Name.ValueString(),
		Description: rule.Description.ValueString(),
		Enabled:     rule.Enabled.ValueBool(),
		RuleType:    cato_models.ApplicationControlRuleType(rule.RuleType.ValueString()),
	}

	switch data.RuleType {
	case cato_models.ApplicationControlRuleTypeApplication:
		if !rule.ApplicationRule.IsNull() {
			ar, d := hydrateACApplicationRuleAdd(ctx, rule.ApplicationRule)
			diags.Append(d...)
			data.ApplicationRule = ar
		}
	case cato_models.ApplicationControlRuleTypeData:
		if !rule.DataRule.IsNull() {
			dr, d := hydrateACDataRuleAdd(ctx, rule.DataRule)
			diags.Append(d...)
			data.DataRule = dr
		}
	case cato_models.ApplicationControlRuleTypeFile:
		if !rule.FileRule.IsNull() {
			fr, d := hydrateACFileRuleAdd(ctx, rule.FileRule)
			diags.Append(d...)
			data.FileRule = fr
		}
	default:
		diags.AddError("Invalid rule_type", string(data.RuleType))
	}

	out.Rule = data
	return out, diags
}

func hydrateApplicationControlUpdateRuleInput(
	ctx context.Context,
	plan ApplicationControlRule,
) (cato_models.ApplicationControlUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	rule := ApplicationControlRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return cato_models.ApplicationControlUpdateRuleInput{}, diags
	}
	rt := cato_models.ApplicationControlRuleType(rule.RuleType.ValueString())
	upd := cato_models.ApplicationControlUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.ApplicationControlUpdateRuleDataInput{
			Name:        rule.Name.ValueStringPointer(),
			Description: rule.Description.ValueStringPointer(),
			Enabled:     rule.Enabled.ValueBoolPointer(),
			RuleType:    &rt,
		},
	}
	switch rt {
	case cato_models.ApplicationControlRuleTypeApplication:
		if !rule.ApplicationRule.IsNull() {
			ar, d := hydrateACApplicationRuleUpdate(ctx, rule.ApplicationRule)
			diags.Append(d...)
			upd.Rule.ApplicationRule = ar
		}
	case cato_models.ApplicationControlRuleTypeData:
		if !rule.DataRule.IsNull() {
			dr, d := hydrateACDataRuleUpdate(ctx, rule.DataRule)
			diags.Append(d...)
			upd.Rule.DataRule = dr
		}
	case cato_models.ApplicationControlRuleTypeFile:
		if !rule.FileRule.IsNull() {
			fr, d := hydrateACFileRuleUpdate(ctx, rule.FileRule)
			diags.Append(d...)
			upd.Rule.FileRule = fr
		}
	}
	return upd, diags
}

func hydrateACApplicationRuleAdd(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlApplicationRuleInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlApplicationRuleInput{
		Action:                     cato_models.ApplicationControlAction(p.Action.ValueString()),
		Severity:                   cato_models.ApplicationControlSeverity(p.Severity.ValueString()),
		ApplicationActivitySatisfy: cato_models.ApplicationControlSatisfyAll,
		ApplicationCriteriaSatisfy: cato_models.ApplicationControlSatisfyAll,
		ApplicationContext:         acEmptyApplicationContextInput(),
		ApplicationCriteria:        acEmptyApplicationCriteriaInput(),
		ApplicationActivity:        []*cato_models.ApplicationControlActivityInput{},
	}
	satisfy := p.ApplicationActivitySatisfy
	if !satisfy.IsNull() && !satisfy.IsUnknown() && satisfy.ValueString() != "" {
		out.ApplicationActivitySatisfy = cato_models.ApplicationControlSatisfy(p.ApplicationActivitySatisfy.ValueString())
	}
	out.ApplicationActivity, diags = acApplicationActivityFromPlan(ctx, p, diags)
	out.ApplicationCriteria, diags = acCriteriaInputFromPlan(ctx, p.ApplicationCriteria, "application_rule.application_criteria", diags)
	if cs := p.ApplicationCriteriaSatisfy; !cs.IsNull() && !cs.IsUnknown() && cs.ValueString() != "" {
		out.ApplicationCriteriaSatisfy = cato_models.ApplicationControlSatisfy(cs.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if out.ActionConfig == nil {
		out.ActionConfig = acEmptyApplicationControlActionConfigInput()
	}
	acRequireApplicationSelection(out.Application, "application_rule", &diags)
	return out, diags
}

func hydrateACApplicationRuleUpdate(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlApplicationRuleUpdateInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlApplicationRuleUpdateInput{}
	if !p.Action.IsNull() {
		v := cato_models.ApplicationControlAction(p.Action.ValueString())
		out.Action = &v
	}
	if !p.Severity.IsNull() {
		v := cato_models.ApplicationControlSeverity(p.Severity.ValueString())
		out.Severity = &v
	}
	crit, diags := acCriteriaInputFromPlan(ctx, p.ApplicationCriteria, "application_rule.application_criteria", diags)
	out.ApplicationCriteria = acCriteriaUpdateFromInput(crit)
	if cs := p.ApplicationCriteriaSatisfy; !cs.IsNull() && !cs.IsUnknown() && cs.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(cs.ValueString())
		out.ApplicationCriteriaSatisfy = &v
	}
	// activities must be sent on update too — omitting them leaves the live
	// rows (incl. granular conditions) unchanged and the refreshed state then
	// mismatches the plan ("inconsistent result after apply").
	out.ApplicationActivity, diags = acApplicationActivityFromPlan(ctx, p, diags)
	if s := p.ApplicationActivitySatisfy; !s.IsNull() && !s.IsUnknown() && s.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(s.ValueString())
		out.ApplicationActivitySatisfy = &v
	}
	diags.Append(acFillCommonUpdate(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	return out, diags
}

func hydrateACDataRuleAdd(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlDataRuleInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlDataRuleInput{
		Action:                     cato_models.ApplicationControlAction(p.Action.ValueString()),
		Severity:                   cato_models.ApplicationControlSeverity(p.Severity.ValueString()),
		ApplicationActivitySatisfy: cato_models.ApplicationControlSatisfyAll,
		ApplicationCriteriaSatisfy: cato_models.ApplicationControlSatisfyAll,
		FileAttributeSatisfy:       cato_models.ApplicationControlSatisfyAll,
		ApplicationContext:         acEmptyApplicationContextInput(),
		ApplicationCriteria:        acEmptyApplicationCriteriaInput(),
		ApplicationActivity:        []*cato_models.ApplicationControlActivityInput{},
	}
	satisfy := p.ApplicationActivitySatisfy
	if !satisfy.IsNull() && !satisfy.IsUnknown() && satisfy.ValueString() != "" {
		out.ApplicationActivitySatisfy = cato_models.ApplicationControlSatisfy(p.ApplicationActivitySatisfy.ValueString())
	}
	out.ApplicationActivity, diags = acApplicationActivityFromPlan(ctx, p, diags)
	out.ApplicationCriteria, diags = acCriteriaInputFromPlan(ctx, p.ApplicationCriteria, "data_rule.application_criteria", diags)
	if cs := p.ApplicationCriteriaSatisfy; !cs.IsNull() && !cs.IsUnknown() && cs.ValueString() != "" {
		out.ApplicationCriteriaSatisfy = cato_models.ApplicationControlSatisfy(cs.ValueString())
	}
	if !p.FileAttributeSatisfy.IsNull() && !p.FileAttributeSatisfy.IsUnknown() && p.FileAttributeSatisfy.ValueString() != "" {
		out.FileAttributeSatisfy = cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	dlp, dlpDiags := acDlpProfileAddOrEmpty(ctx, p.DlpProfile, diags)
	diags.Append(dlpDiags...)
	out.DlpProfile = dlp

	if !p.FileAttribute.IsNull() && !p.FileAttribute.IsUnknown() {
		out.FileAttribute, diags = acFileAttributesAdd(ctx, p.FileAttribute, diags)
	}
	if out.Application == nil {
		// non-nullable input; an absent block means "any app" for data/file
		// rules (selection comes from activities/file attributes/dlp).
		out.Application = &cato_models.ApplicationControlApplicationInput{}
	}
	return out, diags
}

//nolint:gocyclo
func hydrateACDataRuleUpdate(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlDataRuleUpdateInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlDataRuleUpdateInput{}
	if !p.Action.IsNull() {
		v := cato_models.ApplicationControlAction(p.Action.ValueString())
		out.Action = &v
	}
	if !p.Severity.IsNull() {
		v := cato_models.ApplicationControlSeverity(p.Severity.ValueString())
		out.Severity = &v
	}
	if !p.FileAttributeSatisfy.IsNull() && !p.FileAttributeSatisfy.IsUnknown() && p.FileAttributeSatisfy.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
		out.FileAttributeSatisfy = &v
	}
	crit, diags := acCriteriaInputFromPlan(ctx, p.ApplicationCriteria, "data_rule.application_criteria", diags)
	out.ApplicationCriteria = acCriteriaUpdateFromInput(crit)
	if cs := p.ApplicationCriteriaSatisfy; !cs.IsNull() && !cs.IsUnknown() && cs.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(cs.ValueString())
		out.ApplicationCriteriaSatisfy = &v
	}
	out.ApplicationActivity, diags = acApplicationActivityFromPlan(ctx, p, diags)
	if s := p.ApplicationActivitySatisfy; !s.IsNull() && !s.IsUnknown() && s.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(s.ValueString())
		out.ApplicationActivitySatisfy = &v
	}
	diags.Append(acFillCommonUpdate(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.DlpProfile.IsNull() && !p.DlpProfile.IsUnknown() {
		dlp, dlpDiags := acDlpProfileUpdate(ctx, p.DlpProfile, diags)
		diags.Append(dlpDiags...)
		out.DlpProfile = dlp
	}
	if !p.FileAttribute.IsNull() && !p.FileAttribute.IsUnknown() {
		out.FileAttribute, diags = acFileAttributesUpdate(ctx, p.FileAttribute, diags)
	}
	return out, diags
}

func hydrateACFileRuleAdd(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlFileRuleInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlFileRuleInput{
		Action:                     cato_models.ApplicationControlAction(p.Action.ValueString()),
		Severity:                   cato_models.ApplicationControlSeverity(p.Severity.ValueString()),
		ApplicationActivitySatisfy: cato_models.ApplicationControlSatisfyAll,
		ApplicationCriteriaSatisfy: cato_models.ApplicationControlSatisfyAll,
		FileAttributeSatisfy:       cato_models.ApplicationControlSatisfyAll,
		ApplicationContext:         acEmptyApplicationContextInput(),
		ApplicationCriteria:        acEmptyApplicationCriteriaInput(),
		ApplicationActivity:        []*cato_models.ApplicationControlActivityInput{},
	}
	satisfy := p.ApplicationActivitySatisfy
	if !satisfy.IsNull() && !satisfy.IsUnknown() && satisfy.ValueString() != "" {
		out.ApplicationActivitySatisfy = cato_models.ApplicationControlSatisfy(p.ApplicationActivitySatisfy.ValueString())
	}
	out.ApplicationActivity, diags = acApplicationActivityFromPlan(ctx, p, diags)
	out.ApplicationCriteria, diags = acCriteriaInputFromPlan(ctx, p.ApplicationCriteria, "file_rule.application_criteria", diags)
	if cs := p.ApplicationCriteriaSatisfy; !cs.IsNull() && !cs.IsUnknown() && cs.ValueString() != "" {
		out.ApplicationCriteriaSatisfy = cato_models.ApplicationControlSatisfy(cs.ValueString())
	}
	if !p.FileAttributeSatisfy.IsNull() && !p.FileAttributeSatisfy.IsUnknown() && p.FileAttributeSatisfy.ValueString() != "" {
		out.FileAttributeSatisfy = cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.FileAttribute.IsNull() && !p.FileAttribute.IsUnknown() {
		out.FileAttribute, diags = acFileAttributesAdd(ctx, p.FileAttribute, diags)
	}
	if out.Application == nil {
		out.Application = &cato_models.ApplicationControlApplicationInput{}
	}
	return out, diags
}

func hydrateACFileRuleUpdate(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlFileRuleUpdateInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlFileRuleUpdateInput{}
	if !p.Action.IsNull() {
		v := cato_models.ApplicationControlAction(p.Action.ValueString())
		out.Action = &v
	}
	if !p.Severity.IsNull() {
		v := cato_models.ApplicationControlSeverity(p.Severity.ValueString())
		out.Severity = &v
	}
	if !p.FileAttributeSatisfy.IsNull() && !p.FileAttributeSatisfy.IsUnknown() && p.FileAttributeSatisfy.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
		out.FileAttributeSatisfy = &v
	}
	crit, diags := acCriteriaInputFromPlan(ctx, p.ApplicationCriteria, "file_rule.application_criteria", diags)
	out.ApplicationCriteria = acCriteriaUpdateFromInput(crit)
	if cs := p.ApplicationCriteriaSatisfy; !cs.IsNull() && !cs.IsUnknown() && cs.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(cs.ValueString())
		out.ApplicationCriteriaSatisfy = &v
	}
	out.ApplicationActivity, diags = acApplicationActivityFromPlan(ctx, p, diags)
	if s := p.ApplicationActivitySatisfy; !s.IsNull() && !s.IsUnknown() && s.ValueString() != "" {
		v := cato_models.ApplicationControlSatisfy(s.ValueString())
		out.ApplicationActivitySatisfy = &v
	}
	diags.Append(acFillCommonUpdate(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.FileAttribute.IsNull() {
		out.FileAttribute, diags = acFileAttributesUpdate(ctx, p.FileAttribute, diags)
	}
	return out, diags
}

func acDlpProfileAdd(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlDlpProfileInput, diag.Diagnostics) {
	var p struct {
		ContentProfile types.Set `tfsdk:"content_profile"`
		EdmProfile     types.Set `tfsdk:"edm_profile"`
	}
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	out := &cato_models.ApplicationControlDlpProfileInput{
		ContentProfile: []*cato_models.DlpContentProfileRefInput{},
		EdmProfile:     []*cato_models.DlpEdmProfileRefInput{},
	}
	if !p.ContentProfile.IsNull() && !p.ContentProfile.IsUnknown() {
		objs := make([]types.Object, 0, len(p.ContentProfile.Elements()))
		diags.Append(p.ContentProfile.ElementsAs(ctx, &objs, false)...)
		var item ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		for _, el := range objs {
			diags.Append(el.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError("dlp_profile.content_profile", err.Error())
				return out, diags
			}
			out.ContentProfile = append(out.ContentProfile, &cato_models.DlpContentProfileRefInput{
				By: cato_models.ObjectRefBy(ref.By), Input: ref.Input,
			})
		}
	}
	if !p.EdmProfile.IsNull() && !p.EdmProfile.IsUnknown() {
		objs := make([]types.Object, 0, len(p.EdmProfile.Elements()))
		diags.Append(p.EdmProfile.ElementsAs(ctx, &objs, false)...)
		var item ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		for _, el := range objs {
			diags.Append(el.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError("dlp_profile.edm_profile", err.Error())
				return out, diags
			}
			out.EdmProfile = append(out.EdmProfile, &cato_models.DlpEdmProfileRefInput{
				By: cato_models.ObjectRefBy(ref.By), Input: ref.Input,
			})
		}
	}
	return out, diags
}

func acDlpProfileUpdate(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlDlpProfileUpdateInput, diag.Diagnostics) {
	add, diags := acDlpProfileAdd(ctx, o, diags)
	return &cato_models.ApplicationControlDlpProfileUpdateInput{
		ContentProfile: add.ContentProfile,
		EdmProfile:     add.EdmProfile,
	}, diags
}

//nolint:gocyclo
func acApplicationActivityFromPlan(
	ctx context.Context,
	p ApplicationControlTypedRulePlan,
	diags diag.Diagnostics,
) ([]*cato_models.ApplicationControlActivityInput, diag.Diagnostics) {
	out := []*cato_models.ApplicationControlActivityInput{}
	if p.ApplicationActivity.IsNull() || p.ApplicationActivity.IsUnknown() {
		return out, diags
	}
	objs := make([]types.Object, 0, len(p.ApplicationActivity.Elements()))
	diags.Append(p.ApplicationActivity.ElementsAs(ctx, &objs, false)...)
	for _, o := range objs {
		var row ApplicationControlActivityPlan
		diags.Append(o.As(ctx, &row, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
		var act ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		diags.Append(row.Activity.As(ctx, &act, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
		ref, err := utils.TransformObjectRefInput(act)
		if err != nil {
			diags.AddError("application_activity", err.Error())
			return out, diags
		}
		in := &cato_models.ApplicationControlActivityInput{
			Activity: &cato_models.ApplicationControlActivityRefInput{
				By:    cato_models.ObjectRefBy(ref.By),
				Input: ref.Input,
			},
		}
		// Granular activity conditions. Some catalog activities (e.g.
		// full_path_url) REQUIRE a value — the API rejects a bare ref with
		// "activity: <name> must have a value".
		if !row.Field.IsNull() && !row.Field.IsUnknown() {
			var fld ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
			diags.Append(row.Field.As(ctx, &fld, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
			fref, ferr := utils.TransformObjectRefInput(fld)
			if ferr != nil {
				diags.AddError("application_activity.field", ferr.Error())
				return out, diags
			}
			in.Field = &cato_models.ApplicationControlActivityFieldRefInput{
				By:    cato_models.ObjectRefBy(fref.By),
				Input: fref.Input,
			}
		}
		if !row.Operator.IsNull() && !row.Operator.IsUnknown() && row.Operator.ValueString() != "" {
			op := cato_models.ApplicationControlOperator(row.Operator.ValueString())
			in.Operator = &op
		}
		if !row.Value.IsNull() && !row.Value.IsUnknown() && row.Value.ValueString() != "" {
			in.Value = row.Value.ValueStringPointer()
		}
		if !row.ValueSet.IsNull() && !row.ValueSet.IsUnknown() {
			var vs ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
			diags.Append(row.ValueSet.As(ctx, &vs, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
			vref, verr := utils.TransformObjectRefInput(vs)
			if verr != nil {
				diags.AddError("application_activity.value_set", verr.Error())
				return out, diags
			}
			in.ValueSet = &cato_models.StringValueSetRefInput{
				By:    cato_models.ObjectRefBy(vref.By),
				Input: vref.Input,
			}
		}
		out = append(out, in)
	}
	return out, diags
}

func acFileAttributesAdd(
	ctx context.Context,
	list types.List,
	diags diag.Diagnostics,
) ([]*cato_models.ApplicationControlFileAttributeInput, diag.Diagnostics) {
	out := []*cato_models.ApplicationControlFileAttributeInput{}
	if list.IsNull() || list.IsUnknown() {
		return out, diags
	}
	objs := make([]types.Object, 0, len(list.Elements()))
	diags.Append(list.ElementsAs(ctx, &objs, false)...)
	var row struct {
		FileAttribute          types.String `tfsdk:"file_attribute"`
		Operator               types.String `tfsdk:"operator"`
		Value                  types.String `tfsdk:"value"`
		ContentTypeValues      types.Set    `tfsdk:"content_type_values"`
		ContentTypeGroupValues types.Set    `tfsdk:"content_type_group_values"`
	}
	refList := func(set types.Set, attrName string) []utils.ObjectRefOutput {
		if set.IsNull() || set.IsUnknown() {
			return nil
		}
		els := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &els, false)...)
		refs := make([]utils.ObjectRefOutput, 0, len(els))
		for _, el := range els {
			var item ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
			diags.Append(el.As(ctx, &item, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError(attrName, err.Error())
				continue
			}
			refs = append(refs, ref)
		}
		return refs
	}
	for _, el := range objs {
		diags.Append(el.As(ctx, &row, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
		in := &cato_models.ApplicationControlFileAttributeInput{
			FileAttribute: cato_models.ApplicationControlFileAttributeType(row.FileAttribute.ValueString()),
			Operator:      cato_models.ApplicationControlOperator(row.Operator.ValueString()),
		}
		// CONTENT_TYPE rows select via the content-type ref sets and the API
		// REQUIRES value to stay null for them — only pass value when set.
		if !row.Value.IsNull() && !row.Value.IsUnknown() && row.Value.ValueString() != "" {
			in.Value = row.Value.ValueStringPointer()
		}
		for _, ref := range refList(row.ContentTypeValues, "file_attribute.content_type_values") {
			in.ContentTypeValues = append(in.ContentTypeValues, &cato_models.ApplicationControlContentTypeRefInput{
				By: cato_models.ObjectRefBy(ref.By), Input: ref.Input,
			})
		}
		for _, ref := range refList(row.ContentTypeGroupValues, "file_attribute.content_type_group_values") {
			in.ContentTypeGroupValues = append(in.ContentTypeGroupValues, &cato_models.ApplicationControlContentTypeGroupRefInput{
				By: cato_models.ObjectRefBy(ref.By), Input: ref.Input,
			})
		}
		out = append(out, in)
	}
	return out, diags
}

func acFileAttributesUpdate(
	ctx context.Context,
	list types.List,
	diags diag.Diagnostics,
) ([]*cato_models.ApplicationControlFileAttributeInput, diag.Diagnostics) {
	return acFileAttributesAdd(ctx, list, diags)
}

//nolint:gocyclo
func acFillCommonAdd(
	ctx context.Context,
	p ApplicationControlTypedRulePlan,
	schedule **cato_models.PolicyScheduleInput,
	source **cato_models.ApplicationControlSourceInput,
	tracking **cato_models.PolicyTrackingInput,
	device *[]*cato_models.DeviceProfileRefInput,
	accessMethod *[]*cato_models.ApplicationControlAccessMethodInput,
	application **cato_models.ApplicationControlApplicationInput,
	actionConfig **cato_models.ApplicationControlActionConfigInput,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := &cato_models.PolicyScheduleInput{ActiveOn: cato_models.PolicyActiveOnEnumAlways}
	if !p.Schedule.IsNull() && !p.Schedule.IsUnknown() {
		sch := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		scheduleTopOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		diags.Append(p.Schedule.As(ctx, &sch, scheduleTopOpts)...)
		if !sch.ActiveOn.IsNull() {
			in.ActiveOn = cato_models.PolicyActiveOnEnum(sch.ActiveOn.ValueString())
		}
		if !sch.CustomTimeframe.IsNull() && !sch.CustomTimeframe.IsUnknown() {
			ctf := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
			diags.Append(sch.CustomTimeframe.As(ctx, &ctf, basetypes.ObjectAsOptions{})...)
			in.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{From: ctf.From.ValueString(), To: ctf.To.ValueString()}
		}
		if !sch.CustomRecurring.IsNull() && !sch.CustomRecurring.IsUnknown() {
			cr := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
			diags.Append(sch.CustomRecurring.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
			in.CustomRecurring = &cato_models.PolicyCustomRecurringInput{
				From: cato_scalars.Time(cr.From.ValueString()),
				To:   cato_scalars.Time(cr.To.ValueString()),
			}
			if !cr.Days.IsNull() && !cr.Days.IsUnknown() {
				diags.Append(cr.Days.ElementsAs(ctx, &in.CustomRecurring.Days, false)...)
			}
		}
	}
	*schedule = in
	if !p.Source.IsNull() && !p.Source.IsUnknown() {
		add, _, sdiags := applicationcontrol.ApplicationControlSourcePairFromTerraformObject(ctx, p.Source)
		diags.Append(sdiags...)
		*source = add
	}
	// source is a non-nullable input (ApplicationControlSourceInput!) — an
	// omitted block means "any source" and must be sent as an empty object,
	// never null (same API class as the application-null rejection).
	if *source == nil {
		*source = &cato_models.ApplicationControlSourceInput{}
	}
	if !p.Tracking.IsNull() && !p.Tracking.IsUnknown() {
		*tracking, diags = acTrackingAdd(ctx, p.Tracking, diags)
	}
	if !p.Device.IsNull() && !p.Device.IsUnknown() {
		objs := make([]types.Object, 0, len(p.Device.Elements()))
		diags.Append(p.Device.ElementsAs(ctx, &objs, false)...)
		var item ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError("device", err.Error())
				return diags
			}
			*device = append(*device, &cato_models.DeviceProfileRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input})
		}
	}
	if !p.AccessMethod.IsNull() && !p.AccessMethod.IsUnknown() {
		objs := make([]types.Object, 0, len(p.AccessMethod.Elements()))
		diags.Append(p.AccessMethod.ElementsAs(ctx, &objs, false)...)
		for _, o := range objs {
			var row ApplicationControlAccessMethodPlan
			diags.Append(o.As(ctx, &row, basetypes.ObjectAsOptions{})...)
			*accessMethod = append(*accessMethod, &cato_models.ApplicationControlAccessMethodInput{
				AccessMethod: cato_models.ApplicationControlAccessMethodType(row.AccessMethod.ValueString()),
				Operator:     cato_models.ApplicationControlOperator(row.Operator.ValueString()),
				Value:        row.Value.ValueStringPointer(),
			})
		}
	}
	if !p.Application.IsNull() && !p.Application.IsUnknown() {
		*application, diags = acApplicationInputFromWanShape(ctx, p.Application, diags)
	}
	if !p.ActionConfig.IsNull() && !p.ActionConfig.IsUnknown() {
		*actionConfig, diags = acActionConfigAdd(ctx, p.ActionConfig, diags)
	}
	return diags
}

//nolint:gocyclo
func acFillCommonUpdate(
	ctx context.Context,
	p ApplicationControlTypedRulePlan,
	schedule **cato_models.PolicyScheduleUpdateInput,
	source **cato_models.ApplicationControlSourceUpdateInput,
	tracking **cato_models.PolicyTrackingUpdateInput,
	device *[]*cato_models.DeviceProfileRefInput,
	accessMethod *[]*cato_models.ApplicationControlAccessMethodInput,
	application **cato_models.ApplicationControlApplicationUpdateInput,
	actionConfig **cato_models.ApplicationControlActionConfigUpdateInput,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if !p.Schedule.IsNull() && !p.Schedule.IsUnknown() {
		sch := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		scheduleTopOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		diags.Append(p.Schedule.As(ctx, &sch, scheduleTopOpts)...)
		upd := &cato_models.PolicyScheduleUpdateInput{
			ActiveOn: (*cato_models.PolicyActiveOnEnum)(sch.ActiveOn.ValueStringPointer()),
		}
		if !sch.CustomTimeframe.IsNull() && !sch.CustomTimeframe.IsUnknown() {
			ctf := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
			diags.Append(sch.CustomTimeframe.As(ctx, &ctf, basetypes.ObjectAsOptions{})...)
			upd.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{
				From: ctf.From.ValueStringPointer(), To: ctf.To.ValueStringPointer(),
			}
		} else {
			upd.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}
		}
		if !sch.CustomRecurring.IsNull() && !sch.CustomRecurring.IsUnknown() {
			cr := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
			diags.Append(sch.CustomRecurring.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
			upd.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{
				From: (*cato_scalars.Time)(cr.From.ValueStringPointer()),
				To:   (*cato_scalars.Time)(cr.To.ValueStringPointer()),
			}
			if !cr.Days.IsNull() && !cr.Days.IsUnknown() {
				diags.Append(cr.Days.ElementsAs(ctx, &upd.CustomRecurring.Days, false)...)
			}
		} else {
			upd.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}
		}
		*schedule = upd
	}
	if !p.Source.IsNull() && !p.Source.IsUnknown() {
		_, supd, sdiags := applicationcontrol.ApplicationControlSourcePairFromTerraformObject(ctx, p.Source)
		diags.Append(sdiags...)
		*source = supd
	}
	if !p.Tracking.IsNull() && !p.Tracking.IsUnknown() {
		*tracking, diags = acTrackingUpdate(ctx, p.Tracking, diags)
	}
	if !p.Device.IsNull() && !p.Device.IsUnknown() {
		objs := make([]types.Object, 0, len(p.Device.Elements()))
		diags.Append(p.Device.ElementsAs(ctx, &objs, false)...)
		var item ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError("device", err.Error())
				return diags
			}
			*device = append(*device, &cato_models.DeviceProfileRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input})
		}
	}
	if !p.AccessMethod.IsNull() && !p.AccessMethod.IsUnknown() {
		objs := make([]types.Object, 0, len(p.AccessMethod.Elements()))
		diags.Append(p.AccessMethod.ElementsAs(ctx, &objs, false)...)
		for _, o := range objs {
			var row ApplicationControlAccessMethodPlan
			diags.Append(o.As(ctx, &row, basetypes.ObjectAsOptions{})...)
			*accessMethod = append(*accessMethod, &cato_models.ApplicationControlAccessMethodInput{
				AccessMethod: cato_models.ApplicationControlAccessMethodType(row.AccessMethod.ValueString()),
				Operator:     cato_models.ApplicationControlOperator(row.Operator.ValueString()),
				Value:        row.Value.ValueStringPointer(),
			})
		}
	}
	if !p.Application.IsNull() && !p.Application.IsUnknown() {
		*application, diags = acApplicationUpdateFromWanShape(ctx, p.Application, diags)
	}
	if !p.ActionConfig.IsNull() && !p.ActionConfig.IsUnknown() {
		*actionConfig, diags = acActionConfigUpdate(ctx, p.ActionConfig, diags)
	}
	return diags
}

func acActionConfigAdd(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlActionConfigInput, diag.Diagnostics) {
	var p struct {
		UserNotification types.Set `tfsdk:"user_notification"`
	}
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	out := &cato_models.ApplicationControlActionConfigInput{}
	if !p.UserNotification.IsNull() && !p.UserNotification.IsUnknown() {
		objs := make([]types.Object, 0, len(p.UserNotification.Elements()))
		diags.Append(p.UserNotification.ElementsAs(ctx, &objs, false)...)
		var item ifrule.PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		for _, el := range objs {
			diags.Append(el.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError("action_config.user_notification", err.Error())
				return out, diags
			}
			out.UserNotification = append(out.UserNotification, &cato_models.UserNotificationTemplateRefInput{
				By: cato_models.ObjectRefBy(ref.By), Input: ref.Input,
			})
		}
	}
	return out, diags
}

func acActionConfigUpdate(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlActionConfigUpdateInput, diag.Diagnostics) {
	add, diags := acActionConfigAdd(ctx, o, diags)
	if add == nil {
		return &cato_models.ApplicationControlActionConfigUpdateInput{}, diags
	}
	return &cato_models.ApplicationControlActionConfigUpdateInput{UserNotification: add.UserNotification}, diags
}

//nolint:gocyclo,funlen
func acApplicationInputFromWanShape(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlApplicationInput, diag.Diagnostics) {
	wan := acApplicationPlan{}
	diags.Append(o.As(ctx, &wan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	out := &cato_models.ApplicationControlApplicationInput{}
	out.ApplicationType = acApplicationTypesFromPlan(ctx, wan.ApplicationType, &diags)
	appendFirst := func(set types.Set, setter func(*cato_models.ApplicationRefInput)) {
		if set.IsNull() || set.IsUnknown() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		if len(objs) == 0 {
			return
		}
		var item wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplicationApplication
		diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
		ref, err := utils.TransformObjectRefInput(item)
		if err != nil {
			diags.AddError("application", err.Error())
			return
		}
		setter(&cato_models.ApplicationRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input})
	}
	appendFirstCat := func(set types.Set, setter func(*cato_models.ApplicationCategoryRefInput)) {
		if set.IsNull() || set.IsUnknown() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		if len(objs) == 0 {
			return
		}
		var item wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplicationAppCategory
		diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
		ref, err := utils.TransformObjectRefInput(item)
		if err != nil {
			diags.AddError("application category", err.Error())
			return
		}
		setter(&cato_models.ApplicationCategoryRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input})
	}
	appendFirst(wan.Application, func(v *cato_models.ApplicationRefInput) { out.Application = v })
	if out.Application == nil && !wan.CustomApp.IsNull() && !wan.CustomApp.IsUnknown() {
		objs := make([]types.Object, 0, len(wan.CustomApp.Elements()))
		diags.Append(wan.CustomApp.ElementsAs(ctx, &objs, false)...)
		if len(objs) > 0 {
			var item wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomApp
			diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err == nil {
				out.CustomApp = &cato_models.CustomApplicationRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input}
			}
		}
	}
	appendFirstCat(wan.AppCategory, func(v *cato_models.ApplicationCategoryRefInput) { out.AppCategory = v })
	if !wan.CustomCategory.IsNull() && !wan.CustomCategory.IsUnknown() {
		objs := make([]types.Object, 0, len(wan.CustomCategory.Elements()))
		diags.Append(wan.CustomCategory.ElementsAs(ctx, &objs, false)...)
		if len(objs) > 0 {
			var item wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomCategory
			diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err == nil {
				out.CustomCategory = &cato_models.CustomCategoryRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input}
			}
		}
	}
	if !wan.SanctionedAppsCategory.IsNull() && !wan.SanctionedAppsCategory.IsUnknown() {
		objs := make([]types.Object, 0, len(wan.SanctionedAppsCategory.Elements()))
		diags.Append(wan.SanctionedAppsCategory.ElementsAs(ctx, &objs, false)...)
		if len(objs) > 0 {
			var item wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplicationSanctionedAppsCategory
			diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err == nil {
				out.SanctionedAppsCategory = &cato_models.SanctionedAppsCategoryRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input}
			}
		}
	}
	return out, diags
}

func acApplicationUpdateFromWanShape(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlApplicationUpdateInput, diag.Diagnostics) {
	in, diags := acApplicationInputFromWanShape(ctx, o, diags)
	if in == nil {
		return &cato_models.ApplicationControlApplicationUpdateInput{}, diags
	}
	return &cato_models.ApplicationControlApplicationUpdateInput{
		Application:            in.Application,
		CustomApp:              in.CustomApp,
		AppCategory:            in.AppCategory,
		CustomCategory:         in.CustomCategory,
		ApplicationType:        in.ApplicationType,
		SanctionedAppsCategory: in.SanctionedAppsCategory,
	}, diags
}

func acTrackingAdd(
	ctx context.Context,
	o types.Object,
	diags diag.Diagnostics,
) (*cato_models.PolicyTrackingInput, diag.Diagnostics) {
	t := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleTracking{}
	trackOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	diags.Append(o.As(ctx, &t, trackOpts)...)
	out := &cato_models.PolicyTrackingInput{
		Event: &cato_models.PolicyRuleTrackingEventInput{},
		Alert: &cato_models.PolicyRuleTrackingAlertInput{Enabled: false, Frequency: "DAILY"},
	}
	if !t.Event.IsNull() && !t.Event.IsUnknown() {
		ev := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleTrackingEvent{}
		diags.Append(t.Event.As(ctx, &ev, trackOpts)...)
		out.Event.Enabled = ev.Enabled.ValueBool()
	}
	if !t.Alert.IsNull() && !t.Alert.IsUnknown() {
		al := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlert{}
		diags.Append(t.Alert.As(ctx, &al, trackOpts)...)
		out.Alert.Enabled = al.Enabled.ValueBool()
		out.Alert.Frequency = cato_models.PolicyRuleTrackingFrequencyEnum(al.Frequency.ValueString())
		// subscription groups etc. omitted for brevity — extend as needed
	}
	return out, diags
}

func acTrackingUpdate(
	ctx context.Context, o types.Object, diags diag.Diagnostics,
) (*cato_models.PolicyTrackingUpdateInput, diag.Diagnostics) {
	in, diags := acTrackingAdd(ctx, o, diags)
	if in == nil {
		return &cato_models.PolicyTrackingUpdateInput{}, diags
	}
	upd := &cato_models.PolicyTrackingUpdateInput{
		Event: &cato_models.PolicyRuleTrackingEventUpdateInput{Enabled: &in.Event.Enabled},
		Alert: &cato_models.PolicyRuleTrackingAlertUpdateInput{
			Enabled:   &in.Alert.Enabled,
			Frequency: &in.Alert.Frequency,
		},
	}
	return upd, diags
}

func applicationControlTypedRuleStateFromApplicationRule(
	ctx context.Context,
	ar *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ar == nil {
		return types.ObjectNull(applicationControlTypedRuleAttrTypes), diags
	}

	sch := ar.GetSchedule()
	schObj := acScheduleObjectFromApplicationRuleSchedule(ctx, sch, &diags)

	src := ar.GetSource()
	srcObj := acSourceObjectFromApplicationRuleSource(ctx, src, &diags)

	tr := ar.GetTracking()
	trObj := acTrackingObjectFromApplicationRuleTracking(ctx, tr, &diags)

	appObj := wanApplicationObjectFromApplicationRuleApplication(ctx, ar.GetApplication(), &diags)

	acList := acAccessMethodListFromApplicationRule(ctx, ar.GetAccessMethod(), &diags)
	activityList := acApplicationActivityListFromApplicationRule(ar.GetApplicationActivity(), &diags)
	actSatisfy := types.StringValue("ALL")
	if s := ar.GetApplicationActivitySatisfy(); s != nil {
		actSatisfy = types.StringValue(string(*s))
	}

	actionCfg := types.ObjectNull(applicationControlActionConfigAttrTypes)
	if cfg := ar.GetActionConfig(); cfg != nil {
		un := idname.ParseIDNameSet(ctx, cfg.GetUserNotification(), &diags)
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	attrs := map[string]attr.Value{
		"action":                       types.StringValue(ar.GetAction().String()),
		"severity":                     types.StringValue(ar.GetSeverity().String()),
		"schedule":                     schObj,
		"source":                       srcObj,
		"tracking":                     trObj,
		"device":                       idname.ParseIDNameSet(ctx, ar.GetDevice(), &diags),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
		"application_criteria":         acCriteriaObjectFromApplicationRule(ar.GetApplicationCriteria(), &diags),
		"application_criteria_satisfy": acSatisfyStateValue(ar.ApplicationCriteriaSatisfy),
		"action_config":                actionCfg,
		"file_attribute":               types.ListNull(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}),
		"file_attribute_satisfy":       types.StringNull(),
		"dlp_profile":                  types.ObjectNull(applicationControlDlpProfileAttrTypes),
	}
	o, d := types.ObjectValue(applicationControlTypedRuleAttrTypes, attrs)
	diags.Append(d...)
	return o, diags
}

func applicationControlTypedRuleStateFromDataRule(
	ctx context.Context,
	dr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if dr == nil {
		return types.ObjectNull(applicationControlTypedRuleAttrTypes), diags
	}

	schObj := acScheduleObjectFromDataRuleSchedule(ctx, dr.GetSchedule(), &diags)
	srcObj := acSourceObjectFromDataRuleSource(ctx, dr.GetSource(), &diags)
	trObj := acTrackingObjectFromDataRuleTracking(ctx, dr.GetTracking(), &diags)
	appObj := wanApplicationObjectFromDataRuleApplication(ctx, dr.GetApplication(), &diags)
	acList := acAccessMethodListFromDataRule(ctx, dr.GetAccessMethod(), &diags)
	activityList := acApplicationActivityListFromDataRule(dr.GetApplicationActivity(), &diags)
	actSatisfy := types.StringValue("ALL")
	if s := dr.GetApplicationActivitySatisfy(); s != nil {
		actSatisfy = types.StringValue(string(*s))
	}

	actionCfg := types.ObjectNull(applicationControlActionConfigAttrTypes)
	if cfg := dr.GetActionConfig(); cfg != nil {
		un := idname.ParseIDNameSet(ctx, cfg.GetUserNotification(), &diags)
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	dlpObj := types.ObjectNull(applicationControlDlpProfileAttrTypes)
	if dp := dr.GetDlpProfile(); dp != nil {
		o, d := types.ObjectValue(applicationControlDlpProfileAttrTypes, map[string]attr.Value{
			"content_profile": idname.ParseIDNameSet(ctx, dp.GetContentProfile(), &diags),
			"edm_profile":     idname.ParseIDNameSet(ctx, dp.GetEdmProfile(), &diags),
		})
		diags.Append(d...)
		dlpObj = o
	}

	faList := acFileAttributeListFromDataRule(ctx, dr.GetFileAttribute(), &diags)

	attrs := map[string]attr.Value{
		"action":                       types.StringValue(dr.GetAction().String()),
		"severity":                     types.StringValue(dr.GetSeverity().String()),
		"schedule":                     schObj,
		"source":                       srcObj,
		"tracking":                     trObj,
		"device":                       idname.ParseIDNameSet(ctx, dr.GetDevice(), &diags),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
		"application_criteria":         acCriteriaObjectFromDataRule(dr.GetApplicationCriteria(), &diags),
		"application_criteria_satisfy": acSatisfyStateValue(dr.ApplicationCriteriaSatisfy),
		"action_config":                actionCfg,
		"file_attribute":               faList,
		"file_attribute_satisfy":       types.StringValue(dr.GetFileAttributeSatisfy().String()),
		"dlp_profile":                  dlpObj,
	}
	o, d := types.ObjectValue(applicationControlTypedRuleAttrTypes, attrs)
	diags.Append(d...)
	return o, diags
}

func applicationControlTypedRuleStateFromFileRule(
	ctx context.Context,
	fr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if fr == nil {
		return types.ObjectNull(applicationControlTypedRuleAttrTypes), diags
	}

	schObj := acScheduleObjectFromFileRuleSchedule(ctx, fr.GetSchedule(), &diags)
	srcObj := acSourceObjectFromFileRuleSource(ctx, fr.GetSource(), &diags)
	trObj := acTrackingObjectFromFileRuleTracking(ctx, fr.GetTracking(), &diags)
	appObj := wanApplicationObjectFromFileRuleApplication(ctx, fr.GetApplication(), &diags)
	acList := acAccessMethodListFromFileRule(ctx, fr.GetAccessMethod(), &diags)
	activityList := acApplicationActivityListFromFileRule(fr.GetApplicationActivity(), &diags)
	actSatisfy := types.StringValue("ALL")
	if s := fr.GetApplicationActivitySatisfy(); s != nil {
		actSatisfy = types.StringValue(string(*s))
	}

	actionCfg := types.ObjectNull(applicationControlActionConfigAttrTypes)
	if cfg := fr.GetActionConfig(); cfg != nil {
		un := idname.ParseIDNameSet(ctx, cfg.GetUserNotification(), &diags)
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	faList := acFileAttributeListFromFileRule(ctx, fr.GetFileAttribute(), &diags)

	attrs := map[string]attr.Value{
		"action":                       types.StringValue(fr.GetAction().String()),
		"severity":                     types.StringValue(fr.GetSeverity().String()),
		"schedule":                     schObj,
		"source":                       srcObj,
		"tracking":                     trObj,
		"device":                       idname.ParseIDNameSet(ctx, fr.GetDevice(), &diags),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
		"application_criteria":         acCriteriaObjectFromFileRule(fr.GetApplicationCriteria(), &diags),
		"application_criteria_satisfy": acSatisfyStateValue(fr.ApplicationCriteriaSatisfy),
		"action_config":                actionCfg,
		"file_attribute":               faList,
		"file_attribute_satisfy":       types.StringValue(fr.GetFileAttributeSatisfy().String()),
		"dlp_profile":                  types.ObjectNull(applicationControlDlpProfileAttrTypes),
	}
	o, d := types.ObjectValue(applicationControlTypedRuleAttrTypes, attrs)
	diags.Append(d...)
	return o, diags
}

func acScheduleObjectFromApplicationRuleSchedule(
	ctx context.Context,
	sch *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	if sch == nil {
		return types.ObjectNull(policyschedule.ScheduleAttrTypes)
	}
	active := ""
	if sch.GetActiveOn() != nil {
		active = sch.GetActiveOn().String()
	}
	var ctfObj types.Object
	if ct := sch.GetCustomTimeframeApplicationRule(); ct != nil {
		o, d := types.ObjectValue(policyschedule.CustomTimeframeAttrTypes, map[string]attr.Value{
			"from": types.StringValue(utils.NormalizePolicyTimestampString(ct.GetFrom())),
			"to":   types.StringValue(utils.NormalizePolicyTimestampString(ct.GetTo())),
		})
		diags.Append(d...)
		ctfObj = o
	} else {
		ctfObj = types.ObjectNull(policyschedule.CustomTimeframeAttrTypes)
	}
	var crObj types.Object
	if cr := sch.GetCustomRecurringApplicationRule(); cr != nil {
		dayStrings := make([]string, 0, len(cr.GetDays()))
		for _, d := range cr.GetDays() {
			dayStrings = append(dayStrings, d.String())
		}
		days := convert.ParseList(ctx, types.StringType, dayStrings,
			"application_rule.schedule.custom_recurring.days")
		o, d := types.ObjectValue(policyschedule.FromToDaysAttrTypes, map[string]attr.Value{
			"from": types.StringValue(utils.NormalizePolicyRecurringClock(string(cr.From))),
			"to":   types.StringValue(utils.NormalizePolicyRecurringClock(string(cr.To))),
			"days": days,
		})
		diags.Append(d...)
		if d.HasError() || o.IsUnknown() {
			crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
		} else {
			crObj = o
		}
	} else {
		crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
	}
	o, d := types.ObjectValue(policyschedule.ScheduleAttrTypes, map[string]attr.Value{
		"active_on":        types.StringValue(active),
		"custom_timeframe": ctfObj,
		"custom_recurring": crObj,
	})
	diags.Append(d...)
	if d.HasError() || o.IsUnknown() {
		return types.ObjectNull(policyschedule.ScheduleAttrTypes)
	}
	return o
}

func acScheduleObjectFromDataRuleSchedule(
	ctx context.Context,
	sch *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	if sch == nil {
		return types.ObjectNull(policyschedule.ScheduleAttrTypes)
	}
	active := ""
	if sch.GetActiveOn() != nil {
		active = sch.GetActiveOn().String()
	}
	var ctfObj types.Object
	if ct := sch.GetCustomTimeframeDataRule(); ct != nil {
		o, d := types.ObjectValue(policyschedule.CustomTimeframeAttrTypes, map[string]attr.Value{
			"from": types.StringValue(utils.NormalizePolicyTimestampString(ct.GetFrom())),
			"to":   types.StringValue(utils.NormalizePolicyTimestampString(ct.GetTo())),
		})
		diags.Append(d...)
		ctfObj = o
	} else {
		ctfObj = types.ObjectNull(policyschedule.CustomTimeframeAttrTypes)
	}
	var crObj types.Object
	if cr := sch.GetCustomRecurringDataRule(); cr != nil {
		dayStrings := make([]string, 0, len(cr.GetDays()))
		for _, d := range cr.GetDays() {
			dayStrings = append(dayStrings, d.String())
		}
		days := convert.ParseList(ctx, types.StringType, dayStrings, "data_rule.schedule.custom_recurring.days")
		fromS := ""
		toS := ""
		if cr.GetFrom() != nil {
			fromS = utils.NormalizePolicyRecurringClock(string(*cr.GetFrom()))
		}
		if cr.GetTo() != nil {
			toS = utils.NormalizePolicyRecurringClock(string(*cr.GetTo()))
		}
		o, d := types.ObjectValue(policyschedule.FromToDaysAttrTypes, map[string]attr.Value{
			"from": types.StringValue(fromS),
			"to":   types.StringValue(toS),
			"days": days,
		})
		diags.Append(d...)
		if d.HasError() || o.IsUnknown() {
			crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
		} else {
			crObj = o
		}
	} else {
		crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
	}
	o, d := types.ObjectValue(policyschedule.ScheduleAttrTypes, map[string]attr.Value{
		"active_on":        types.StringValue(active),
		"custom_timeframe": ctfObj,
		"custom_recurring": crObj,
	})
	diags.Append(d...)
	if d.HasError() || o.IsUnknown() {
		return types.ObjectNull(policyschedule.ScheduleAttrTypes)
	}
	return o
}

func acScheduleObjectFromFileRuleSchedule(
	ctx context.Context,
	sch *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	if sch == nil {
		return types.ObjectNull(policyschedule.ScheduleAttrTypes)
	}
	active := ""
	if sch.GetActiveOn() != nil {
		active = sch.GetActiveOn().String()
	}
	var ctfObj types.Object
	if ct := sch.GetCustomTimeframeFileRule(); ct != nil {
		o, d := types.ObjectValue(policyschedule.CustomTimeframeAttrTypes, map[string]attr.Value{
			"from": types.StringValue(utils.NormalizePolicyTimestampString(ct.GetFrom())),
			"to":   types.StringValue(utils.NormalizePolicyTimestampString(ct.GetTo())),
		})
		diags.Append(d...)
		ctfObj = o
	} else {
		ctfObj = types.ObjectNull(policyschedule.CustomTimeframeAttrTypes)
	}
	var crObj types.Object
	if cr := sch.GetCustomRecurringFileRule(); cr != nil {
		dayStrings := make([]string, 0, len(cr.GetDays()))
		for _, d := range cr.GetDays() {
			dayStrings = append(dayStrings, d.String())
		}
		days := convert.ParseList(ctx, types.StringType, dayStrings, "file_rule.schedule.custom_recurring.days")
		fromS := ""
		toS := ""
		if cr.GetFrom() != nil {
			fromS = utils.NormalizePolicyRecurringClock(string(*cr.GetFrom()))
		}
		if cr.GetTo() != nil {
			toS = utils.NormalizePolicyRecurringClock(string(*cr.GetTo()))
		}
		o, d := types.ObjectValue(policyschedule.FromToDaysAttrTypes, map[string]attr.Value{
			"from": types.StringValue(fromS),
			"to":   types.StringValue(toS),
			"days": days,
		})
		diags.Append(d...)
		if d.HasError() || o.IsUnknown() {
			crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
		} else {
			crObj = o
		}
	} else {
		crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
	}
	o, d := types.ObjectValue(policyschedule.ScheduleAttrTypes, map[string]attr.Value{
		"active_on":        types.StringValue(active),
		"custom_timeframe": ctfObj,
		"custom_recurring": crObj,
	})
	diags.Append(d...)
	if d.HasError() || o.IsUnknown() {
		return types.ObjectNull(policyschedule.ScheduleAttrTypes)
	}
	return o
}

func acSourceObjectFromApplicationRuleSource(
	ctx context.Context,
	src *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Source,
	diags *diag.Diagnostics,
) types.Object {
	if src == nil {
		return types.ObjectNull(applicationcontrol.ApplicationControlSourceAttrTypes)
	}
	srcAttrs := map[string]attr.Value{
		"country": idname.ParseIDNameSet(ctx, src.GetCountry(), diags),
		"ip": convert.ParseList(ctx, types.StringType, src.GetIP(),
			"application_rule.source.ip"),
		"host": idname.ParseIDNameSet(ctx, src.GetHost(), diags),
		"site": idname.ParseIDNameSet(ctx, src.GetSite(), diags),
		"subnet": convert.ParseList(ctx, types.StringType, src.GetSubnet(),
			"application_rule.source.subnet"),
		"ip_range": convert.ParseFromToList(ctx, src.GetIPRange(),
			"application_rule.source.ip_range"),
		"global_ip_range":     idname.ParseIDNameSet(ctx, src.GetGlobalIPRange(), diags),
		"network_interface":   idname.ParseIDNameSet(ctx, src.GetNetworkInterface(), diags),
		"site_network_subnet": idname.ParseIDNameSet(ctx, src.GetSiteNetworkSubnet(), diags),
		"floating_subnet":     idname.ParseIDNameSet(ctx, src.GetFloatingSubnet(), diags),
		"user":                idname.ParseIDNameSet(ctx, src.GetUser(), diags),
		"users_group":         idname.ParseIDNameSet(ctx, src.GetUsersGroup(), diags),
		"group":               idname.ParseIDNameSet(ctx, src.GetGroup(), diags),
		"system_group":        idname.ParseIDNameSet(ctx, src.GetSystemGroup(), diags),
	}
	o, d := types.ObjectValue(applicationcontrol.ApplicationControlSourceAttrTypes, srcAttrs)
	diags.Append(d...)
	return o
}

func acSourceObjectFromDataRuleSource(
	ctx context.Context,
	src *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Source,
	diags *diag.Diagnostics,
) types.Object {
	if src == nil {
		return types.ObjectNull(applicationcontrol.ApplicationControlSourceAttrTypes)
	}
	srcAttrs := map[string]attr.Value{
		"country":             idname.ParseIDNameSet(ctx, src.GetCountry(), diags),
		"ip":                  convert.ParseList(ctx, types.StringType, src.GetIP(), "data_rule.source.ip"),
		"host":                idname.ParseIDNameSet(ctx, src.GetHost(), diags),
		"site":                idname.ParseIDNameSet(ctx, src.GetSite(), diags),
		"subnet":              convert.ParseList(ctx, types.StringType, src.GetSubnet(), "data_rule.source.subnet"),
		"ip_range":            convert.ParseFromToList(ctx, src.GetIPRange(), "data_rule.source.ip_range"),
		"global_ip_range":     idname.ParseIDNameSet(ctx, src.GetGlobalIPRange(), diags),
		"network_interface":   idname.ParseIDNameSet(ctx, src.GetNetworkInterface(), diags),
		"site_network_subnet": idname.ParseIDNameSet(ctx, src.GetSiteNetworkSubnet(), diags),
		"floating_subnet":     idname.ParseIDNameSet(ctx, src.GetFloatingSubnet(), diags),
		"user":                idname.ParseIDNameSet(ctx, src.GetUser(), diags),
		"users_group":         idname.ParseIDNameSet(ctx, src.GetUsersGroup(), diags),
		"group":               idname.ParseIDNameSet(ctx, src.GetGroup(), diags),
		"system_group":        idname.ParseIDNameSet(ctx, src.GetSystemGroup(), diags),
	}
	o, d := types.ObjectValue(applicationcontrol.ApplicationControlSourceAttrTypes, srcAttrs)
	diags.Append(d...)
	return o
}

func acSourceObjectFromFileRuleSource(
	ctx context.Context,
	src *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Source,
	diags *diag.Diagnostics,
) types.Object {
	if src == nil {
		return types.ObjectNull(applicationcontrol.ApplicationControlSourceAttrTypes)
	}
	srcAttrs := map[string]attr.Value{
		"country":             idname.ParseIDNameSet(ctx, src.GetCountry(), diags),
		"ip":                  convert.ParseList(ctx, types.StringType, src.GetIP(), "file_rule.source.ip"),
		"host":                idname.ParseIDNameSet(ctx, src.GetHost(), diags),
		"site":                idname.ParseIDNameSet(ctx, src.GetSite(), diags),
		"subnet":              convert.ParseList(ctx, types.StringType, src.GetSubnet(), "file_rule.source.subnet"),
		"ip_range":            convert.ParseFromToList(ctx, src.GetIPRange(), "file_rule.source.ip_range"),
		"global_ip_range":     idname.ParseIDNameSet(ctx, src.GetGlobalIPRange(), diags),
		"network_interface":   idname.ParseIDNameSet(ctx, src.GetNetworkInterface(), diags),
		"site_network_subnet": idname.ParseIDNameSet(ctx, src.GetSiteNetworkSubnet(), diags),
		"floating_subnet":     idname.ParseIDNameSet(ctx, src.GetFloatingSubnet(), diags),
		"user":                idname.ParseIDNameSet(ctx, src.GetUser(), diags),
		"users_group":         idname.ParseIDNameSet(ctx, src.GetUsersGroup(), diags),
		"group":               idname.ParseIDNameSet(ctx, src.GetGroup(), diags),
		"system_group":        idname.ParseIDNameSet(ctx, src.GetSystemGroup(), diags),
	}
	o, d := types.ObjectValue(applicationcontrol.ApplicationControlSourceAttrTypes, srcAttrs)
	diags.Append(d...)
	return o
}

func acTrackingObjectFromApplicationRuleTracking(
	ctx context.Context,
	tr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	if tr == nil {
		return types.ObjectNull(policyrule.TrackingAttrTypes)
	}
	ev := tr.GetEvent()
	evObj, d := types.ObjectValue(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(ev.GetEnabled()),
	})
	diags.Append(d...)
	al := tr.GetAlert()
	freq := ""
	if al.GetFrequency() != nil {
		freq = al.GetFrequency().String()
	}
	alObj, d := types.ObjectValue(policyrule.TrackingAlertAttrTypes, map[string]attr.Value{
		"enabled":            types.BoolValue(al.GetEnabled()),
		"frequency":          types.StringValue(freq),
		"subscription_group": idname.ParseIDNameSet(ctx, al.GetSubscriptionGroup(), diags),
		"webhook":            idname.ParseIDNameSet(ctx, al.GetWebhook(), diags),
		"mailing_list":       idname.ParseIDNameSet(ctx, al.GetMailingList(), diags),
	})
	diags.Append(d...)
	o, d := types.ObjectValue(policyrule.TrackingAttrTypes, map[string]attr.Value{
		"event": evObj,
		"alert": alObj,
	})
	diags.Append(d...)
	return o
}

func acTrackingObjectFromDataRuleTracking(
	ctx context.Context,
	tr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	if tr == nil {
		return types.ObjectNull(policyrule.TrackingAttrTypes)
	}
	ev := tr.GetEvent()
	evObj, d := types.ObjectValue(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(ev.GetEnabled()),
	})
	diags.Append(d...)
	al := tr.GetAlert()
	freq := ""
	if al.GetFrequency() != nil {
		freq = al.GetFrequency().String()
	}
	alObj, d := types.ObjectValue(policyrule.TrackingAlertAttrTypes, map[string]attr.Value{
		"enabled":            types.BoolValue(al.GetEnabled()),
		"frequency":          types.StringValue(freq),
		"subscription_group": idname.ParseIDNameSet(ctx, al.GetSubscriptionGroup(), diags),
		"webhook":            idname.ParseIDNameSet(ctx, al.GetWebhook(), diags),
		"mailing_list":       idname.ParseIDNameSet(ctx, al.GetMailingList(), diags),
	})
	diags.Append(d...)
	o, d := types.ObjectValue(policyrule.TrackingAttrTypes, map[string]attr.Value{
		"event": evObj,
		"alert": alObj,
	})
	diags.Append(d...)
	return o
}

func acTrackingObjectFromFileRuleTracking(
	ctx context.Context,
	tr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	if tr == nil {
		return types.ObjectNull(policyrule.TrackingAttrTypes)
	}
	ev := tr.GetEvent()
	evObj, d := types.ObjectValue(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(ev.GetEnabled()),
	})
	diags.Append(d...)
	al := tr.GetAlert()
	freq := ""
	if al.GetFrequency() != nil {
		freq = al.GetFrequency().String()
	}
	alObj, d := types.ObjectValue(policyrule.TrackingAlertAttrTypes, map[string]attr.Value{
		"enabled":            types.BoolValue(al.GetEnabled()),
		"frequency":          types.StringValue(freq),
		"subscription_group": idname.ParseIDNameSet(ctx, al.GetSubscriptionGroup(), diags),
		"webhook":            idname.ParseIDNameSet(ctx, al.GetWebhook(), diags),
		"mailing_list":       idname.ParseIDNameSet(ctx, al.GetMailingList(), diags),
	})
	diags.Append(d...)
	o, d := types.ObjectValue(policyrule.TrackingAttrTypes, map[string]attr.Value{
		"event": evObj,
		"alert": alObj,
	})
	diags.Append(d...)
	return o
}

//nolint:lll
func wanApplicationObjectFromApplicationRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(acApplicationAttrTypes)
	}
	var apps []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_Application
	if p := a.GetApplication(); p != nil {
		apps = append(apps, p)
	}
	var cats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_AppCategory
	if p := a.GetAppCategory(); p != nil {
		cats = append(cats, p)
	}
	var customs []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_CustomApp
	if p := a.GetCustomApp(); p != nil {
		customs = append(customs, p)
	}
	var customCats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_CustomCategory
	if p := a.GetCustomCategory(); p != nil {
		customCats = append(customCats, p)
	}
	var sac []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_SanctionedAppsCategory
	if p := a.GetSanctionedAppsCategory(); p != nil {
		sac = append(sac, p)
	}
	emptyStrList, _ := types.ListValue(types.StringType, []attr.Value{})
	emptyIPRangeList, _ := types.ListValue(iprange.RangeObjectType, []attr.Value{})
	emptyNameIDSet, _ := types.SetValue(idname.RefObjectType, []attr.Value{})
	attrs := map[string]attr.Value{
		"application":              idname.ParseIDNameSet(ctx, apps, diags),
		"custom_app":               idname.ParseIDNameSet(ctx, customs, diags),
		"app_category":             idname.ParseIDNameSet(ctx, cats, diags),
		"custom_category":          idname.ParseIDNameSet(ctx, customCats, diags),
		"sanctioned_apps_category": idname.ParseIDNameSet(ctx, sac, diags),
		"domain":                   emptyStrList,
		"fqdn":                     emptyStrList,
		"ip":                       emptyStrList,
		"subnet":                   emptyStrList,
		"ip_range":                 emptyIPRangeList,
		"global_ip_range":          emptyNameIDSet,
		"application_type":         acApplicationTypeListValue(a.GetApplicationType(), diags),
	}
	o, d := types.ObjectValue(acApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func wanApplicationObjectFromDataRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(acApplicationAttrTypes)
	}
	var apps []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_Application
	if p := a.GetApplication(); p != nil {
		apps = append(apps, p)
	}
	var cats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_AppCategory
	if p := a.GetAppCategory(); p != nil {
		cats = append(cats, p)
	}
	var customs []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_CustomApp
	if p := a.GetCustomApp(); p != nil {
		customs = append(customs, p)
	}
	var customCats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_CustomCategory
	if p := a.GetCustomCategory(); p != nil {
		customCats = append(customCats, p)
	}
	var sac []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_SanctionedAppsCategory
	if p := a.GetSanctionedAppsCategory(); p != nil {
		sac = append(sac, p)
	}
	emptyStrList2, _ := types.ListValue(types.StringType, []attr.Value{})
	emptyIPRangeList2, _ := types.ListValue(iprange.RangeObjectType, []attr.Value{})
	emptyNameIDSet2, _ := types.SetValue(idname.RefObjectType, []attr.Value{})
	attrs := map[string]attr.Value{
		"application":              idname.ParseIDNameSet(ctx, apps, diags),
		"custom_app":               idname.ParseIDNameSet(ctx, customs, diags),
		"app_category":             idname.ParseIDNameSet(ctx, cats, diags),
		"custom_category":          idname.ParseIDNameSet(ctx, customCats, diags),
		"sanctioned_apps_category": idname.ParseIDNameSet(ctx, sac, diags),
		"domain":                   emptyStrList2,
		"fqdn":                     emptyStrList2,
		"ip":                       emptyStrList2,
		"subnet":                   emptyStrList2,
		"ip_range":                 emptyIPRangeList2,
		"global_ip_range":          emptyNameIDSet2,
		"application_type":         acApplicationTypeListValue(a.GetApplicationType(), diags),
	}
	o, d := types.ObjectValue(acApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func wanApplicationObjectFromFileRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(acApplicationAttrTypes)
	}
	var apps []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_Application
	if p := a.GetApplication(); p != nil {
		apps = append(apps, p)
	}
	var cats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_AppCategory
	if p := a.GetAppCategory(); p != nil {
		cats = append(cats, p)
	}
	var customs []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_CustomApp
	if p := a.GetCustomApp(); p != nil {
		customs = append(customs, p)
	}
	var customCats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_CustomCategory
	if p := a.GetCustomCategory(); p != nil {
		customCats = append(customCats, p)
	}
	var sac []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_SanctionedAppsCategory
	if p := a.GetSanctionedAppsCategory(); p != nil {
		sac = append(sac, p)
	}
	emptyStrList3, _ := types.ListValue(types.StringType, []attr.Value{})
	emptyIPRangeList3, _ := types.ListValue(iprange.RangeObjectType, []attr.Value{})
	emptyNameIDSet3, _ := types.SetValue(idname.RefObjectType, []attr.Value{})
	attrs := map[string]attr.Value{
		"application":              idname.ParseIDNameSet(ctx, apps, diags),
		"custom_app":               idname.ParseIDNameSet(ctx, customs, diags),
		"app_category":             idname.ParseIDNameSet(ctx, cats, diags),
		"custom_category":          idname.ParseIDNameSet(ctx, customCats, diags),
		"sanctioned_apps_category": idname.ParseIDNameSet(ctx, sac, diags),
		"domain":                   emptyStrList3,
		"fqdn":                     emptyStrList3,
		"ip":                       emptyStrList3,
		"subnet":                   emptyStrList3,
		"ip_range":                 emptyIPRangeList3,
		"global_ip_range":          emptyNameIDSet3,
		"application_type":         acApplicationTypeListValue(a.GetApplicationType(), diags),
	}
	o, d := types.ObjectValue(acApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func acAccessMethodValueString(
	row *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_AccessMethod,
) types.String {
	if row == nil {
		return types.StringNull()
	}
	if row.GetValue() != nil {
		return types.StringValue(string(*row.GetValue()))
	}
	if vs := row.GetValueSet(); vs != nil {
		if vs.GetID() != "" {
			return types.StringValue(vs.GetID())
		}
		return types.StringValue(vs.GetName())
	}
	return types.StringNull()
}

func acAccessMethodListFromApplicationRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_AccessMethod,
	diags *diag.Diagnostics,
) types.List {
	elemType := types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}
	if len(rows) == 0 {
		return types.ListNull(elemType)
	}
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		o, d := types.ObjectValue(ApplicationControlAccessMethodAttrTypes, map[string]attr.Value{
			"access_method": types.StringValue(row.GetAccessMethod().String()),
			"operator":      types.StringValue(row.GetOperator().String()),
			"value":         acAccessMethodValueString(row),
		})
		diags.Append(d...)
		if d.HasError() {
			continue
		}
		elems = append(elems, o)
	}
	if len(elems) == 0 {
		return types.ListNull(elemType)
	}
	lst, d := types.ListValue(elemType, elems)
	diags.Append(d...)
	if d.HasError() || lst.IsUnknown() {
		return types.ListNull(elemType)
	}
	return lst
}

func acAccessMethodValueStringData(
	row *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_AccessMethod,
) types.String {
	if row == nil {
		return types.StringNull()
	}
	if row.GetValue() != nil {
		return types.StringValue(string(*row.GetValue()))
	}
	if vs := row.GetValueSet(); vs != nil {
		if vs.GetID() != "" {
			return types.StringValue(vs.GetID())
		}
		return types.StringValue(vs.GetName())
	}
	return types.StringNull()
}

func acAccessMethodListFromDataRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_AccessMethod,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		o, d := types.ObjectValue(ApplicationControlAccessMethodAttrTypes, map[string]attr.Value{
			"access_method": types.StringValue(row.GetAccessMethod().String()),
			"operator":      types.StringValue(row.GetOperator().String()),
			"value":         acAccessMethodValueStringData(row),
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acAccessMethodValueStringFile(
	row *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_AccessMethod,
) types.String {
	if row == nil {
		return types.StringNull()
	}
	if row.GetValue() != nil {
		return types.StringValue(string(*row.GetValue()))
	}
	if vs := row.GetValueSet(); vs != nil {
		if vs.GetID() != "" {
			return types.StringValue(vs.GetID())
		}
		return types.StringValue(vs.GetName())
	}
	return types.StringNull()
}

func acAccessMethodListFromFileRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_AccessMethod,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		o, d := types.ObjectValue(ApplicationControlAccessMethodAttrTypes, map[string]attr.Value{
			"access_method": types.StringValue(row.GetAccessMethod().String()),
			"operator":      types.StringValue(row.GetOperator().String()),
			"value":         acAccessMethodValueStringFile(row),
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acFileAttributeListFromDataRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_FileAttribute,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		val := types.StringNull()
		if row.GetValue() != nil {
			val = types.StringValue(string(*row.GetValue()))
		}
		ctv := make([]acNameIDState, 0, len(row.GetContentTypeValues()))
		for _, x := range row.GetContentTypeValues() {
			if x != nil {
				ctv = append(ctv, acNameIDState{ID: x.GetID(), Name: x.GetName()})
			}
		}
		ctgv := make([]acNameIDState, 0, len(row.GetContentTypeGroupValues()))
		for _, x := range row.GetContentTypeGroupValues() {
			if x != nil {
				ctgv = append(ctgv, acNameIDState{ID: x.GetID(), Name: x.GetName()})
			}
		}
		o, d := types.ObjectValue(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute":            types.StringValue(row.GetFileAttribute().String()),
			"operator":                  types.StringValue(row.GetOperator().String()),
			"value":                     val,
			"content_type_values":       acNameIDSetValue(ctv, diags),
			"content_type_group_values": acNameIDSetValue(ctgv, diags),
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acFileAttributeListFromFileRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_FileAttribute,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		val := types.StringNull()
		if row.GetValue() != nil {
			val = types.StringValue(string(*row.GetValue()))
		}
		ctv := make([]acNameIDState, 0, len(row.GetContentTypeValues()))
		for _, x := range row.GetContentTypeValues() {
			if x != nil {
				ctv = append(ctv, acNameIDState{ID: x.GetID(), Name: x.GetName()})
			}
		}
		ctgv := make([]acNameIDState, 0, len(row.GetContentTypeGroupValues()))
		for _, x := range row.GetContentTypeGroupValues() {
			if x != nil {
				ctgv = append(ctgv, acNameIDState{ID: x.GetID(), Name: x.GetName()})
			}
		}
		o, d := types.ObjectValue(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute":            types.StringValue(row.GetFileAttribute().String()),
			"operator":                  types.StringValue(row.GetOperator().String()),
			"value":                     val,
			"content_type_values":       acNameIDSetValue(ctv, diags),
			"content_type_group_values": acNameIDSetValue(ctgv, diags),
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

// acActivityRowState is the rule-type-agnostic read shape of one
// application_activity row (the generated response structs are identical in
// shape but distinct Go types).
type acActivityRowState struct {
	ID, Name                 string
	HasField                 bool
	FieldID, FieldName       string
	Operator                 string // "" = absent
	Value                    string // "" = absent
	HasValueSet              bool
	ValueSetID, ValueSetName string
}

// acBuildActivityList builds the application_activity list from extracted rows.
func acBuildActivityList(rows []acActivityRowState, diags *diag.Diagnostics) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, p := range rows {
		actObj, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(p.ID),
			"name": types.StringValue(p.Name),
		})
		diags.Append(d...)
		fieldObj := types.ObjectNull(idname.RefAttrTypes)
		if p.HasField {
			o, fd := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(p.FieldID),
				"name": types.StringValue(p.FieldName),
			})
			diags.Append(fd...)
			fieldObj = o
		}
		valueSetObj := types.ObjectNull(idname.RefAttrTypes)
		if p.HasValueSet {
			o, vd := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(p.ValueSetID),
				"name": types.StringValue(p.ValueSetName),
			})
			diags.Append(vd...)
			valueSetObj = o
		}
		operator := types.StringNull()
		if p.Operator != "" {
			operator = types.StringValue(p.Operator)
		}
		value := types.StringNull()
		if p.Value != "" {
			value = types.StringValue(p.Value)
		}
		outer, d2 := types.ObjectValue(applicationControlActivityAttrTypes, map[string]attr.Value{
			"activity":  actObj,
			"field":     fieldObj,
			"operator":  operator,
			"value":     value,
			"value_set": valueSetObj,
		})
		diags.Append(d2...)
		elems = append(elems, outer)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlActivityAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acApplicationActivityListFromApplicationRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	out := make([]acActivityRowState, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		row := acActivityRowState{ID: a.GetID(), Name: a.GetName()}
		if f := r.GetField(); f != nil {
			row.HasField, row.FieldID, row.FieldName = true, f.GetID(), f.GetName()
		}
		if op := r.GetOperator(); op != nil {
			row.Operator = string(*op)
		}
		if v := r.GetValue(); v != nil {
			row.Value = string(*v)
		}
		if vs := r.GetValueSet(); vs != nil {
			row.HasValueSet, row.ValueSetID, row.ValueSetName = true, vs.GetID(), vs.GetName()
		}
		out = append(out, row)
	}
	return acBuildActivityList(out, diags)
}

func acApplicationActivityListFromDataRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	out := make([]acActivityRowState, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		row := acActivityRowState{ID: a.GetID(), Name: a.GetName()}
		if f := r.GetField(); f != nil {
			row.HasField, row.FieldID, row.FieldName = true, f.GetID(), f.GetName()
		}
		if op := r.GetOperator(); op != nil {
			row.Operator = string(*op)
		}
		if v := r.GetValue(); v != nil {
			row.Value = string(*v)
		}
		if vs := r.GetValueSet(); vs != nil {
			row.HasValueSet, row.ValueSetID, row.ValueSetName = true, vs.GetID(), vs.GetName()
		}
		out = append(out, row)
	}
	return acBuildActivityList(out, diags)
}

func acApplicationActivityListFromFileRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	out := make([]acActivityRowState, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		row := acActivityRowState{ID: a.GetID(), Name: a.GetName()}
		if f := r.GetField(); f != nil {
			row.HasField, row.FieldID, row.FieldName = true, f.GetID(), f.GetName()
		}
		if op := r.GetOperator(); op != nil {
			row.Operator = string(*op)
		}
		if v := r.GetValue(); v != nil {
			row.Value = string(*v)
		}
		if vs := r.GetValueSet(); vs != nil {
			row.HasValueSet, row.ValueSetID, row.ValueSetName = true, vs.GetID(), vs.GetName()
		}
		out = append(out, row)
	}
	return acBuildActivityList(out, diags)
}

func applicationControlRuleRuleObjectFromPlan(p ApplicationControlRuleRulePlan) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	o, d := types.ObjectValue(ApplicationControlRuleRuleAttrTypes, map[string]attr.Value{
		"id":               p.ID,
		"name":             p.Name,
		"description":      p.Description,
		"enabled":          p.Enabled,
		"rule_type":        p.RuleType,
		"application_rule": p.ApplicationRule,
		"data_rule":        p.DataRule,
		"file_rule":        p.FileRule,
	})
	diags.Append(d...)
	return o, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(pos policyrule.Position) (resultAtObj basetypes.ObjectValue, resultAdiags diag.Diagnostics) {
	resultAtObj, resultAdiags = types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": pos.Position,
		"ref":      pos.Ref,
	})
	return resultAtObj, resultAdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(newID string, ruleObj types.Object, atObj basetypes.ObjectValue) (resultSt ApplicationControlRule) {
	resultSt = ApplicationControlRule{
		ID:   types.StringValue(newID),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	state *ApplicationControlRule,
	cur *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule,
	ruleObj types.Object,
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
	plan ApplicationControlRule,
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
func convertUpdateInput2(pos policyrule.Position) (resultAtObj basetypes.ObjectValue, resultAdiags diag.Diagnostics) {
	resultAtObj, resultAdiags = types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": pos.Position,
		"ref":      pos.Ref,
	})
	return resultAtObj, resultAdiags
}

// convertUpdateInput3 converts between the Terraform model and SDK values.
func convertUpdateInput3(
	rule ApplicationControlRuleRulePlan,
	ruleObj types.Object,
	atObj basetypes.ObjectValue,
) (resultSt ApplicationControlRule) {
	resultSt = ApplicationControlRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule ApplicationControlRuleRulePlan) (resultRm cato_models.ApplicationControlRemoveRuleInput) {
	resultRm = cato_models.ApplicationControlRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateApplicationControlRuleStateFromClient maps an Application Control policy rule into Terraform nested objects.
func hydrateApplicationControlRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule,
) (ApplicationControlRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := ApplicationControlRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	out.RuleType = types.StringValue(string(r.RuleType))

	out.ApplicationRule = types.ObjectNull(applicationControlTypedRuleAttrTypes)
	out.DataRule = types.ObjectNull(applicationControlTypedRuleAttrTypes)
	out.FileRule = types.ObjectNull(applicationControlTypedRuleAttrTypes)

	switch r.RuleType {
	case cato_models.ApplicationControlRuleTypeApplication:
		if ar := r.GetApplicationRule(); ar != nil {
			obj, d := applicationControlTypedRuleStateFromApplicationRule(ctx, ar)
			diags.Append(d...)
			out.ApplicationRule = obj
		}
	case cato_models.ApplicationControlRuleTypeData:
		if dr := r.GetDataRule(); dr != nil {
			obj, d := applicationControlTypedRuleStateFromDataRule(ctx, dr)
			diags.Append(d...)
			out.DataRule = obj
		}
	case cato_models.ApplicationControlRuleTypeFile:
		if fr := r.GetFileRule(); fr != nil {
			obj, d := applicationControlTypedRuleStateFromFileRule(ctx, fr)
			diags.Append(d...)
			out.FileRule = obj
		}
	default:
		diags.AddError("application control rule", fmt.Sprintf("unsupported rule_type %q", r.RuleType))
	}

	return out, diags
}

func acDlpProfileAddOrEmpty(
	ctx context.Context, value types.Object, diags diag.Diagnostics,
) (*cato_models.ApplicationControlDlpProfileInput, diag.Diagnostics) {
	// An omitted DLP profile still requires a non-null API input with empty lists.
	if value.IsNull() || value.IsUnknown() {
		return &cato_models.ApplicationControlDlpProfileInput{
			ContentProfile: []*cato_models.DlpContentProfileRefInput{},
			EdmProfile:     []*cato_models.DlpEdmProfileRefInput{},
		}, nil
	}
	return acDlpProfileAdd(ctx, value, diags)
}
