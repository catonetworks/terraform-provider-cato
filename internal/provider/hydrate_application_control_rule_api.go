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

func hydrateApplicationControlAddRuleInput(
	ctx context.Context,
	plan ApplicationControlRule,
) (cato_models.ApplicationControlAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.ApplicationControlAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := PolicyRulePositionInput{}
		diags.Append(plan.At.As(ctx, &pos, basetypes.ObjectAsOptions{})...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := ApplicationControlRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
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
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlApplicationRuleInput{
		Action:                     cato_models.ApplicationControlAction(p.Action.ValueString()),
		Severity:                   cato_models.ApplicationControlSeverity(p.Severity.ValueString()),
		ApplicationActivitySatisfy: cato_models.ApplicationControlSatisfyAll,
		ApplicationCriteriaSatisfy: cato_models.ApplicationControlSatisfyAll,
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	return out, diags
}

func hydrateACApplicationRuleUpdate(
	ctx context.Context,
	o types.Object,
) (*cato_models.ApplicationControlApplicationRuleUpdateInput, diag.Diagnostics) {
	p := ApplicationControlTypedRulePlan{}
	var diags diag.Diagnostics
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlDataRuleInput{
		Action:                     cato_models.ApplicationControlAction(p.Action.ValueString()),
		Severity:                   cato_models.ApplicationControlSeverity(p.Severity.ValueString()),
		ApplicationActivitySatisfy: cato_models.ApplicationControlSatisfyAll,
		FileAttributeSatisfy:       cato_models.ApplicationControlSatisfyAll,
	}
	if !p.FileAttributeSatisfy.IsNull() {
		out.FileAttributeSatisfy = cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.DlpProfile.IsNull() {
		out.DlpProfile, diags = acDlpProfileAdd(ctx, p.DlpProfile, diags)
	}
	if !p.FileAttribute.IsNull() {
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
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
	if !p.FileAttributeSatisfy.IsNull() {
		v := cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
		out.FileAttributeSatisfy = &v
	}
	diags.Append(acFillCommonUpdate(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.DlpProfile.IsNull() {
		out.DlpProfile, diags = acDlpProfileUpdate(ctx, p.DlpProfile, diags)
	}
	if !p.FileAttribute.IsNull() {
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	out := &cato_models.ApplicationControlFileRuleInput{
		Action:                     cato_models.ApplicationControlAction(p.Action.ValueString()),
		Severity:                   cato_models.ApplicationControlSeverity(p.Severity.ValueString()),
		ApplicationActivitySatisfy: cato_models.ApplicationControlSatisfyAll,
		FileAttributeSatisfy:       cato_models.ApplicationControlSatisfyAll,
	}
	if !p.FileAttributeSatisfy.IsNull() {
		out.FileAttributeSatisfy = cato_models.ApplicationControlSatisfy(p.FileAttributeSatisfy.ValueString())
	}
	diags.Append(acFillCommonAdd(
		ctx, p, &out.Schedule, &out.Source, &out.Tracking, &out.Device, &out.AccessMethod, &out.Application, &out.ActionConfig,
	)...)
	if !p.FileAttribute.IsNull() {
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
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
	if !p.FileAttributeSatisfy.IsNull() {
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
	out := &cato_models.ApplicationControlDlpProfileInput{}
	if !p.ContentProfile.IsNull() {
		objs := make([]types.Object, 0, len(p.ContentProfile.Elements()))
		diags.Append(p.ContentProfile.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
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
	if !p.EdmProfile.IsNull() {
		objs := make([]types.Object, 0, len(p.EdmProfile.Elements()))
		diags.Append(p.EdmProfile.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
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
	if add == nil {
		return &cato_models.ApplicationControlDlpProfileUpdateInput{}, diags
	}
	return &cato_models.ApplicationControlDlpProfileUpdateInput{
		ContentProfile: add.ContentProfile,
		EdmProfile:     add.EdmProfile,
	}, diags
}

func acFileAttributesAdd(
	ctx context.Context,
	list types.List,
	diags diag.Diagnostics,
) ([]*cato_models.ApplicationControlFileAttributeInput, diag.Diagnostics) {
	out := []*cato_models.ApplicationControlFileAttributeInput{}
	if list.IsNull() {
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
	if !p.Schedule.IsNull() {
		sch := PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(p.Schedule.As(ctx, &sch, basetypes.ObjectAsOptions{})...)
		in := &cato_models.PolicyScheduleInput{ActiveOn: cato_models.PolicyActiveOnEnum(policyActiveOnAlways)}
		if !sch.ActiveOn.IsNull() {
			in.ActiveOn = cato_models.PolicyActiveOnEnum(sch.ActiveOn.ValueString())
		}
		if !sch.CustomTimeframe.IsNull() {
			ctf := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
			diags.Append(sch.CustomTimeframe.As(ctx, &ctf, basetypes.ObjectAsOptions{})...)
			in.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{From: ctf.From.ValueString(), To: ctf.To.ValueString()}
		}
		if !sch.CustomRecurring.IsNull() {
			cr := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
			diags.Append(sch.CustomRecurring.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
			in.CustomRecurring = &cato_models.PolicyCustomRecurringInput{
				From: cato_scalars.Time(cr.From.ValueString()),
				To:   cato_scalars.Time(cr.To.ValueString()),
			}
			diags.Append(cr.Days.ElementsAs(ctx, &in.CustomRecurring.Days, false)...)
		}
		*schedule = in
	}
	if !p.Source.IsNull() {
		add, _, sdiags := applicationControlSourcePairFromTerraformObject(ctx, p.Source)
		diags.Append(sdiags...)
		*source = add
	}
	if !p.Tracking.IsNull() {
		*tracking, diags = acTrackingAdd(ctx, p.Tracking, diags)
	}
	if !p.Device.IsNull() {
		objs := make([]types.Object, 0, len(p.Device.Elements()))
		diags.Append(p.Device.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
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
	if !p.AccessMethod.IsNull() {
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
	if !p.Application.IsNull() {
		*application, diags = acApplicationInputFromWanShape(ctx, p.Application, diags)
	}
	if !p.ActionConfig.IsNull() {
		*actionConfig, diags = acActionConfigAdd(ctx, p.ActionConfig, diags)
	}
	return diags
}

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
	if !p.Schedule.IsNull() {
		sch := PolicyPolicyWanFirewallPolicyRulesRuleSchedule{}
		diags.Append(p.Schedule.As(ctx, &sch, basetypes.ObjectAsOptions{})...)
		upd := &cato_models.PolicyScheduleUpdateInput{
			ActiveOn: (*cato_models.PolicyActiveOnEnum)(sch.ActiveOn.ValueStringPointer()),
		}
		if !sch.CustomTimeframe.IsNull() {
			ctf := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe{}
			diags.Append(sch.CustomTimeframe.As(ctx, &ctf, basetypes.ObjectAsOptions{})...)
			upd.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{
				From: ctf.From.ValueStringPointer(), To: ctf.To.ValueStringPointer(),
			}
		} else {
			upd.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}
		}
		if !sch.CustomRecurring.IsNull() {
			cr := PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring{}
			diags.Append(sch.CustomRecurring.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
			upd.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{
				From: (*cato_scalars.Time)(cr.From.ValueStringPointer()),
				To:   (*cato_scalars.Time)(cr.To.ValueStringPointer()),
			}
			diags.Append(cr.Days.ElementsAs(ctx, &upd.CustomRecurring.Days, false)...)
		} else {
			upd.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}
		}
		*schedule = upd
	}
	if !p.Source.IsNull() {
		_, supd, sdiags := applicationControlSourcePairFromTerraformObject(ctx, p.Source)
		diags.Append(sdiags...)
		*source = supd
	}
	if !p.Tracking.IsNull() {
		*tracking, diags = acTrackingUpdate(ctx, p.Tracking, diags)
	}
	if !p.Device.IsNull() {
		objs := make([]types.Object, 0, len(p.Device.Elements()))
		diags.Append(p.Device.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
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
	if !p.AccessMethod.IsNull() {
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
	if !p.Application.IsNull() {
		*application, diags = acApplicationUpdateFromWanShape(ctx, p.Application, diags)
	}
	if !p.ActionConfig.IsNull() {
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
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{})...)
	out := &cato_models.ApplicationControlActionConfigInput{}
	if !p.UserNotification.IsNull() {
		objs := make([]types.Object, 0, len(p.UserNotification.Elements()))
		diags.Append(p.UserNotification.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
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
	wan := PolicyPolicyWanFirewallPolicyRulesRuleApplication{}
	diags.Append(o.As(ctx, &wan, basetypes.ObjectAsOptions{})...)
	out := &cato_models.ApplicationControlApplicationInput{}
	appendFirst := func(set types.Set, setter func(*cato_models.ApplicationRefInput)) {
		if set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		if len(objs) == 0 {
			return
		}
		var item PolicyPolicyWanFirewallPolicyRulesRuleApplicationApplication
		diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
		ref, err := utils.TransformObjectRefInput(item)
		if err != nil {
			diags.AddError("application", err.Error())
			return
		}
		setter(&cato_models.ApplicationRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input})
	}
	appendFirstCat := func(set types.Set, setter func(*cato_models.ApplicationCategoryRefInput)) {
		if set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		if len(objs) == 0 {
			return
		}
		var item PolicyPolicyWanFirewallPolicyRulesRuleApplicationAppCategory
		diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
		ref, err := utils.TransformObjectRefInput(item)
		if err != nil {
			diags.AddError("application category", err.Error())
			return
		}
		setter(&cato_models.ApplicationCategoryRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input})
	}
	appendFirst(wan.Application, func(v *cato_models.ApplicationRefInput) { out.Application = v })
	if out.Application == nil && !wan.CustomApp.IsNull() {
		objs := make([]types.Object, 0, len(wan.CustomApp.Elements()))
		diags.Append(wan.CustomApp.ElementsAs(ctx, &objs, false)...)
		if len(objs) > 0 {
			var item PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomApp
			diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err == nil {
				out.CustomApp = &cato_models.CustomApplicationRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input}
			}
		}
	}
	appendFirstCat(wan.AppCategory, func(v *cato_models.ApplicationCategoryRefInput) { out.AppCategory = v })
	if !wan.CustomCategory.IsNull() {
		objs := make([]types.Object, 0, len(wan.CustomCategory.Elements()))
		diags.Append(wan.CustomCategory.ElementsAs(ctx, &objs, false)...)
		if len(objs) > 0 {
			var item PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomCategory
			diags.Append(objs[0].As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err == nil {
				out.CustomCategory = &cato_models.CustomCategoryRefInput{By: cato_models.ObjectRefBy(ref.By), Input: ref.Input}
			}
		}
	}
	if !wan.SanctionedAppsCategory.IsNull() {
		objs := make([]types.Object, 0, len(wan.SanctionedAppsCategory.Elements()))
		diags.Append(wan.SanctionedAppsCategory.ElementsAs(ctx, &objs, false)...)
		if len(objs) > 0 {
			var item PolicyPolicyWanFirewallPolicyRulesRuleApplicationSanctionedAppsCategory
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
	t := PolicyPolicyWanFirewallPolicyRulesRuleTracking{}
	diags.Append(o.As(ctx, &t, basetypes.ObjectAsOptions{})...)
	out := &cato_models.PolicyTrackingInput{
		Event: &cato_models.PolicyRuleTrackingEventInput{},
		Alert: &cato_models.PolicyRuleTrackingAlertInput{Enabled: false, Frequency: "DAILY"},
	}
	if !t.Event.IsNull() {
		ev := PolicyPolicyWanFirewallPolicyRulesRuleTrackingEvent{}
		diags.Append(t.Event.As(ctx, &ev, basetypes.ObjectAsOptions{})...)
		out.Event.Enabled = ev.Enabled.ValueBool()
	}
	if !t.Alert.IsNull() {
		al := PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlert{}
		diags.Append(t.Alert.As(ctx, &al, basetypes.ObjectAsOptions{})...)
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
