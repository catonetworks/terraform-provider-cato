package provider

import (
	"context"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHydrateLanFw(t *testing.T) {
	var ctx = context.Background()
	var diags diag.Diagnostics
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{},
	}
	plan := lPMockClient.createPlan(lanPolicyPlans["default"])
	(res.catov2Client).(*lanPolicyMockClient).policy = mockLanPolicy["default"]
	newState, indexMap := res.hydrateLanFwRulesIndex(ctx, plan, &diags)
	_, _ = newState, indexMap
}

type lanPolicyMockClient struct {
	policy *cato_go_sdk.PolicySocketLanPolicy
}

var lPMockClient lanPolicyMockClient

func (m *lanPolicyMockClient) PolicySocketLanPolicy(_ context.Context, _ string,
	_ *cato_models.SocketLanPolicyInput, _ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPolicy, error) {
	return m.policy, nil
}

type lanPolicyPlanItem struct {
	sectionData map[string]lanPolicySectionData
	ruleData    map[string]lanPolocyRuleData
}
type lanPolicySectionData struct {
	id            string
	subPolicyName string
	sectionIndex  int
}
type lanPolocyRuleData struct {
	id             string
	ruleType       string
	sectionName    string
	indexInSection int
}

//	   section_1
//	   		rule_1.1
//	   		rule_1.2
//	   		sub_policy_1.1
//	   			section_1.1.1
//	   			section_1.1.2
//						rule_1.1.2.1
//						rule_1.1.2.2
//						rule_1.1.2.3
//	   			section_1.1.3
//	   		sub_policy_1.2
//	   		sub_policy_1.3
//	   		rule_1.3
//	   section_2
//	   section_3
var lanPolicyPlans = map[string]lanPolicyPlanItem{
	"default": {
		sectionData: map[string]lanPolicySectionData{
			"section_1":     {id: "s1", subPolicyName: "", sectionIndex: 1},
			"section_2":     {id: "s2", subPolicyName: "", sectionIndex: 2},
			"section_3":     {id: "s3", subPolicyName: "", sectionIndex: 3},
			"section_1.1.1": {id: "s1.1.1", subPolicyName: "sub_policy_1.1", sectionIndex: 1},
			"section_1.1.2": {id: "s1.1.2", subPolicyName: "sub_policy_1.1", sectionIndex: 2},
			"section_1.1.3": {id: "s1.1.3", subPolicyName: "sub_policy_1.1", sectionIndex: 3},
		},
		ruleData: map[string]lanPolocyRuleData{
			"rule_1.1":       {id: "r1.1", ruleType: "POLICY_RULE", sectionName: "section_1", indexInSection: 1},
			"rule_1.2":       {id: "r1.2", ruleType: "POLICY_RULE", sectionName: "section_1", indexInSection: 2},
			"sub_policy_1.1": {id: "sp1.1", ruleType: "SUB_POLICY_SCOPE", sectionName: "section_1", indexInSection: 3},
			"sub_policy_1.2": {id: "sp1.2", ruleType: "SUB_POLICY_SCOPE", sectionName: "section_1", indexInSection: 4},
			"sub_policy_1.3": {id: "sp1.3", ruleType: "SUB_POLICY_SCOPE", sectionName: "section_1", indexInSection: 5},
			"rule_1.3":       {id: "r1.3", ruleType: "POLICY_RULE", sectionName: "section_1", indexInSection: 6},
			"rule_1.1.2.1":   {id: "r1.1.2.1", ruleType: "POLICY_RULE", sectionName: "section_1.1.2", indexInSection: 1},
			"rule_1.1.2.2":   {id: "r1.1.2.2", ruleType: "POLICY_RULE", sectionName: "section_1.1.2", indexInSection: 2},
			"rule_1.1.2.3":   {id: "r1.1.2.3", ruleType: "POLICY_RULE", sectionName: "section_1.1.2", indexInSection: 3},
		},
	},
}

