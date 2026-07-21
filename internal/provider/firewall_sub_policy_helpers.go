package provider

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// subPolicyScopeRuleType is the ruleType value the API assigns to the scope rule
// that a sub-policy is created with.
const subPolicyScopeRuleType = cato_models.PolicyRuleTypeEnumSubPolicyScope

// ifwSubPolicyInfo returns the sub-policy info block for the given sub-policy id,
// or nil when it is not present in the policy snapshot.
func ifwSubPolicyInfo(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies_Policy {
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

// wanSubPolicyInfo returns the sub-policy info block for the given sub-policy id.
func wanSubPolicyInfo(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy {
	if body == nil || subID == "" {
		return nil
	}
	for _, sp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetSubPolicies() {
		if sp.GetPolicy().GetID() == subID {
			return sp.GetPolicy()
		}
	}
	return nil
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

// wanScopeRule returns the SUB_POLICY_SCOPE rule owned by the given sub-policy.
func wanScopeRule(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule {
	if body == nil || subID == "" {
		return nil
	}
	for _, rp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetRules() {
		rt := rp.GetRuleType()
		sp := rp.GetSubPolicy()
		if rt != nil && *rt == subPolicyScopeRuleType && sp != nil && sp.GetID() == subID {
			return rp.GetRule()
		}
	}
	return nil
}

// ifwSubPolicyCleanupRuleID returns the id of the auto-created cleanup rule for
// the given sub-policy. The cleanup rule is a POLICY_RULE owned by the
// sub-policy named "<sub-policy name> - Cleanup Rule"; it always exists and is
// always last, making it a stable anchor for inserting rules into the
// sub-policy. Falls back to any non-scope rule owned by the sub-policy.
func ifwSubPolicyCleanupRuleID(body *cato_go_sdk.Policy, subID string) string {
	info := ifwSubPolicyInfo(body, subID)
	if info == nil {
		return ""
	}
	cleanupName := info.GetName() + " - Cleanup Rule"
	fallback := ""
	for _, rp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetRules() {
		sp := rp.GetSubPolicy()
		rt := rp.GetRuleType()
		if sp == nil || sp.GetID() != subID || rt == nil || *rt != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}
		if rp.GetRule().GetName() == cleanupName {
			return rp.GetRule().GetID()
		}
		if fallback == "" {
			fallback = rp.GetRule().GetID()
		}
	}
	return fallback
}

// wanSubPolicyCleanupRuleID mirrors ifwSubPolicyCleanupRuleID for WAN.
func wanSubPolicyCleanupRuleID(body *cato_go_sdk.Policy, subID string) string {
	info := wanSubPolicyInfo(body, subID)
	if info == nil {
		return ""
	}
	cleanupName := info.GetName() + " - Cleanup Rule"
	fallback := ""
	for _, rp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetRules() {
		sp := rp.GetSubPolicy()
		rt := rp.GetRuleType()
		if sp == nil || sp.GetID() != subID || rt == nil || *rt != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}
		if rp.GetRule().GetName() == cleanupName {
			return rp.GetRule().GetID()
		}
		if fallback == "" {
			fallback = rp.GetRule().GetID()
		}
	}
	return fallback
}

// objectRefByID builds an ObjectRefBy=ID reference input.
func objectRefByID(id string) *cato_models.InternetFirewallPolicyRefInput {
	return &cato_models.InternetFirewallPolicyRefInput{By: cato_models.ObjectRefByID, Input: id}
}

// wanObjectRefByID builds an ObjectRefBy=ID reference input for WAN.
func wanObjectRefByID(id string) *cato_models.WanFirewallPolicyRefInput {
	return &cato_models.WanFirewallPolicyRefInput{By: cato_models.ObjectRefByID, Input: id}
}
