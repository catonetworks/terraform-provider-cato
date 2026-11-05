package tsrule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// globalIPRangeRefFromObject converts an {id,name} object into a GlobalIPRangeRefInput.
func globalIPRangeRefFromObject(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.GlobalIPRangeRefInput {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	ref := idname.Ref{}
	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	diags.Append(obj.As(ctx, &ref, asOpts)...)
	if diags.HasError() {
		return nil
	}
	out, err := utils.TransformObjectRefInput(ref)
	if err != nil {
		diags.AddError("Invalid global IP range reference", err.Error())
		return nil
	}
	return &cato_models.GlobalIPRangeRefInput{
		By:    cato_models.ObjectRefBy(out.By),
		Input: out.Input,
	}
}

// globalIPRangeRefsFromSet converts a set of {id,name} objects into GlobalIPRangeRefInputs.
func globalIPRangeRefsFromSet(
	ctx context.Context,
	set types.Set,
	diags *diag.Diagnostics,
) []*cato_models.GlobalIPRangeRefInput {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var refs []idname.Ref
	diags.Append(set.ElementsAs(ctx, &refs, false)...)
	if diags.HasError() {
		return nil
	}
	out := make([]*cato_models.GlobalIPRangeRefInput, 0, len(refs))
	for _, ref := range refs {
		r, err := utils.TransformObjectRefInput(ref)
		if err != nil {
			diags.AddError("Invalid global IP range reference", err.Error())
			return nil
		}
		out = append(out, &cato_models.GlobalIPRangeRefInput{
			By:    cato_models.ObjectRefBy(r.By),
			Input: r.Input,
		})
	}
	return out
}

func hydrateTerminalServerAddRuleInput(
	ctx context.Context,
	plan TerminalServerRule,
) (cato_models.TerminalServerAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.TerminalServerAddRuleInput{}

	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, asOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := TerminalServerRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, asOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.TerminalServerAddRuleDataInput{
		Name:        rule.Name.ValueString(),
		Description: rule.Description.ValueString(),
		Enabled:     rule.Enabled.ValueBool(),
	}
	data.AllowedHostIP = globalIPRangeRefFromObject(ctx, rule.AllowedHostIP, &diags)
	data.ExcludeTraffic = globalIPRangeRefsFromSet(ctx, rule.ExcludeTraffic, &diags)
	if data.ExcludeTraffic == nil {
		data.ExcludeTraffic = []*cato_models.GlobalIPRangeRefInput{}
	}

	out.Rule = data
	return out, diags
}

func hydrateTerminalServerUpdateRuleInput(
	ctx context.Context,
	plan TerminalServerRule,
) (cato_models.TerminalServerUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	asOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	rule := TerminalServerRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, asOpts)...)
	if diags.HasError() {
		return cato_models.TerminalServerUpdateRuleInput{}, diags
	}

	upd := cato_models.TerminalServerUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.TerminalServerUpdateRuleDataInput{
			Name:        rule.Name.ValueStringPointer(),
			Description: rule.Description.ValueStringPointer(),
			Enabled:     rule.Enabled.ValueBoolPointer(),
		},
	}
	upd.Rule.AllowedHostIP = globalIPRangeRefFromObject(ctx, rule.AllowedHostIP, &diags)
	upd.Rule.ExcludeTraffic = globalIPRangeRefsFromSet(ctx, rule.ExcludeTraffic, &diags)
	if upd.Rule.ExcludeTraffic == nil {
		upd.Rule.ExcludeTraffic = []*cato_models.GlobalIPRangeRefInput{}
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *TerminalServerRuleRulePlan,
	_ TerminalServerRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, TerminalServerRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *TerminalServerRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	_ TerminalServerRuleRulePlan,
	priorHydrated *TerminalServerRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, TerminalServerRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *TerminalServerRule,
	cur *cato_go_sdk.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Rules_Rule,
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
	plan TerminalServerRule,
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
	_ TerminalServerRuleRulePlan,
	priorRulePlan *TerminalServerRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, TerminalServerRuleRuleAttrTypes, rulePlan)
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
	rule TerminalServerRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt TerminalServerRule) {
	resultSt = TerminalServerRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule TerminalServerRuleRulePlan) (resultRm cato_models.TerminalServerRemoveRuleInput) {
	resultRm = cato_models.TerminalServerRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateTerminalServerRuleStateFromClient maps a policy read rule into Terraform nested objects.
func hydrateTerminalServerRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Rules_Rule,
) (TerminalServerRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := TerminalServerRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.AllowedHostIP = types.ObjectNull(idname.RefAttrTypes)
	out.ExcludeTraffic = types.SetNull(idname.RefObjectType)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())

	if h := r.GetAllowedHostIP(); h != nil && (h.GetID() != "" || h.GetName() != "") {
		hostObj, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(h.GetID()),
			"name": types.StringValue(h.GetName()),
		})
		diags.Append(d...)
		out.AllowedHostIP = hostObj
	}

	if ex := r.GetExcludeTraffic(); len(ex) > 0 {
		out.ExcludeTraffic = oldidname.ParseNameIDList(ctx, ex, "rule.exclude_traffic")
	}

	return out, diags
}
