package wfsubpolicy

import (
	"context"
	"fmt"
	"maps"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	wfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
)

// wanSubPolicyIDs returns the set of sub-policy ids currently in the snapshot.
func wanSubPolicyIDs(body *cato_go_sdk.Policy) map[string]struct{} {
	ids := map[string]struct{}{}
	if body == nil {
		return ids
	}
	for _, sp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetSubPolicies() {
		ids[sp.GetPolicy().GetID()] = struct{}{}
	}
	return ids
}

// wanSubPolicyIDByName returns the id of the sub-policy with the given name that
// is not present in the exclude set.
func wanSubPolicyIDByName(body *cato_go_sdk.Policy, name string, exclude map[string]struct{}) string {
	if body == nil {
		return ""
	}
	for _, sp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetSubPolicies() {
		info := sp.GetPolicy()
		if info.GetName() != name {
			continue
		}
		if _, skip := exclude[info.GetID()]; skip {
			continue
		}
		return info.GetID()
	}
	return ""
}

// derefStr safely dereferences an optional string pointer.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// wanScopeRule returns the SUB_POLICY_SCOPE rule owned by the given sub-policy.
func wanScopeRule(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule {
	if body == nil || subID == "" {
		return nil
	}
	for _, rp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetRules() {
		rt := rp.GetRuleType()
		sp := rp.GetSubPolicy()
		if rt != nil && *rt == cato_models.PolicyRuleTypeEnumSubPolicyScope && sp != nil && sp.GetID() == subID {
			return rp.GetRule()
		}
	}
	return nil
}

// wanObjectRefByID builds an ObjectRefBy=ID reference input for WAN.
func wanObjectRefByID(id string) *cato_models.WanFirewallPolicyRefInput {
	return &cato_models.WanFirewallPolicyRefInput{By: cato_models.ObjectRefByID, Input: id}
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(plan WanFirewallSubPolicy) (resultCreateScope basetypes.ObjectValue, resultDiags diag.Diagnostics) {
	scopeAttrs := maps.Clone(plan.Scope.Attributes())
	scopeAttrs["action"] = types.StringValue(string(cato_models.WanFirewallActionEnumAllow))
	scopeAttrs["name"] = plan.Name
	scopeAttrs["description"] = plan.Description
	resultCreateScope, resultDiags = types.ObjectValue(wfrule.WanFirewallRuleRuleAttrTypes, scopeAttrs)
	return resultCreateScope, resultDiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	plan WanFirewallSubPolicy,
	hydrated wfrule.HydrateWanAPITypes,
	before *cato_go_sdk.Policy,
) (resultExisting map[string]struct{}, resultAddInput cato_models.WanFirewallAddSubPolicyInput) {
	resultExisting = wanSubPolicyIDs(before)

	resultAddInput = cato_models.WanFirewallAddSubPolicyInput{
		At: hydrated.Create.At,
		Policy: &cato_models.WanFirewallAddSubPolicyDataInput{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		},
		Scope: hydrated.Create.Rule,
	}
	return resultExisting, resultAddInput
}

// convertCreateInput3 converts between the Terraform model and SDK values.
func convertCreateInput3(
	ctx context.Context,
	plan *WanFirewallSubPolicy,
	priorDiags *diag.Diagnostics,
	scopeRule wfrule.Rule,
	existing map[string]struct{},
	after *cato_go_sdk.Policy,
	diagnostics *diag.Diagnostics,
) (
	resultScope *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
	resultScopeState wfrule.PolicyPolicyWanFirewallPolicyRulesRule,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	subID := wanSubPolicyIDByName(after, plan.Name.ValueString(), existing)
	if subID == "" {
		diagnostics.AddError("Sub-Policy Not Found", "Created WAN Firewall sub-policy could not be located after publish.")
		return resultScope, resultScopeState
	}
	resultScope = wanScopeRule(after, subID)
	if resultScope == nil {
		diagnostics.AddError("Scope Rule Not Found", fmt.Sprintf("No SUB_POLICY_SCOPE rule found for sub-policy %s.", subID))
		return resultScope, resultScopeState
	}

	plan.ID = types.StringValue(subID)
	plan.ScopeRuleID = types.StringValue(resultScope.GetID())
	resultScopeState, diags = wfrule.HydrateWanRuleState(ctx, scopeRule, resultScope)
	return resultScope, resultScopeState
}

// convertCreateInput4 converts between the Terraform model and SDK values.
func convertCreateInput4(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	scope *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
	priorScopeState *wfrule.PolicyPolicyWanFirewallPolicyRulesRule,
) (resultScopeObj basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	scopeState := *priorScopeState
	defer func() { *priorScopeState = scopeState }()
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, diags = types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, scopeState)
	return resultScopeObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state *WanFirewallSubPolicy,
	info *cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy,
	scope *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
) (resultScopeState wfrule.PolicyPolicyWanFirewallPolicyRulesRule, resultDiags diag.Diagnostics) {
	state.Name = types.StringValue(info.GetName())
	if info.GetDescription() == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(info.GetDescription())
	}
	state.ScopeRuleID = types.StringValue(scope.GetID())

	scopeRule := wfrule.Rule{Rule: state.Scope, At: state.At}
	resultScopeState, resultDiags = wfrule.HydrateWanRuleState(ctx, scopeRule, scope)
	return resultScopeState, resultDiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	ctx context.Context,
	scope *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
	priorScopeState *wfrule.PolicyPolicyWanFirewallPolicyRulesRule,
	priorDiags *diag.Diagnostics,
) (resultScopeObj basetypes.ObjectValue) {
	scopeState := *priorScopeState
	defer func() { *priorScopeState = scopeState }()
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, diags = types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, scopeState)
	return resultScopeObj
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	scope *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
	priorScopeState *wfrule.PolicyPolicyWanFirewallPolicyRulesRule,
) (resultScopeObj basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	scopeState := *priorScopeState
	defer func() { *priorScopeState = scopeState }()
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, diags = types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, scopeState)
	return resultScopeObj
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state WanFirewallSubPolicy) (resultRemoveInput cato_models.WanFirewallRemoveSubPolicyInput) {
	resultRemoveInput = cato_models.WanFirewallRemoveSubPolicyInput{Ref: wanObjectRefByID(state.ID.ValueString())}
	return resultRemoveInput
}
