package ccrule

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

var ccAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func ccSourceInput(
	ctx context.Context,
	sourceObj types.Object,
	diags *diag.Diagnostics,
) *cato_models.ClientConnectivitySourceInput {
	out := &cato_models.ClientConnectivitySourceInput{
		User:       []*cato_models.UserRefInput{},
		UsersGroup: []*cato_models.UsersGroupRefInput{},
	}
	if sourceObj.IsNull() || sourceObj.IsUnknown() {
		return out
	}
	src := ClientConnectivitySourcePlan{}
	diags.Append(sourceObj.As(ctx, &src, ccAsOpts)...)
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

func ccSourceRanges(
	ctx context.Context,
	list types.List,
	diags *diag.Diagnostics,
) []*cato_models.ClientConnectivitySourceRangeInput {
	out := []*cato_models.ClientConnectivitySourceRangeInput{}
	if list.IsNull() || list.IsUnknown() {
		return out
	}
	var rows []ClientConnectivitySourceRangePlan
	diags.Append(list.ElementsAs(ctx, &rows, false)...)
	if diags.HasError() {
		return out
	}
	for _, row := range rows {
		gir, ok := refinput.FromObject(ctx, row.GlobalIPRange, diags, "global IP range", refinput.GlobalIPRange)
		if !ok {
			continue
		}
		out = append(out, &cato_models.ClientConnectivitySourceRangeInput{GlobalIPRange: gir})
	}
	return out
}

func ccOrigins(ctx context.Context, list types.List, diags *diag.Diagnostics) []cato_models.ClientConnectivityOriginEnum {
	out := []cato_models.ClientConnectivityOriginEnum{}
	for _, s := range refinput.StringsFromList(ctx, list, diags) {
		out = append(out, cato_models.ClientConnectivityOriginEnum(s))
	}
	return out
}

func ccPlatforms(ctx context.Context, list types.List, diags *diag.Diagnostics) []cato_models.OperatingSystem {
	out := []cato_models.OperatingSystem{}
	for _, s := range refinput.StringsFromList(ctx, list, diags) {
		out = append(out, cato_models.OperatingSystem(s))
	}
	return out
}

func hydrateClientConnectivityAddRuleInput(
	ctx context.Context,
	plan ClientConnectivityRule,
) (cato_models.ClientConnectivityAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.ClientConnectivityAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, ccAsOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := ClientConnectivityRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, ccAsOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.ClientConnectivityAddRuleDataInput{
		Name:             rule.Name.ValueString(),
		Description:      rule.Description.ValueString(),
		Enabled:          rule.Enabled.ValueBool(),
		Action:           cato_models.ClientConnectivityActionEnum(rule.Action.ValueString()),
		ConfidenceLevel:  cato_models.ClientConnectivityConfidenceLevelEnum(rule.ConfidenceLevel.ValueString()),
		ConnectionOrigin: ccOrigins(ctx, rule.ConnectionOrigin, &diags),
		Country:          []*cato_models.CountryRefInput{},
		Device:           []*cato_models.DeviceProfileRefInput{},
		Platform:         ccPlatforms(ctx, rule.Platform, &diags),
		Source:           ccSourceInput(ctx, rule.Source, &diags),
		SourceRange:      ccSourceRanges(ctx, rule.SourceRange, &diags),
	}
	if c := refinput.FromSet(ctx, rule.Country, &diags, "country", refinput.Country); c != nil {
		data.Country = c
	}
	if d := refinput.FromSet(ctx, rule.Device, &diags, "device profile", refinput.DeviceProfile); d != nil {
		data.Device = d
	}

	out.Rule = data
	return out, diags
}

func hydrateClientConnectivityUpdateRuleInput(
	ctx context.Context,
	plan ClientConnectivityRule,
) (cato_models.ClientConnectivityUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := ClientConnectivityRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, ccAsOpts)...)
	if diags.HasError() {
		return cato_models.ClientConnectivityUpdateRuleInput{}, diags
	}

	src := ccSourceInput(ctx, rule.Source, &diags)
	upd := cato_models.ClientConnectivityUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.ClientConnectivityUpdateRuleDataInput{
			Name:             rule.Name.ValueStringPointer(),
			Description:      rule.Description.ValueStringPointer(),
			Enabled:          rule.Enabled.ValueBoolPointer(),
			Action:           (*cato_models.ClientConnectivityActionEnum)(rule.Action.ValueStringPointer()),
			ConfidenceLevel:  (*cato_models.ClientConnectivityConfidenceLevelEnum)(rule.ConfidenceLevel.ValueStringPointer()),
			ConnectionOrigin: ccOrigins(ctx, rule.ConnectionOrigin, &diags),
			Country:          []*cato_models.CountryRefInput{},
			Device:           []*cato_models.DeviceProfileRefInput{},
			Platform:         ccPlatforms(ctx, rule.Platform, &diags),
			Source: &cato_models.ClientConnectivitySourceUpdateInput{
				User:       src.User,
				UsersGroup: src.UsersGroup,
			},
			SourceRange: ccSourceRanges(ctx, rule.SourceRange, &diags),
		},
	}
	if c := refinput.FromSet(ctx, rule.Country, &diags, "country", refinput.Country); c != nil {
		upd.Rule.Country = c
	}
	if d := refinput.FromSet(ctx, rule.Device, &diags, "device profile", refinput.DeviceProfile); d != nil {
		upd.Rule.Device = d
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *ClientConnectivityRuleRulePlan,
	_ ClientConnectivityRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, ClientConnectivityRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *ClientConnectivityRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	_ ClientConnectivityRuleRulePlan,
	priorHydrated *ClientConnectivityRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, ClientConnectivityRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *ClientConnectivityRule,
	cur *cato_go_sdk.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule,
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
	plan ClientConnectivityRule,
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
	_ ClientConnectivityRuleRulePlan,
	priorRulePlan *ClientConnectivityRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, ClientConnectivityRuleRuleAttrTypes, rulePlan)
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
	rule ClientConnectivityRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt ClientConnectivityRule) {
	resultSt = ClientConnectivityRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule ClientConnectivityRuleRulePlan) (resultRm cato_models.ClientConnectivityRemoveRuleInput) {
	resultRm = cato_models.ClientConnectivityRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateClientConnectivityRuleStateFromClient maps a policy read rule into Terraform nested objects.
//
//nolint:gocyclo,funlen // Hydrate all rule fields together.
func hydrateClientConnectivityRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule,
) (ClientConnectivityRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := ClientConnectivityRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.ConnectionOrigin = types.ListNull(types.StringType)
	out.Country = types.SetNull(idname.RefObjectType)
	out.Device = types.SetNull(idname.RefObjectType)
	out.Platform = types.ListNull(types.StringType)
	out.Source = types.ObjectNull(ClientConnectivitySourceAttrTypes)
	out.SourceRange = types.ListNull(types.ObjectType{AttrTypes: ClientConnectivitySourceRangeAttrTypes})

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
	if r.GetConfidenceLevel() != nil {
		out.ConfidenceLevel = types.StringValue(string(*r.GetConfidenceLevel()))
	}

	if co := r.GetConnectionOrigin(); len(co) > 0 {
		vals := make([]string, 0, len(co))
		for _, v := range co {
			vals = append(vals, string(v))
		}
		out.ConnectionOrigin = convert.ParseList(ctx, types.StringType, vals, "rule.connection_origin")
	}
	if c := r.GetCountry(); len(c) > 0 {
		out.Country = oldidname.ParseNameIDList(ctx, c, "rule.country")
	}
	if d := r.GetDevice(); len(d) > 0 {
		out.Device = oldidname.ParseNameIDList(ctx, d, "rule.device")
	}
	if p := r.GetPlatform(); len(p) > 0 {
		plats := make([]string, 0, len(p))
		for _, os := range p {
			plats = append(plats, string(os))
		}
		out.Platform = convert.ParseList(ctx, types.StringType, plats, "rule.platform")
	}

	if src := r.GetSource(); src != nil && (len(src.GetUser()) > 0 || len(src.GetUsersGroup()) > 0) {
		srcObj, d := types.ObjectValue(ClientConnectivitySourceAttrTypes, map[string]attr.Value{
			"user":        oldidname.ParseNameIDList(ctx, src.GetUser(), "rule.source.user"),
			"users_group": oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "rule.source.users_group"),
		})
		diags.Append(d...)
		out.Source = srcObj
	}

	if srs := r.GetSourceRange(); len(srs) > 0 {
		elems := make([]attr.Value, 0, len(srs))
		for _, sr := range srs {
			if sr == nil {
				continue
			}
			girObj := types.ObjectNull(idname.RefAttrTypes)
			if g := sr.GetGlobalIPRange(); g != nil && (g.GetID() != "" || g.GetName() != "") {
				o, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
					"id":   types.StringValue(g.GetID()),
					"name": types.StringValue(g.GetName()),
				})
				diags.Append(d...)
				girObj = o
			}
			o, d := types.ObjectValue(ClientConnectivitySourceRangeAttrTypes, map[string]attr.Value{
				"global_ip_range": girObj,
			})
			diags.Append(d...)
			elems = append(elems, o)
		}
		lst, d := types.ListValue(types.ObjectType{AttrTypes: ClientConnectivitySourceRangeAttrTypes}, elems)
		diags.Append(d...)
		out.SourceRange = lst
	}

	return out, diags
}
