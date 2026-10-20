package ifrule

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

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
