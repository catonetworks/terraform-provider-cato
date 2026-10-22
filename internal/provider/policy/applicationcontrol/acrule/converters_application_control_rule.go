package acrule

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
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
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
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if out.ActionConfig == nil {
		out.ActionConfig = acEmptyApplicationControlActionConfigInput()
	}
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
	if !p.FileAttributeSatisfy.IsNull() && !p.FileAttributeSatisfy.IsUnknown() && p.FileAttributeSatisfy.ValueString() != "" {
		out.FileAttributeSatisfy = cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if p.DlpProfile.IsNull() || p.DlpProfile.IsUnknown() {
		diags.AddError(
			"data_rule.dlp_profile required",
			"The Cato API requires a data rule to include dlp_profile with at least one content_profile or edm_profile reference.",
		)
		return nil, diags
	}
	dlp, dlpDiags := acDlpProfileAdd(ctx, p.DlpProfile, diags)
	diags.Append(dlpDiags...)
	if len(dlp.ContentProfile) == 0 && len(dlp.EdmProfile) == 0 {
		diags.AddError(
			"data_rule.dlp_profile invalid",
			"dlp_profile must contain at least one content_profile or edm_profile entry.",
		)
		return nil, diags
	}
	out.DlpProfile = dlp
	if !p.FileAttribute.IsNull() && !p.FileAttribute.IsUnknown() {
		out.FileAttribute, diags = acFileAttributesAdd(ctx, p.FileAttribute, diags)
	}
	return out, diags
}

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
	if !p.FileAttributeSatisfy.IsNull() && !p.FileAttributeSatisfy.IsUnknown() && p.FileAttributeSatisfy.ValueString() != "" {
		out.FileAttributeSatisfy = cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.FileAttribute.IsNull() && !p.FileAttribute.IsUnknown() {
		out.FileAttribute, diags = acFileAttributesAdd(ctx, p.FileAttribute, diags)
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
		out = append(out, &cato_models.ApplicationControlActivityInput{
			Activity: &cato_models.ApplicationControlActivityRefInput{
				By:    cato_models.ObjectRefBy(ref.By),
				Input: ref.Input,
			},
		})
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
		FileAttribute types.String `tfsdk:"file_attribute"`
		Operator      types.String `tfsdk:"operator"`
		Value         types.String `tfsdk:"value"`
	}
	for _, el := range objs {
		diags.Append(el.As(ctx, &row, basetypes.ObjectAsOptions{})...)
		out = append(out, &cato_models.ApplicationControlFileAttributeInput{
			FileAttribute: cato_models.ApplicationControlFileAttributeType(row.FileAttribute.ValueString()),
			Operator:      cato_models.ApplicationControlOperator(row.Operator.ValueString()),
			Value:         row.Value.ValueStringPointer(),
		})
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
	wan := wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplication{}
	diags.Append(o.As(ctx, &wan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	out := &cato_models.ApplicationControlApplicationInput{}
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
		un := oldidname.ParseNameIDList(ctx, cfg.GetUserNotification(),
			"application_rule.action_config.user_notification")
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	attrs := map[string]attr.Value{
		"action":   types.StringValue(ar.GetAction().String()),
		"severity": types.StringValue(ar.GetSeverity().String()),
		"schedule": schObj,
		"source":   srcObj,
		"tracking": trObj,
		"device": oldidname.ParseNameIDList(ctx, ar.GetDevice(),
			"application_rule.device"),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
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
		un := oldidname.ParseNameIDList(ctx, cfg.GetUserNotification(), "data_rule.action_config.user_notification")
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	dlpObj := types.ObjectNull(applicationControlDlpProfileAttrTypes)
	if dp := dr.GetDlpProfile(); dp != nil {
		o, d := types.ObjectValue(applicationControlDlpProfileAttrTypes, map[string]attr.Value{
			"content_profile": oldidname.ParseNameIDList(ctx, dp.GetContentProfile(), "data_rule.dlp_profile.content_profile"),
			"edm_profile":     oldidname.ParseNameIDList(ctx, dp.GetEdmProfile(), "data_rule.dlp_profile.edm_profile"),
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
		"device":                       oldidname.ParseNameIDList(ctx, dr.GetDevice(), "data_rule.device"),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
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
		un := oldidname.ParseNameIDList(ctx, cfg.GetUserNotification(), "file_rule.action_config.user_notification")
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
		"device":                       oldidname.ParseNameIDList(ctx, fr.GetDevice(), "file_rule.device"),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
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
		"country": oldidname.ParseNameIDList(ctx, src.GetCountry(),
			"application_rule.source.country"),
		"ip": convert.ParseList(ctx, types.StringType, src.GetIP(),
			"application_rule.source.ip"),
		"host": oldidname.ParseNameIDList(ctx, src.GetHost(),
			"application_rule.source.host"),
		"site": oldidname.ParseNameIDList(ctx, src.GetSite(),
			"application_rule.source.site"),
		"subnet": convert.ParseList(ctx, types.StringType, src.GetSubnet(),
			"application_rule.source.subnet"),
		"ip_range": convert.ParseFromToList(ctx, src.GetIPRange(),
			"application_rule.source.ip_range"),
		"global_ip_range": oldidname.ParseNameIDList(ctx, src.GetGlobalIPRange(),
			"application_rule.source.global_ip_range"),
		"network_interface": oldidname.ParseNameIDList(ctx, src.GetNetworkInterface(),
			"application_rule.source.network_interface"),
		"site_network_subnet": oldidname.ParseNameIDList(ctx, src.GetSiteNetworkSubnet(),
			"application_rule.source.site_network_subnet"),
		"floating_subnet": oldidname.ParseNameIDList(ctx, src.GetFloatingSubnet(),
			"application_rule.source.floating_subnet"),
		"user": oldidname.ParseNameIDList(ctx, src.GetUser(),
			"application_rule.source.user"),
		"users_group": oldidname.ParseNameIDList(ctx, src.GetUsersGroup(),
			"application_rule.source.users_group"),
		"group": oldidname.ParseNameIDList(ctx, src.GetGroup(),
			"application_rule.source.group"),
		"system_group": oldidname.ParseNameIDList(ctx, src.GetSystemGroup(),
			"application_rule.source.system_group"),
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
		"country":             oldidname.ParseNameIDList(ctx, src.GetCountry(), "data_rule.source.country"),
		"ip":                  convert.ParseList(ctx, types.StringType, src.GetIP(), "data_rule.source.ip"),
		"host":                oldidname.ParseNameIDList(ctx, src.GetHost(), "data_rule.source.host"),
		"site":                oldidname.ParseNameIDList(ctx, src.GetSite(), "data_rule.source.site"),
		"subnet":              convert.ParseList(ctx, types.StringType, src.GetSubnet(), "data_rule.source.subnet"),
		"ip_range":            convert.ParseFromToList(ctx, src.GetIPRange(), "data_rule.source.ip_range"),
		"global_ip_range":     oldidname.ParseNameIDList(ctx, src.GetGlobalIPRange(), "data_rule.source.global_ip_range"),
		"network_interface":   oldidname.ParseNameIDList(ctx, src.GetNetworkInterface(), "data_rule.source.network_interface"),
		"site_network_subnet": oldidname.ParseNameIDList(ctx, src.GetSiteNetworkSubnet(), "data_rule.source.site_network_subnet"),
		"floating_subnet":     oldidname.ParseNameIDList(ctx, src.GetFloatingSubnet(), "data_rule.source.floating_subnet"),
		"user":                oldidname.ParseNameIDList(ctx, src.GetUser(), "data_rule.source.user"),
		"users_group":         oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "data_rule.source.users_group"),
		"group":               oldidname.ParseNameIDList(ctx, src.GetGroup(), "data_rule.source.group"),
		"system_group":        oldidname.ParseNameIDList(ctx, src.GetSystemGroup(), "data_rule.source.system_group"),
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
		"country":             oldidname.ParseNameIDList(ctx, src.GetCountry(), "file_rule.source.country"),
		"ip":                  convert.ParseList(ctx, types.StringType, src.GetIP(), "file_rule.source.ip"),
		"host":                oldidname.ParseNameIDList(ctx, src.GetHost(), "file_rule.source.host"),
		"site":                oldidname.ParseNameIDList(ctx, src.GetSite(), "file_rule.source.site"),
		"subnet":              convert.ParseList(ctx, types.StringType, src.GetSubnet(), "file_rule.source.subnet"),
		"ip_range":            convert.ParseFromToList(ctx, src.GetIPRange(), "file_rule.source.ip_range"),
		"global_ip_range":     oldidname.ParseNameIDList(ctx, src.GetGlobalIPRange(), "file_rule.source.global_ip_range"),
		"network_interface":   oldidname.ParseNameIDList(ctx, src.GetNetworkInterface(), "file_rule.source.network_interface"),
		"site_network_subnet": oldidname.ParseNameIDList(ctx, src.GetSiteNetworkSubnet(), "file_rule.source.site_network_subnet"),
		"floating_subnet":     oldidname.ParseNameIDList(ctx, src.GetFloatingSubnet(), "file_rule.source.floating_subnet"),
		"user":                oldidname.ParseNameIDList(ctx, src.GetUser(), "file_rule.source.user"),
		"users_group":         oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "file_rule.source.users_group"),
		"group":               oldidname.ParseNameIDList(ctx, src.GetGroup(), "file_rule.source.group"),
		"system_group":        oldidname.ParseNameIDList(ctx, src.GetSystemGroup(), "file_rule.source.system_group"),
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
		"enabled":   types.BoolValue(al.GetEnabled()),
		"frequency": types.StringValue(freq),
		"subscription_group": oldidname.ParseNameIDListOrEmptySet(ctx, al.GetSubscriptionGroup(),
			"application_rule.tracking.alert.subscription_group"),
		"webhook": oldidname.ParseNameIDListOrEmptySet(ctx, al.GetWebhook(),
			"application_rule.tracking.alert.webhook"),
		"mailing_list": oldidname.ParseNameIDListOrEmptySet(ctx, al.GetMailingList(),
			"application_rule.tracking.alert.mailing_list"),
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
		"subscription_group": oldidname.ParseNameIDListOrEmptySet(ctx, al.GetSubscriptionGroup(), "data_rule.tracking.alert.subscription_group"),
		"webhook":            oldidname.ParseNameIDListOrEmptySet(ctx, al.GetWebhook(), "data_rule.tracking.alert.webhook"),
		"mailing_list":       oldidname.ParseNameIDListOrEmptySet(ctx, al.GetMailingList(), "data_rule.tracking.alert.mailing_list"),
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
		"subscription_group": oldidname.ParseNameIDListOrEmptySet(ctx, al.GetSubscriptionGroup(), "file_rule.tracking.alert.subscription_group"),
		"webhook":            oldidname.ParseNameIDListOrEmptySet(ctx, al.GetWebhook(), "file_rule.tracking.alert.webhook"),
		"mailing_list":       oldidname.ParseNameIDListOrEmptySet(ctx, al.GetMailingList(), "file_rule.tracking.alert.mailing_list"),
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
		return types.ObjectNull(wfrule.WanApplicationAttrTypes)
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
		"application": oldidname.ParseNameIDListOrEmptySet(ctx, apps,
			"application_rule.application.application"),
		"custom_app": oldidname.ParseNameIDListOrEmptySet(ctx, customs,
			"application_rule.application.custom_app"),
		"app_category": oldidname.ParseNameIDListOrEmptySet(ctx, cats,
			"application_rule.application.app_category"),
		"custom_category": oldidname.ParseNameIDListOrEmptySet(ctx, customCats,
			"application_rule.application.custom_category"),
		"sanctioned_apps_category": oldidname.ParseNameIDListOrEmptySet(ctx, sac,
			"application_rule.application.sanctioned_apps_category"),
		"domain":          emptyStrList,
		"fqdn":            emptyStrList,
		"ip":              emptyStrList,
		"subnet":          emptyStrList,
		"ip_range":        emptyIPRangeList,
		"global_ip_range": emptyNameIDSet,
	}
	o, d := types.ObjectValue(wfrule.WanApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func wanApplicationObjectFromDataRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(wfrule.WanApplicationAttrTypes)
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
		"application":              oldidname.ParseNameIDListOrEmptySet(ctx, apps, "data_rule.application.application"),
		"custom_app":               oldidname.ParseNameIDListOrEmptySet(ctx, customs, "data_rule.application.custom_app"),
		"app_category":             oldidname.ParseNameIDListOrEmptySet(ctx, cats, "data_rule.application.app_category"),
		"custom_category":          oldidname.ParseNameIDListOrEmptySet(ctx, customCats, "data_rule.application.custom_category"),
		"sanctioned_apps_category": oldidname.ParseNameIDListOrEmptySet(ctx, sac, "data_rule.application.sanctioned_apps_category"),
		"domain":                   emptyStrList2,
		"fqdn":                     emptyStrList2,
		"ip":                       emptyStrList2,
		"subnet":                   emptyStrList2,
		"ip_range":                 emptyIPRangeList2,
		"global_ip_range":          emptyNameIDSet2,
	}
	o, d := types.ObjectValue(wfrule.WanApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func wanApplicationObjectFromFileRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(wfrule.WanApplicationAttrTypes)
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
		"application":              oldidname.ParseNameIDListOrEmptySet(ctx, apps, "file_rule.application.application"),
		"custom_app":               oldidname.ParseNameIDListOrEmptySet(ctx, customs, "file_rule.application.custom_app"),
		"app_category":             oldidname.ParseNameIDListOrEmptySet(ctx, cats, "file_rule.application.app_category"),
		"custom_category":          oldidname.ParseNameIDListOrEmptySet(ctx, customCats, "file_rule.application.custom_category"),
		"sanctioned_apps_category": oldidname.ParseNameIDListOrEmptySet(ctx, sac, "file_rule.application.sanctioned_apps_category"),
		"domain":                   emptyStrList3,
		"fqdn":                     emptyStrList3,
		"ip":                       emptyStrList3,
		"subnet":                   emptyStrList3,
		"ip_range":                 emptyIPRangeList3,
		"global_ip_range":          emptyNameIDSet3,
	}
	o, d := types.ObjectValue(wfrule.WanApplicationAttrTypes, attrs)
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
		o, d := types.ObjectValue(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute": types.StringValue(row.GetFileAttribute().String()),
			"operator":       types.StringValue(row.GetOperator().String()),
			"value":          val,
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
		o, d := types.ObjectValue(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute": types.StringValue(row.GetFileAttribute().String()),
			"operator":       types.StringValue(row.GetOperator().String()),
			"value":          val,
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

// acBuildActivityList builds the application_activity list from pre-extracted (id, name) pairs.
func acBuildActivityList(pairs [][2]string, diags *diag.Diagnostics) types.List {
	elems := make([]attr.Value, 0, len(pairs))
	for _, p := range pairs {
		actObj, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(p[0]),
			"name": types.StringValue(p[1]),
		})
		diags.Append(d...)
		outer, d2 := types.ObjectValue(applicationControlActivityAttrTypes, map[string]attr.Value{
			"activity": actObj,
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
	pairs := make([][2]string, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		pairs = append(pairs, [2]string{a.GetID(), a.GetName()})
	}
	return acBuildActivityList(pairs, diags)
}

func acApplicationActivityListFromDataRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	pairs := make([][2]string, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		pairs = append(pairs, [2]string{a.GetID(), a.GetName()})
	}
	return acBuildActivityList(pairs, diags)
}

func acApplicationActivityListFromFileRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	pairs := make([][2]string, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		pairs = append(pairs, [2]string{a.GetID(), a.GetName()})
	}
	return acBuildActivityList(pairs, diags)
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
