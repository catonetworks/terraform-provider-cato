package ifsubpolicy

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
)

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
