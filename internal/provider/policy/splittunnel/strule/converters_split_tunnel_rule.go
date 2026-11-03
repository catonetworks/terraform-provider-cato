package strule

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

var stAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func stPlatforms(ctx context.Context, list types.List, diags *diag.Diagnostics) []cato_models.OperatingSystem {
	out := []cato_models.OperatingSystem{}
	for _, s := range refinput.StringsFromList(ctx, list, diags) {
		out = append(out, cato_models.OperatingSystem(s))
	}
	return out
}

func stDestinationInput(
	ctx context.Context,
	destObj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SplitTunnelDestinationInput {
	out := &cato_models.SplitTunnelDestinationInput{
		Application:   []*cato_models.ApplicationRefInput{},
		Domain:        []string{},
		Fqdn:          []string{},
		GlobalIPRange: []*cato_models.GlobalIPRangeRefInput{},
	}
	if destObj.IsNull() || destObj.IsUnknown() {
		return out
	}
	dst := SplitTunnelDestinationPlan{}
	diags.Append(destObj.As(ctx, &dst, stAsOpts)...)
	if diags.HasError() {
		return out
	}
	if a := refinput.FromSet(ctx, dst.Application, diags, "application", refinput.Application); a != nil {
		out.Application = a
	}
	if d := refinput.StringsFromList(ctx, dst.Domain, diags); d != nil {
		out.Domain = d
	}
	if f := refinput.StringsFromList(ctx, dst.Fqdn, diags); f != nil {
		out.Fqdn = f
	}
	if g := refinput.FromSet(ctx, dst.GlobalIPRange, diags, "global IP range", refinput.GlobalIPRange); g != nil {
		out.GlobalIPRange = g
	}
	return out
}

func stSourceInput(
	ctx context.Context,
	sourceObj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SplitTunnelSourceInput {
	out := &cato_models.SplitTunnelSourceInput{
		User:       []*cato_models.UserRefInput{},
		UsersGroup: []*cato_models.UsersGroupRefInput{},
	}
	if sourceObj.IsNull() || sourceObj.IsUnknown() {
		return out
	}
	src := SplitTunnelSourcePlan{}
	diags.Append(sourceObj.As(ctx, &src, stAsOpts)...)
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

func stSourceNetworkInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SplitTunnelSourceNetworkInput {
	out := &cato_models.SplitTunnelSourceNetworkInput{
		SourceNetworkType: cato_models.SplitTunnelSourceNetworkTypeEnumAny,
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	sn := SplitTunnelSourceNetworkPlan{}
	diags.Append(obj.As(ctx, &sn, stAsOpts)...)
	if !sn.SourceNetworkType.IsNull() && !sn.SourceNetworkType.IsUnknown() && sn.SourceNetworkType.ValueString() != "" {
		out.SourceNetworkType = cato_models.SplitTunnelSourceNetworkTypeEnum(sn.SourceNetworkType.ValueString())
	}
	return out
}

func stDNSExclusionInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SplitTunnelDNSExclusionInput {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	dx := SplitTunnelDnsExclusionPlan{}
	diags.Append(obj.As(ctx, &dx, stAsOpts)...)
	if diags.HasError() {
		return nil
	}
	domains := refinput.StringsFromList(ctx, dx.Domain, diags)
	if domains == nil {
		domains = []string{}
	}
	return &cato_models.SplitTunnelDNSExclusionInput{Domain: domains}
}

func hydrateSplitTunnelAddRuleInput(
	ctx context.Context,
	plan SplitTunnelRule,
) (cato_models.SplitTunnelAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.SplitTunnelAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, stAsOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := SplitTunnelRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, stAsOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.SplitTunnelAddRuleDataInput{
		Name:            rule.Name.ValueString(),
		Description:     rule.Description.ValueString(),
		Enabled:         rule.Enabled.ValueBool(),
		Action:          cato_models.SplitTunnelActionEnum(rule.Action.ValueString()),
		RoutingPriority: cato_models.SplitTunnelRoutingPriorityEnum(rule.RoutingPriority.ValueString()),
		Country:         []*cato_models.CountryRefInput{},
		Platform:        stPlatforms(ctx, rule.Platform, &diags),
		Destination:     stDestinationInput(ctx, rule.Destination, &diags),
		Source:          stSourceInput(ctx, rule.Source, &diags),
		SourceNetwork:   stSourceNetworkInput(ctx, rule.SourceNetwork, &diags),
	}
	if !rule.Coverage.IsNull() && !rule.Coverage.IsUnknown() && rule.Coverage.ValueString() != "" {
		cov := cato_models.SplitTunnelCoverageEnum(rule.Coverage.ValueString())
		data.Coverage = &cov
	}
	if c := refinput.FromSet(ctx, rule.Country, &diags, "country", refinput.Country); c != nil {
		data.Country = c
	}
	data.DNSExclusion = stDNSExclusionInput(ctx, rule.DnsExclusion, &diags)

	out.Rule = data
	return out, diags
}

func hydrateSplitTunnelUpdateRuleInput(
	ctx context.Context,
	plan SplitTunnelRule,
) (cato_models.SplitTunnelUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := SplitTunnelRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, stAsOpts)...)
	if diags.HasError() {
		return cato_models.SplitTunnelUpdateRuleInput{}, diags
	}

	dst := stDestinationInput(ctx, rule.Destination, &diags)
	src := stSourceInput(ctx, rule.Source, &diags)
	sn := stSourceNetworkInput(ctx, rule.SourceNetwork, &diags)

	upd := cato_models.SplitTunnelUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.SplitTunnelUpdateRuleDataInput{
			Name:            rule.Name.ValueStringPointer(),
			Description:     rule.Description.ValueStringPointer(),
			Enabled:         rule.Enabled.ValueBoolPointer(),
			Action:          (*cato_models.SplitTunnelActionEnum)(rule.Action.ValueStringPointer()),
			RoutingPriority: (*cato_models.SplitTunnelRoutingPriorityEnum)(rule.RoutingPriority.ValueStringPointer()),
			Country:         []*cato_models.CountryRefInput{},
			Platform:        stPlatforms(ctx, rule.Platform, &diags),
			Destination: &cato_models.SplitTunnelDestinationUpdateInput{
				Application:   dst.Application,
				Domain:        dst.Domain,
				Fqdn:          dst.Fqdn,
				GlobalIPRange: dst.GlobalIPRange,
			},
			Source: &cato_models.SplitTunnelSourceUpdateInput{
				User:       src.User,
				UsersGroup: src.UsersGroup,
			},
			SourceNetwork: &cato_models.SplitTunnelSourceNetworkUpdateInput{
				SourceNetworkType: &sn.SourceNetworkType,
			},
		},
	}
	if !rule.Coverage.IsNull() && !rule.Coverage.IsUnknown() && rule.Coverage.ValueString() != "" {
		upd.Rule.Coverage = (*cato_models.SplitTunnelCoverageEnum)(rule.Coverage.ValueStringPointer())
	}
	if c := refinput.FromSet(ctx, rule.Country, &diags, "country", refinput.Country); c != nil {
		upd.Rule.Country = c
	}
	if dx := stDNSExclusionInput(ctx, rule.DnsExclusion, &diags); dx != nil {
		upd.Rule.DNSExclusion = &cato_models.SplitTunnelDNSExclusionUpdateInput{Domain: dx.Domain}
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *SplitTunnelRuleRulePlan,
	_ SplitTunnelRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SplitTunnelRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *SplitTunnelRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	_ SplitTunnelRuleRulePlan,
	priorHydrated *SplitTunnelRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SplitTunnelRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *SplitTunnelRule,
	cur *cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule,
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
	plan SplitTunnelRule,
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
	_ SplitTunnelRuleRulePlan,
	priorRulePlan *SplitTunnelRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SplitTunnelRuleRuleAttrTypes, rulePlan)
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
	rule SplitTunnelRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt SplitTunnelRule) {
	resultSt = SplitTunnelRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule SplitTunnelRuleRulePlan) (resultRm cato_models.SplitTunnelRemoveRuleInput) {
	resultRm = cato_models.SplitTunnelRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateSplitTunnelRuleStateFromClient maps a policy read rule into Terraform nested objects.
//
//nolint:gocyclo
func hydrateSplitTunnelRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule,
) (SplitTunnelRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := SplitTunnelRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.Country = types.SetNull(idname.RefObjectType)
	out.Platform = types.ListNull(types.StringType)
	out.Destination = types.ObjectNull(SplitTunnelDestinationAttrTypes)
	out.DnsExclusion = types.ObjectNull(SplitTunnelDnsExclusionAttrTypes)
	out.Source = types.ObjectNull(SplitTunnelSourceAttrTypes)
	out.SourceNetwork = types.ObjectNull(SplitTunnelSourceNetworkAttrTypes)

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
	if r.GetCoverage() != nil {
		out.Coverage = types.StringValue(string(*r.GetCoverage()))
	} else {
		out.Coverage = types.StringNull()
	}
	if r.GetRoutingPriority() != nil {
		out.RoutingPriority = types.StringValue(string(*r.GetRoutingPriority()))
	}

	if c := r.GetCountry(); len(c) > 0 {
		out.Country = oldidname.ParseNameIDList(ctx, c, "rule.country")
	}
	if p := r.GetPlatform(); len(p) > 0 {
		plats := make([]string, 0, len(p))
		for _, os := range p {
			plats = append(plats, string(os))
		}
		out.Platform = convert.ParseList(ctx, types.StringType, plats, "rule.platform")
	}

	if dst := r.GetDestination(); dst != nil {
		dstObj, d := types.ObjectValue(SplitTunnelDestinationAttrTypes, map[string]attr.Value{
			"application":     oldidname.ParseNameIDList(ctx, dst.GetApplication(), "rule.destination.application"),
			"domain":          convert.ParseList(ctx, types.StringType, dst.GetDomain(), "rule.destination.domain"),
			"fqdn":            convert.ParseList(ctx, types.StringType, dst.GetFqdn(), "rule.destination.fqdn"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, dst.GetGlobalIPRange(), "rule.destination.global_ip_range"),
		})
		diags.Append(d...)
		out.Destination = dstObj
	}

	if dx := r.GetDNSExclusion(); dx != nil {
		dxObj, d := types.ObjectValue(SplitTunnelDnsExclusionAttrTypes, map[string]attr.Value{
			"domain": convert.ParseList(ctx, types.StringType, dx.GetDomain(), "rule.dns_exclusion.domain"),
		})
		diags.Append(d...)
		out.DnsExclusion = dxObj
	}

	if src := r.GetSource(); src != nil {
		srcObj, d := types.ObjectValue(SplitTunnelSourceAttrTypes, map[string]attr.Value{
			"user":        oldidname.ParseNameIDList(ctx, src.GetUser(), "rule.source.user"),
			"users_group": oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "rule.source.users_group"),
		})
		diags.Append(d...)
		out.Source = srcObj
	}

	if sn := r.GetSourceNetwork(); sn != nil && sn.GetSourceNetworkType() != nil {
		snObj, d := types.ObjectValue(SplitTunnelSourceNetworkAttrTypes, map[string]attr.Value{
			"source_network_type": types.StringValue(string(*sn.GetSourceNetworkType())),
		})
		diags.Append(d...)
		out.SourceNetwork = snObj
	}

	return out, diags
}