func (m *lanPolicyMockClient) createPlan(p lanPolicyPlanItem) *LanFwRulesIndex {
	ctx := context.Background()
	sectionObjects := make(map[string]types.Object, len(p.sectionData))
	for sectionName, section := range p.sectionData {
		sectionObject, diags := types.ObjectValueFrom(ctx, LanFwSectionDataTypes, LanFwSectionData{
			ID:            types.StringValue(section.id),
			SectionIndex:  types.Int64Value(int64(section.sectionIndex)),
			SubPolicyName: types.StringValue(section.subPolicyName),
		})
		if diags.HasError() {
			return &LanFwRulesIndex{}
		}
		sectionObjects[sectionName] = sectionObject
	}
	sectionData, sectionDiags := types.MapValueFrom(
		ctx,
		types.ObjectType{AttrTypes: LanFwSectionDataTypes},
		sectionObjects,
	)
	if sectionDiags.HasError() {
		return &LanFwRulesIndex{}
	}

	ruleObjects := make(map[string]types.Object, len(p.ruleData))
	for ruleName, rule := range p.ruleData {
		ruleObject, diags := types.ObjectValueFrom(ctx, LanFwRuleDataTypes, LanFwRuleData{
			ID:             types.StringValue(rule.id),
			RuleType:       types.StringValue(rule.ruleType),
			SectionName:    types.StringValue(rule.sectionName),
			IndexInSection: types.Int64Value(int64(rule.indexInSection)),
		})
		if diags.HasError() {
			return &LanFwRulesIndex{}
		}
		ruleObjects[ruleName] = ruleObject
	}
	ruleData, ruleDiags := types.MapValueFrom(
		ctx,
		types.ObjectType{AttrTypes: LanFwRuleDataTypes},
		ruleObjects,
	)
	if ruleDiags.HasError() {
		return &LanFwRulesIndex{}
	}

	return &LanFwRulesIndex{
		SectionData: sectionData,
		RuleData:    ruleData,
	}
}

var mockLanPolicy = map[string]*cato_go_sdk.PolicySocketLanPolicy{
	"default": {
		Policy: &cato_go_sdk.PolicySocketLanPolicy_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan{
				Policy: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy{
					Enabled: true,
					Rules: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules{
						{
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.2",
								Name: "rule_1.2",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1",
									Name: "section_1",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.1",
								Name: "rule_1.1",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1",
									Name: "section_1",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumSubPolicyScope,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "sub1.1",
								Name: "sub_policy_1.1",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1",
									Name: "section_1",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumSubPolicyScope,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "sub1.2",
								Name: "sub_policy_1.2",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1",
									Name: "section_1",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumSubPolicyScope,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "sub1.3",
								Name: "sub_policy_1.3",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1",
									Name: "section_1",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.1.2.1",
								Name: "rule_1.1.2.1",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1.1.2",
									Name: "section_1.1.2",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.1.2.3",
								Name: "rule_1.1.2.3",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1.1.2",
									Name: "section_1.1.2",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.3",
								Name: "rule_1.3",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1",
									Name: "section_1",
								},
							},
						},
						{
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.1.2.2",
								Name: "rule_1.1.2.2",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1.1.2",
									Name: "section_1.1.2",
								},
							},
						},
					},
					Sections: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections{
						{
							Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
								ID:          "s1",
								Name:        "section_1",
								SubPolicyID: nil,
							},
						},
						{
							Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
								ID:          "s3",
								Name:        "section_3",
								SubPolicyID: nil,
							},
						},
						{
							Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
								ID:          "s1.1.3",
								Name:        "section_1.1.3",
								SubPolicyID: ptr("sub1.1"),
							},
						},
						{
							Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
								ID:          "s1.1.1",
								Name:        "section_1.1.1",
								SubPolicyID: ptr("sub1.1"),
							},
						},
						{
							Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
								ID:          "s1.1.2",
								Name:        "section_1.1.2",
								SubPolicyID: ptr("sub1.1"),
							},
						},
						{
							Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
								ID:          "s2",
								Name:        "section_2",
								SubPolicyID: nil,
							},
						},
					},
					SubPolicies: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies{
						{
							Policy: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies_Policy{
								ID:   "sub1.1",
								Name: "sub_policy_1.1",
							},
						},
						{
							Policy: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies_Policy{
								ID:   "sub1.2",
								Name: "sub_policy_1.2",
							},
						},
					},
				},
			},
		},
	},
}
