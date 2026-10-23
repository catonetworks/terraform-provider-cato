package ifsubpolicy

import (
	"context"
	"fmt"
	"maps"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	ifrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
)

// ifwSubPolicyInfo returns the sub-policy info block for the given sub-policy id,
// or nil when it is not present in the policy snapshot.
func ifwSubPolicyInfo(
	body *cato_go_sdk.Policy,
	subID string,
) *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies_Policy {
	if body == nil || subID == "" {
		return nil
	}
	for _, sp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetSubPolicies() {
		if sp.GetPolicy().GetID() == subID {
			return sp.GetPolicy()
		}
	}
	return nil
}

// derefStr safely dereferences an optional string pointer.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ifwScopeRule returns the SUB_POLICY_SCOPE rule owned by the given sub-policy,
// or nil when not found.
func ifwScopeRule(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule {
	if body == nil || subID == "" {
		return nil
	}
	for _, rp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetRules() {
		rt := rp.GetRuleType()
		sp := rp.GetSubPolicy()
		if rt != nil && *rt == subPolicyScopeRuleType && sp != nil && sp.GetID() == subID {
			return rp.GetRule()
		}
	}
	return nil
}

// ifwSubPolicyIDByName returns the id of the sub-policy with the given name that
// is not present in the exclude set. It is used to identify a freshly created
// sub-policy by diffing the sub-policy list before and after creation.
func ifwSubPolicyIDByName(body *cato_go_sdk.Policy, name string, exclude map[string]struct{}) string {
	if body == nil {
		return ""
	}
	for _, sp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetSubPolicies() {
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

// ifwSubPolicyIDs returns the set of sub-policy ids currently in the snapshot.
func ifwSubPolicyIDs(body *cato_go_sdk.Policy) map[string]struct{} {
	ids := map[string]struct{}{}
	if body == nil {
		return ids
	}
	for _, sp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetSubPolicies() {
		ids[sp.GetPolicy().GetID()] = struct{}{}
	}
	return ids
}

// objectRefByID builds an ObjectRefBy=ID reference input.
func objectRefByID(id string) *cato_models.InternetFirewallPolicyRefInput {
	return &cato_models.InternetFirewallPolicyRefInput{By: cato_models.ObjectRefByID, Input: id}
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	plan InternetFirewallSubPolicy,
) (resultCreateScope basetypes.ObjectValue, resultDiags diag.Diagnostics) {
	scopeAttrs := maps.Clone(plan.Scope.Attributes())
	scopeAttrs["action"] = types.StringValue(string(cato_models.InternetFirewallActionEnumAllow))
	scopeAttrs["name"] = plan.Name
	scopeAttrs["description"] = plan.Description
	resultCreateScope, resultDiags = types.ObjectValue(ifrule.InternetFirewallRuleRuleAttrTypes, scopeAttrs)
	return resultCreateScope, resultDiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	plan InternetFirewallSubPolicy,
	hydrated ifrule.HydrateIfwAPITypes,
	before *cato_go_sdk.Policy,
) (resultExisting map[string]struct{}, resultAddInput cato_models.InternetFirewallAddSubPolicyInput) {
	resultExisting = ifwSubPolicyIDs(before)

	resultAddInput = cato_models.InternetFirewallAddSubPolicyInput{
		At: hydrated.Create.At,
		Policy: &cato_models.InternetFirewallAddSubPolicyDataInput{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		},
		Scope: hydrated.Create.Rule,
	}
	return resultExisting, resultAddInput
}

// internetFirewallSubPolicyStateFromPolicy finds the created sub-policy and encodes its published scope rule.
func internetFirewallSubPolicyStateFromPolicy(
	ctx context.Context,
	plan *InternetFirewallSubPolicy,
	priorDiags *diag.Diagnostics,
	scopeRule ifrule.InternetFirewallRule,
	existing map[string]struct{},
	after *cato_go_sdk.Policy,
	diagnostics *diag.Diagnostics,
) (resultScopeObj basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	subID := ifwSubPolicyIDByName(after, plan.Name.ValueString(), existing)
	if subID == "" {
		diagnostics.AddError("Sub-Policy Not Found", "Created Internet Firewall sub-policy could not be located after publish.")
		return resultScopeObj
	}
	scope := ifwScopeRule(after, subID)
	if scope == nil {
		diagnostics.AddError("Scope Rule Not Found", fmt.Sprintf("No SUB_POLICY_SCOPE rule found for sub-policy %s.", subID))
		return resultScopeObj
	}

	plan.ID = types.StringValue(subID)
	plan.ScopeRuleID = types.StringValue(scope.GetID())
	scopeState := ifrule.HydrateIfwRuleState(ctx, scopeRule, scope)
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, diags = types.ObjectValueFrom(ctx, ifrule.InternetFirewallRuleRuleAttrTypes, scopeState)
	return resultScopeObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state *InternetFirewallSubPolicy,
	info *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies_Policy,
	scope *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule,
) (resultScopeObj basetypes.ObjectValue, resultDiags diag.Diagnostics) {
	state.Name = types.StringValue(info.GetName())
	if info.GetDescription() == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(info.GetDescription())
	}
	state.ScopeRuleID = types.StringValue(scope.GetID())

	scopeRule := ifrule.InternetFirewallRule{Rule: state.Scope, At: state.At}
	scopeState := ifrule.HydrateIfwRuleState(ctx, scopeRule, scope)
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, resultDiags = types.ObjectValueFrom(ctx, ifrule.InternetFirewallRuleRuleAttrTypes, scopeState)
	return resultScopeObj, resultDiags
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan InternetFirewallSubPolicy,
	scopeRule ifrule.InternetFirewallRule,
	priorDiags *diag.Diagnostics,
	body *cato_go_sdk.Policy,
	diagnostics *diag.Diagnostics,
) (resultScopeObj basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	scope := ifwScopeRule(body, plan.ID.ValueString())
	if scope == nil {
		diagnostics.AddError("Scope Rule Not Found", "Sub-policy scope rule not found after update.")
		return resultScopeObj
	}
	scopeState := ifrule.HydrateIfwRuleState(ctx, scopeRule, scope)
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, diags = types.ObjectValueFrom(ctx, ifrule.InternetFirewallRuleRuleAttrTypes, scopeState)
	return resultScopeObj
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	state InternetFirewallSubPolicy,
) (resultRemoveInput cato_models.InternetFirewallRemoveSubPolicyInput) {
	resultRemoveInput = cato_models.InternetFirewallRemoveSubPolicyInput{Ref: objectRefByID(state.ID.ValueString())}
	return resultRemoveInput
}
