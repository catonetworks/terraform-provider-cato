package lfsubpolicy

import (
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// The cleanup rule is ALWAYS the sub-policy's LAST rule (max index), with the
// SUB_POLICY_SCOPE row and other sub-policies' rules excluded.
func TestLanCleanupRulePositionalSelection(t *testing.T) {
	scopeType := cato_models.PolicyRuleTypeEnumSubPolicyScope
	ruleType := cato_models.PolicyRuleTypeEnumPolicyRule
	mk := func(id string, idx int64, subID string, rt cato_models.PolicyRuleTypeEnum) *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules {
		return &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules{
			RuleType:  rt,
			SubPolicy: &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_SubPolicy{ID: subID},
			Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
				ID: id, Index: idx, Name: "r-" + id,
			},
		}
	}
	policy := &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy{
		Rules: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules{
			mk("scope", 1, "sub-1", scopeType),
			mk("mid", 2, "sub-1", ruleType),
			mk("last", 3, "sub-1", ruleType),
			mk("other", 9, "sub-2", ruleType),
		},
	}
	got := lanCleanupRule(policy, "sub-1")
	if got == nil || got.GetID() != "last" {
		t.Fatalf("lanCleanupRule = %+v, want rule id last", got)
	}
	if lanCleanupRule(policy, "missing") != nil {
		t.Fatal("expected nil for unknown sub-policy")
	}
}
