package wan

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// SubPolicyInfo returns the sub-policy info block for the given sub-policy id.
func SubPolicyInfo(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy {
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

// SubPolicyCleanupRuleID mirrors ifwSubPolicyCleanupRuleID for WAN.
func SubPolicyCleanupRuleID(body *cato_go_sdk.Policy, subID string) string {
	info := SubPolicyInfo(body, subID)
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
