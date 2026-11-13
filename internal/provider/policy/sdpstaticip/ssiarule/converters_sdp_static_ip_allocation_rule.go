package ssiarule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// userRefFromSource extracts the single-user source selector as a UserRefInput.
func userRefFromSource(
	ctx context.Context,
	sourceObj types.Object,
	diags *diag.Diagnostics,
) *cato_models.UserRefInput {
	if sourceObj.IsNull() || sourceObj.IsUnknown() {
		return nil
	}
	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	src := SdpStaticIPAllocationSourcePlan{}
	diags.Append(sourceObj.As(ctx, &src, asOpts)...)
	if diags.HasError() || src.User.IsNull() || src.User.IsUnknown() {
		return nil
	}
	ref := idname.Ref{}
	diags.Append(src.User.As(ctx, &ref, asOpts)...)
	if diags.HasError() {
		return nil
	}
	out, err := utils.TransformObjectRefInput(ref)
	if err != nil {
		diags.AddError("Invalid user reference", err.Error())
		return nil
	}
	return &cato_models.UserRefInput{
		By:    cato_models.ObjectRefBy(out.By),
		Input: out.Input,
	}
}

func hydrateSdpStaticIPAllocationAddRuleInput(
	ctx context.Context,
	plan SdpStaticIPAllocationRule,
) (cato_models.SdpStaticIPAllocationAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.SdpStaticIPAllocationAddRuleInput{}

	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, asOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := SdpStaticIPAllocationRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, asOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.SdpStaticIPAllocationAddRuleDataInput{
		Name:        rule.Name.ValueString(),
		Description: rule.Description.ValueString(),
		Enabled:     rule.Enabled.ValueBool(),
		IP:          rule.IP.ValueString(),
		Source:      &cato_models.SdpStaticIPAllocationSourceInput{},
	}
	if u := userRefFromSource(ctx, rule.Source, &diags); u != nil {
		data.Source.User = u
	}

	out.Rule = data
	return out, diags
}

func hydrateSdpStaticIPAllocationUpdateRuleInput(
	ctx context.Context,
	plan SdpStaticIPAllocationRule,
) (cato_models.SdpStaticIPAllocationUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	rule := SdpStaticIPAllocationRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, asOpts)...)
	if diags.HasError() {
		return cato_models.SdpStaticIPAllocationUpdateRuleInput{}, diags
	}

	upd := cato_models.SdpStaticIPAllocationUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.SdpStaticIPAllocationUpdateRuleDataInput{
			Name:        rule.Name.ValueStringPointer(),
			Description: rule.Description.ValueStringPointer(),
			Enabled:     rule.Enabled.ValueBoolPointer(),
			IP:          rule.IP.ValueStringPointer(),
			Source:      &cato_models.SdpStaticIPAllocationSourceUpdateInput{},
		},
	}
	if u := userRefFromSource(ctx, rule.Source, &diags); u != nil {
		upd.Rule.Source.User = u
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *SdpStaticIPAllocationRuleRulePlan,
	_ SdpStaticIPAllocationRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SdpStaticIPAllocationRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *SdpStaticIPAllocationRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	_ SdpStaticIPAllocationRuleRulePlan,
	priorHydrated *SdpStaticIPAllocationRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SdpStaticIPAllocationRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *SdpStaticIPAllocationRule,
	cur *cato_go_sdk.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Rules_Rule,
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
	plan SdpStaticIPAllocationRule,
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
	_ SdpStaticIPAllocationRuleRulePlan,
	priorRulePlan *SdpStaticIPAllocationRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SdpStaticIPAllocationRuleRuleAttrTypes, rulePlan)
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
	rule SdpStaticIPAllocationRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt SdpStaticIPAllocationRule) {
	resultSt = SdpStaticIPAllocationRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule SdpStaticIPAllocationRuleRulePlan) (resultRm cato_models.SdpStaticIPAllocationRemoveRuleInput) {
	resultRm = cato_models.SdpStaticIPAllocationRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateSdpStaticIPAllocationRuleStateFromClient maps a policy read rule into Terraform nested objects.
func hydrateSdpStaticIPAllocationRuleStateFromClient(
	_ context.Context,
	r *cato_go_sdk.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Rules_Rule,
) (SdpStaticIPAllocationRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := SdpStaticIPAllocationRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.Source = types.ObjectNull(SdpStaticIPAllocationSourceAttrTypes)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	out.IP = types.StringValue(r.GetIP())

	if src := r.GetSource(); src != nil {
		userObj := types.ObjectNull(idname.RefAttrTypes)
		if u := src.GetUser(); u != nil && (u.GetID() != "" || u.GetName() != "") {
			o, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(u.GetID()),
				"name": types.StringValue(u.GetName()),
			})
			diags.Append(d...)
			userObj = o
		}
		srcObj, d := types.ObjectValue(SdpStaticIPAllocationSourceAttrTypes, map[string]attr.Value{
			"user": userObj,
		})
		diags.Append(d...)
		out.Source = srcObj
	}

	return out, diags
}
