package wan

import (
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestWanSubPolicyCleanupRuleID(t *testing.T) {
	subID, subName := "sub-w", "MyWanSub"
	rules := []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{
		{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{ID: "wuser-1", Name: "wan-user"},
		},
		{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{ID: "wcleanup-1", Name: subName + " - Cleanup Rule"},
		},
	}
	body := &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					SubPolicies: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies{
						{Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy{ID: subID, Name: subName}},
					},
					Rules: rules,
				},
			},
		},
	}

	if got := SubPolicyCleanupRuleID(body, subID); got != "wcleanup-1" {
		t.Fatalf("expected wcleanup-1, got %q", got)
	}
	if got := SubPolicyCleanupRuleID(body, "nope"); got != "" {
		t.Fatalf("expected empty for missing sub-policy, got %q", got)
	}
}
