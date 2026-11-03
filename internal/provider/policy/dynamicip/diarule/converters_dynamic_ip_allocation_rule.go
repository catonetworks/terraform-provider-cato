package diarule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
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

var diaAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func diaPlatforms(ctx context.Context, list types.List, diags *diag.Diagnostics) []cato_models.OperatingSystem {
	out := []cato_models.OperatingSystem{}
	for _, s := range refinput.StringsFromList(ctx, list, diags) {
		out = append(out, cato_models.OperatingSystem(s))
	}
	return out
}

func diaRangeInput(
	ctx context.Context,
	rangeObj types.Object,
	diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationRangeInput {
	if rangeObj.IsNull() || rangeObj.IsUnknown() {
		return nil
	}
	rng := DynamicIpAllocationRangePlan{}
	diags.Append(rangeObj.As(ctx, &rng, diaAsOpts)...)
	if diags.HasError() {
		return nil
	}
	gir, ok := refinput.FromObject(ctx, rng.GlobalIPRange, diags, "global IP range", refinput.GlobalIPRange)
	if !ok {
		return nil
	}
	return &cato_models.DynamicIPAllocationRangeInput{GlobalIPRange: gir}
}

func diaSourceInput(
	ctx context.Context,
	sourceObj types.Object,
	diags *diag.Diagnostics,
) *cato_models.DynamicIPAllocationSourceInput {
	out := &cato_models.DynamicIPAllocationSourceInput{
		User:       []*cato_models.UserRefInput{},
		UsersGroup: []*cato_models.UsersGroupRefInput{},
	}
	if sourceObj.IsNull() || sourceObj.IsUnknown() {
		return out
	}
	src := DynamicIpAllocationSourcePlan{}
	diags.Append(sourceObj.As(ctx, &src, diaAsOpts)...)
	if diags.HasError() {
		return out
	}
	if u := refinput.FromSet(ctx, src.User, diags, "user", refinput.User); u != nil {
		out.User = u
	}
	if g := refinput.FromSet(ctx, src.UsersGroup, diags, "users group", refinput.UsersGroup); g != nil {
		out.UsersGroup = g
	}
	return out
}

func hydrateDynamicIpAllocationAddRuleInput(
	ctx context.Context,
	plan DynamicIpAllocationRule,
) (cato_models.DynamicIPAllocationAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.DynamicIPAllocationAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, diaAsOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := DynamicIpAllocationRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, diaAsOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.DynamicIPAllocationAddRuleDataInput{
		Name:                 rule.Name.ValueString(),
		Description:          rule.Description.ValueString(),
		Enabled:              rule.Enabled.ValueBool(),
		Country:              []*cato_models.CountryRefInput{},
		DevicePostureProfile: []*cato_models.DeviceProfileRefInput{},
		Platform:             diaPlatforms(ctx, rule.Platform, &diags),
	}
	if c := refinput.FromSet(ctx, rule.Country, &diags, "country", refinput.Country); c != nil {
		data.Country = c
	}
	if d := refinput.FromSet(ctx, rule.DevicePostureProfile, &diags, "device posture profile", refinput.DeviceProfile); d != nil {
		data.DevicePostureProfile = d
	}
	data.Range = diaRangeInput(ctx, rule.Range, &diags)
	data.Source = diaSourceInput(ctx, rule.Source, &diags)

	out.Rule = data
	return out, diags
}

func hydrateDynamicIpAllocationUpdateRuleInput(
	ctx context.Context,
	plan DynamicIpAllocationRule,
) (cato_models.DynamicIPAllocationUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := DynamicIpAllocationRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, diaAsOpts)...)
	if diags.HasError() {
		return cato_models.DynamicIPAllocationUpdateRuleInput{}, diags
	}

	upd := cato_models.DynamicIPAllocationUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.DynamicIPAllocationUpdateRuleDataInput{
			Name:                 rule.Name.ValueStringPointer(),
			Description:          rule.Description.ValueStringPointer(),
			Enabled:              rule.Enabled.ValueBoolPointer(),
			Country:              []*cato_models.CountryRefInput{},
			DevicePostureProfile: []*cato_models.DeviceProfileRefInput{},
			Platform:             diaPlatforms(ctx, rule.Platform, &diags),
		},
	}
	if c := refinput.FromSet(ctx, rule.Country, &diags, "country", refinput.Country); c != nil {
		upd.Rule.Country = c
	}
	if d := refinput.FromSet(ctx, rule.DevicePostureProfile, &diags, "device posture profile", refinput.DeviceProfile); d != nil {
		upd.Rule.DevicePostureProfile = d
	}
	if rng := diaRangeInput(ctx, rule.Range, &diags); rng != nil {
		upd.Rule.Range = &cato_models.DynamicIPAllocationRangeUpdateInput{GlobalIPRange: rng.GlobalIPRange}
	}
	src := diaSourceInput(ctx, rule.Source, &diags)
	upd.Rule.Source = &cato_models.DynamicIPAllocationSourceUpdateInput{
		User:       src.User,
		UsersGroup: src.UsersGroup,
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *DynamicIpAllocationRuleRulePlan,
	_ DynamicIpAllocationRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, DynamicIpAllocationRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *DynamicIpAllocationRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	_ DynamicIpAllocationRuleRulePlan,
	priorHydrated *DynamicIpAllocationRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, DynamicIpAllocationRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *DynamicIpAllocationRule,
	cur *cato_go_sdk.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule,
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
	plan DynamicIpAllocationRule,
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
	_ DynamicIpAllocationRuleRulePlan,
	priorRulePlan *DynamicIpAllocationRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, DynamicIpAllocationRuleRuleAttrTypes, rulePlan)
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
	rule DynamicIpAllocationRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt DynamicIpAllocationRule) {
	resultSt = DynamicIpAllocationRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule DynamicIpAllocationRuleRulePlan) (resultRm cato_models.DynamicIPAllocationRemoveRuleInput) {
	resultRm = cato_models.DynamicIPAllocationRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateDynamicIpAllocationRuleStateFromClient maps a policy read rule into Terraform nested objects.
func hydrateDynamicIpAllocationRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule,
) (DynamicIpAllocationRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := DynamicIpAllocationRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.Country = types.SetNull(idname.RefObjectType)
	out.DevicePostureProfile = types.SetNull(idname.RefObjectType)
	out.Platform = types.ListNull(types.StringType)
	out.Range = types.ObjectNull(DynamicIpAllocationRangeAttrTypes)
	out.Source = types.ObjectNull(DynamicIpAllocationSourceAttrTypes)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())

	if c := r.GetCountry(); len(c) > 0 {
		out.Country = oldidname.ParseNameIDList(ctx, c, "rule.country")
	}
	if dp := r.GetDevicePostureProfile(); len(dp) > 0 {
		out.DevicePostureProfile = oldidname.ParseNameIDList(ctx, dp, "rule.device_posture_profile")
	}
	if p := r.GetPlatform(); len(p) > 0 {
		plats := make([]string, 0, len(p))
		for _, os := range p {
			plats = append(plats, string(os))
		}
		out.Platform = convert.ParseList(ctx, types.StringType, plats, "rule.platform")
	}

	if rng := r.GetRange(); rng != nil {
		girObj := types.ObjectNull(idname.RefAttrTypes)
		if g := rng.GetGlobalIPRange(); g != nil && (g.GetID() != "" || g.GetName() != "") {
			o, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.GetID()),
				"name": types.StringValue(g.GetName()),
			})
			diags.Append(d...)
			girObj = o
		}
		rngObj, d := types.ObjectValue(DynamicIpAllocationRangeAttrTypes, map[string]attr.Value{
			"global_ip_range": girObj,
		})
		diags.Append(d...)
		out.Range = rngObj
	}

	if src := r.GetSource(); src != nil {
		srcObj, d := types.ObjectValue(DynamicIpAllocationSourceAttrTypes, map[string]attr.Value{
			"user":        oldidname.ParseNameIDList(ctx, src.GetUser(), "rule.source.user"),
			"users_group": oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "rule.source.users_group"),
		})
		diags.Append(d...)
		out.Source = srcObj
	}

	return out, diags
}
