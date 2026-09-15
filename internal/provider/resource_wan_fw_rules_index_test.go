package provider

import (
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestWanSubPolicyOrderingDataFromPolicy(t *testing.T) {
	t.Parallel()

	policy := &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					SubPolicies: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies{
						{
							Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy{
								ID:   "sub-policy-id",
								Name: "Sub-policy",
							},
						},
					},
					Rules: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{
						{
							SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{
								ID:   "sub-policy-id",
								Name: "Sub-policy",
							},
							Rule: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{
								ID:    "rule-a-id",
								Name:  "Rule A",
								Index: 1,
								Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule_Section{
									ID:   "section-id",
									Name: "Section",
								},
							},
						},
					},
				},
			},
		},
	}

	got := wanSubPolicyOrderingDataFromPolicy(policy)

	require.Equal(t, wanSubPolicyOrderingData{
		ID:   "sub-policy-id",
		Name: "Sub-policy",
		Sections: []BulkPolicySectionRef{
			{ID: "section-id", Name: "Section"},
		},
		Rules: []BulkPolicyRuleRow{
			{
				SectionID:   "section-id",
				SectionName: "Section",
				RuleID:      "rule-a-id",
				RuleName:    "Rule A",
				Index:       1,
				IsSystem:    false,
			},
		},
	}, got["Sub-policy"])
}
