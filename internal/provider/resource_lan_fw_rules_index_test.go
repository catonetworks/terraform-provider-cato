package provider

import (
	"context"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
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
	require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
	require.NotNil(t, newState)
	require.NotNil(t, indexMap)

	var sectionData map[string]LanFwSectionData
	sectionDiags := newState.SectionData.ElementsAs(ctx, &sectionData, false)
	require.False(t, sectionDiags.HasError(), "unexpected section state diagnostics: %+v", sectionDiags)
	require.Equal(t, map[string]LanFwSectionData{
		"section_1":     {ID: types.StringValue("s1"), SectionIndex: types.Int64Value(1), SubPolicyName: types.StringValue("")},
		"section_3":     {ID: types.StringValue("s3"), SectionIndex: types.Int64Value(2), SubPolicyName: types.StringValue("")},
		"section_2":     {ID: types.StringValue("s2"), SectionIndex: types.Int64Value(3), SubPolicyName: types.StringValue("")},
		"section_1.1.3": {ID: types.StringValue("s1.1.3"), SectionIndex: types.Int64Value(1), SubPolicyName: types.StringValue("sub_policy_1.1")},
		"section_1.1.1": {ID: types.StringValue("s1.1.1"), SectionIndex: types.Int64Value(2), SubPolicyName: types.StringValue("sub_policy_1.1")},
		"section_1.1.2": {ID: types.StringValue("s1.1.2"), SectionIndex: types.Int64Value(3), SubPolicyName: types.StringValue("sub_policy_1.1")},
	}, sectionData)

	var networkRules map[string]LanNetworkRule
	networkDiags := newState.NetworkRules.ElementsAs(ctx, &networkRules, false)
	require.False(t, networkDiags.HasError(), "unexpected network rule state diagnostics: %+v", networkDiags)
	require.Equal(t, map[string]LanNetworkRule{
		"rule_1.2":       {ID: types.StringValue("r1.2"), RuleType: types.StringValue("POLICY_RULE"), SectionName: types.StringValue("section_1"), IndexInSection: types.Int64Value(1)},
		"rule_1.1":       {ID: types.StringValue("r1.1"), RuleType: types.StringValue("POLICY_RULE"), SectionName: types.StringValue("section_1"), IndexInSection: types.Int64Value(2)},
		"sub_policy_1.1": {ID: types.StringValue("sub1.1"), RuleType: types.StringValue("SUB_POLICY_SCOPE"), SectionName: types.StringValue("section_1"), IndexInSection: types.Int64Value(3)},
		"sub_policy_1.2": {ID: types.StringValue("sub1.2"), RuleType: types.StringValue("SUB_POLICY_SCOPE"), SectionName: types.StringValue("section_1"), IndexInSection: types.Int64Value(5)},
		"sub_policy_1.3": {ID: types.StringValue("sub1.3"), RuleType: types.StringValue("SUB_POLICY_SCOPE"), SectionName: types.StringValue("section_1"), IndexInSection: types.Int64Value(4)},
		"rule_1.3":       {ID: types.StringValue("r1.3"), RuleType: types.StringValue("POLICY_RULE"), SectionName: types.StringValue("section_1"), IndexInSection: types.Int64Value(6)},
		"rule_1.1.2.1":   {ID: types.StringValue("r1.1.2.1"), RuleType: types.StringValue("POLICY_RULE"), SectionName: types.StringValue("section_1.1.2"), IndexInSection: types.Int64Value(1)},
		"rule_1.1.2.3":   {ID: types.StringValue("r1.1.2.3"), RuleType: types.StringValue("POLICY_RULE"), SectionName: types.StringValue("section_1.1.2"), IndexInSection: types.Int64Value(2)},
		"rule_1.1.2.2":   {ID: types.StringValue("r1.1.2.2"), RuleType: types.StringValue("POLICY_RULE"), SectionName: types.StringValue("section_1.1.2"), IndexInSection: types.Int64Value(3)},
	}, networkRules)

	var firewallRules map[string]LanFirewallRule
	firewallDiags := newState.FirewallRules.ElementsAs(ctx, &firewallRules, false)
	require.False(t, firewallDiags.HasError(), "unexpected firewall rule state diagnostics: %+v", firewallDiags)
	require.Equal(t, map[string]LanFirewallRule{
		"firewall_rule_1.1.2.1.1": {
			ID:          types.StringValue("f1.1.2.1.1"),
			NetRuleName: types.StringValue("rule_1.1.2.1"),
			IndexInRule: types.Int64Value(1),
		},
		"firewall_rule_1.1.2.1.3": {
			ID:          types.StringValue("f1.1.2.1.3"),
			NetRuleName: types.StringValue("rule_1.1.2.1"),
			IndexInRule: types.Int64Value(2),
		},
		"firewall_rule_1.1.2.1.2": {
			ID:          types.StringValue("f1.1.2.1.2"),
			NetRuleName: types.StringValue("rule_1.1.2.1"),
			IndexInRule: types.Int64Value(3),
		},
	}, firewallRules)

	require.Equal(t, &lfIndexMap{
		sections: map[string]itemOrder{
			"": {
				parentID: "",
				current:  []nameID{{name: "section_1", id: "s1"}, {name: "section_3", id: "s3"}, {name: "section_2", id: "s2"}},
				target:   []nameID{{name: "section_1", id: "s1"}, {name: "section_2", id: "s2"}, {name: "section_3", id: "s3"}},
			},
			"sub_policy_1.1": {
				parentID: "sub1.1",
				current:  []nameID{{name: "section_1.1.3", id: "s1.1.3"}, {name: "section_1.1.1", id: "s1.1.1"}, {name: "section_1.1.2", id: "s1.1.2"}},
				target:   []nameID{{name: "section_1.1.1", id: "s1.1.1"}, {name: "section_1.1.2", id: "s1.1.2"}, {name: "section_1.1.3", id: "s1.1.3"}},
			},
		},
		rulesOrSubPols: map[string]itemOrderType{
			"section_1": {
				parentID: "s1",
				current: []nameIDType{
					{name: "rule_1.2", id: "r1.2", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "rule_1.1", id: "r1.1", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "sub_policy_1.1", id: "sub1.1", ruleType: cato_models.PolicyRuleTypeEnumSubPolicyScope},
					{name: "sub_policy_1.3", id: "sub1.3", ruleType: cato_models.PolicyRuleTypeEnumSubPolicyScope},
					{name: "sub_policy_1.2", id: "sub1.2", ruleType: cato_models.PolicyRuleTypeEnumSubPolicyScope},
					{name: "rule_1.3", id: "r1.3", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
				},
				target: []nameIDType{
					{name: "rule_1.1", id: "r1.1", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "rule_1.2", id: "r1.2", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "sub_policy_1.1", id: "sub1.1", ruleType: cato_models.PolicyRuleTypeEnumSubPolicyScope},
					{name: "sub_policy_1.2", id: "sub1.2", ruleType: cato_models.PolicyRuleTypeEnumSubPolicyScope},
					{name: "sub_policy_1.3", id: "sub1.3", ruleType: cato_models.PolicyRuleTypeEnumSubPolicyScope},
					{name: "rule_1.3", id: "r1.3", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
				},
			},
			"section_1.1.2": {
				parentID: "s1.1.2",
				current: []nameIDType{
					{name: "rule_1.1.2.1", id: "r1.1.2.1", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "rule_1.1.2.3", id: "r1.1.2.3", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "rule_1.1.2.2", id: "r1.1.2.2", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
				},
				target: []nameIDType{
					{name: "rule_1.1.2.1", id: "r1.1.2.1", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "rule_1.1.2.2", id: "r1.1.2.2", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
					{name: "rule_1.1.2.3", id: "r1.1.2.3", ruleType: cato_models.PolicyRuleTypeEnumPolicyRule},
				},
			},
		},
		firewallRules: map[string]itemOrder{
			"rule_1.2": {parentID: "r1.2"},
			"rule_1.1": {parentID: "r1.1"},
			"rule_1.3": {parentID: "r1.3"},
			"rule_1.1.2.1": {
				parentID: "r1.1.2.1",
				current: []nameID{
					{name: "firewall_rule_1.1.2.1.1", id: "f1.1.2.1.1"},
					{name: "firewall_rule_1.1.2.1.3", id: "f1.1.2.1.3"},
					{name: "firewall_rule_1.1.2.1.2", id: "f1.1.2.1.2"},
				},
				target: []nameID{
					{name: "firewall_rule_1.1.2.1.1", id: "f1.1.2.1.1"},
					{name: "firewall_rule_1.1.2.1.2", id: "f1.1.2.1.2"},
					{name: "firewall_rule_1.1.2.1.3", id: "f1.1.2.1.3"},
				},
			},
			"rule_1.1.2.3": {parentID: "r1.1.2.3"},
			"rule_1.1.2.2": {parentID: "r1.1.2.2"},
		},
	}, indexMap)
}

func TestHydrateLanFwLeavesOmittedRuleMapsUnmanaged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics
	plan := lPMockClient.createPlan(lanPolicyPlans["default"])
	plan.NetworkRules = types.MapNull(types.ObjectType{AttrTypes: LanNetworkRuleTypes})
	plan.FirewallRules = types.MapNull(types.ObjectType{AttrTypes: LanFirewallRuleTypes})
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{policy: mockLanPolicy["default"]},
	}

	newState, indexMap := res.hydrateLanFwRulesIndex(ctx, plan, &diags)

	require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
	require.True(t, newState.NetworkRules.IsNull())
	require.True(t, newState.FirewallRules.IsNull())
	require.NotEmpty(t, indexMap.rulesOrSubPols)
	require.NotEmpty(t, indexMap.firewallRules)
	for _, order := range indexMap.rulesOrSubPols {
		require.Empty(t, order.target)
	}
	for _, order := range indexMap.firewallRules {
		require.Empty(t, order.target)
	}
}

func TestReadLanFwPreservesUnmanagedRuleMaps(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stateModel := lPMockClient.createPlan(lanPolicyPlans["default"])
	stateModel.NetworkRules = types.MapNull(types.ObjectType{AttrTypes: LanNetworkRuleTypes})
	stateModel.FirewallRules = types.MapNull(types.ObjectType{AttrTypes: LanFirewallRuleTypes})
	resourceSchema := getLanFwRulesIndexSchema(ctx, t)
	requestState := tfsdk.State{Schema: resourceSchema}
	require.False(t, requestState.Set(ctx, stateModel).HasError())
	response := &resource.ReadResponse{State: tfsdk.State{Schema: resourceSchema}}
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{policy: mockLanPolicy["default"]},
	}

	res.Read(ctx, resource.ReadRequest{State: requestState}, response)

	require.False(t, response.Diagnostics.HasError(), "unexpected diagnostics: %+v", response.Diagnostics)
	var state LanFwRulesIndex
	require.False(t, response.State.Get(ctx, &state).HasError())
	require.True(t, state.NetworkRules.IsNull())
	require.True(t, state.FirewallRules.IsNull())
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	mockClient := &lanPolicyMockClient{policy: mockLanPolicy["default"]}
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: mockClient,
	}

	request := resource.CreateRequest{Plan: newLanFwRulesIndexPlan(ctx, t)}
	response := &resource.CreateResponse{
		State: tfsdk.State{Schema: getLanFwRulesIndexSchema(ctx, t)},
	}
	res.Create(ctx, request, response)

	require.False(t, response.Diagnostics.HasError(), "unexpected diagnostics: %+v", response.Diagnostics)
	var state LanFwRulesIndex
	stateDiags := response.State.Get(ctx, &state)
	require.False(t, stateDiags.HasError(), "unexpected state diagnostics: %+v", stateDiags)
	var firewallRules map[string]LanFirewallRule
	firewallDiags := state.FirewallRules.ElementsAs(ctx, &firewallRules, false)
	require.False(t, firewallDiags.HasError(), "unexpected firewall state diagnostics: %+v", firewallDiags)
	require.Len(t, firewallRules, 3)
	require.Equal(t, 2, mockClient.policyCalls, "expected initial and final policy reads")
	require.ElementsMatch(t, []lanPolicyMoveSectionCall{
		{
			input: cato_models.PolicyMoveSectionInput{
				ID: "s3",
				To: &cato_models.PolicySectionPositionInput{
					Position: cato_models.PolicySectionPositionEnumLastInPolicy,
				},
			},
			accountID: "testID",
		},
		{
			input: cato_models.PolicyMoveSectionInput{
				ID: "s1.1.3",
				To: &cato_models.PolicySectionPositionInput{
					Position: cato_models.PolicySectionPositionEnumLastInPolicy,
					Ref:      ptr("sub1.1"),
				},
			},
			accountID: "testID",
		},
	}, mockClient.moveSectionCalls)
	require.ElementsMatch(t, []lanPolicyMoveRuleCall{
		{
			input: cato_models.PolicyMoveRuleInput{
				ID: "sub1.3",
				To: &cato_models.PolicyRulePositionInput{
					Position: ptr(cato_models.PolicyRulePositionEnumBeforeRule),
					Ref:      ptr("r1.3"),
				},
			},
			accountID: "testID",
		},
		{
			input: cato_models.PolicyMoveRuleInput{
				ID: "r1.2",
				To: &cato_models.PolicyRulePositionInput{
					Position: ptr(cato_models.PolicyRulePositionEnumBeforeRule),
					Ref:      ptr("sub1.1"),
				},
			},
			accountID: "testID",
		},
		{
			input: cato_models.PolicyMoveRuleInput{
				ID: "r1.1.2.3",
				To: &cato_models.PolicyRulePositionInput{
					Position: ptr(cato_models.PolicyRulePositionEnumLastInSection),
					Ref:      ptr("s1.1.2"),
				},
			},
			accountID: "testID",
		},
	}, mockClient.moveRuleCalls)
	require.ElementsMatch(t, []lanPolicyFirewallMoveRuleCall{
		{
			accountID: "testID",
			input: cato_models.PolicyMoveSubRuleInput{
				ID: "f1.1.2.1.3",
				To: &cato_models.PolicySubRulePositionInput{
					Position: cato_models.PolicySubRulePositionEnumLastInRule,
					Ref:      "r1.1.2.1",
				},
			},
		},
	}, mockClient.firewallMoveRuleCalls)
	require.Equal(t, 1, mockClient.publishCalls)
}

func getLanFwRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	(&lanRulesIndexResource{}).Schema(ctx, resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "unexpected schema diagnostics: %+v", resp.Diagnostics)
	return resp.Schema
}

func newLanFwRulesIndexPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: getLanFwRulesIndexSchema(ctx, t)}
	if diags := plan.Set(ctx, lPMockClient.createPlan(lanPolicyPlans["default"])); diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}

type lanPolicyMockClient struct {
	policy                *cato_go_sdk.PolicySocketLanPolicy
	policyCalls           int
	moveSectionCalls      []lanPolicyMoveSectionCall
	moveRuleCalls         []lanPolicyMoveRuleCall
	firewallMoveRuleCalls []lanPolicyFirewallMoveRuleCall
	publishCalls          int
}

type lanPolicyMoveSectionCall struct {
	input     cato_models.PolicyMoveSectionInput
	accountID string
}

type lanPolicyMoveRuleCall struct {
	input     cato_models.PolicyMoveRuleInput
	accountID string
}

type lanPolicyFirewallMoveRuleCall struct {
	accountID string
	input     cato_models.PolicyMoveSubRuleInput
}

var lPMockClient lanPolicyMockClient

func (m *lanPolicyMockClient) PolicySocketLanPolicy(_ context.Context, _ string,
	_ *cato_models.SocketLanPolicyInput, _ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPolicy, error) {
	m.policyCalls++
	return m.policy, nil
}

func (m *lanPolicyMockClient) PolicySocketLanMoveSection(_ context.Context, input cato_models.PolicyMoveSectionInput,
	accountID string, _ ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveSection, error) {
	m.moveSectionCalls = append(m.moveSectionCalls, lanPolicyMoveSectionCall{input: input, accountID: accountID})
	return &cato_go_sdk.PolicySocketLanMoveSection{
		Policy: &cato_go_sdk.PolicySocketLanMoveSection_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanMoveSection_Policy_SocketLan{},
		},
	}, nil
}

func (m *lanPolicyMockClient) PolicySocketLanMoveRule(_ context.Context, input cato_models.PolicyMoveRuleInput,
	accountID string, _ ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveRule, error) {
	m.moveRuleCalls = append(m.moveRuleCalls, lanPolicyMoveRuleCall{input: input, accountID: accountID})
	return &cato_go_sdk.PolicySocketLanMoveRule{
		Policy: &cato_go_sdk.PolicySocketLanMoveRule_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanMoveRule_Policy_SocketLan{},
		},
	}, nil
}
func (m *lanPolicyMockClient) PolicySocketLanFirewallMoveRule(_ context.Context, accountID string,
	_ *cato_models.SocketLanPolicyMutationInput, input cato_models.PolicyMoveSubRuleInput, _ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanFirewallMoveRule, error) {
	m.firewallMoveRuleCalls = append(m.firewallMoveRuleCalls, lanPolicyFirewallMoveRuleCall{
		accountID: accountID,
		input:     input,
	})
	return &cato_go_sdk.PolicySocketLanFirewallMoveRule{
		Policy: &cato_go_sdk.PolicySocketLanFirewallMoveRule_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanFirewallMoveRule_Policy_SocketLan{},
		},
	}, nil
}

func (m *lanPolicyMockClient) PolicySocketLanPublishPolicyRevision(_ context.Context,
	_ *cato_models.SocketLanPolicyMutationInput, _ *cato_models.PolicyPublishRevisionInput, _ string,
	_ ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error) {
	m.publishCalls++
	return &cato_go_sdk.PolicySocketLanPublishPolicyRevision{
		Policy: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan{},
		},
	}, nil
}

type lanPolicyPlanItem struct {
	sectionData      map[string]lanPolicySectionData
	networkRuleData  map[string]lanPolocyNetRuleData
	firewallRuleData map[string]lanPolocyFwRuleData
}
type lanPolicySectionData struct {
	id            string
	subPolicyName string
	sectionIndex  int
}
type lanPolocyNetRuleData struct {
	id             string
	ruleType       string
	sectionName    string
	indexInSection int
}

type lanPolocyFwRuleData struct {
	id          string
	netRuleName string
	indexInRule int
}

//	   section_1
//	   		rule_1.1
//	   		rule_1.2
//	   		sub_policy_1.1
//	   			section_1.1.1
//	   			section_1.1.2
//						rule_1.1.2.1
//	                        firewall_rule_1.1.2.1.1
//	                        firewall_rule_1.1.2.1.2
//	                        firewall_rule_1.1.2.1.3
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
		networkRuleData: map[string]lanPolocyNetRuleData{
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
		firewallRuleData: map[string]lanPolocyFwRuleData{
			"firewall_rule_1.1.2.1.1": {id: "fr1.1.2.1.1", netRuleName: "rule_1.1.2.1", indexInRule: 1},
			"firewall_rule_1.1.2.1.2": {id: "fr1.1.2.1.2", netRuleName: "rule_1.1.2.1", indexInRule: 2},
			"firewall_rule_1.1.2.1.3": {id: "fr1.1.2.1.3", netRuleName: "rule_1.1.2.1", indexInRule: 3},
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

	ruleObjects := make(map[string]types.Object, len(p.networkRuleData))
	for ruleName, rule := range p.networkRuleData {
		ruleObject, diags := types.ObjectValueFrom(ctx, LanNetworkRuleTypes, LanNetworkRule{
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
		types.ObjectType{AttrTypes: LanNetworkRuleTypes},
		ruleObjects,
	)
	if ruleDiags.HasError() {
		return &LanFwRulesIndex{}
	}

	firewallRuleObjects := make(map[string]types.Object, len(p.firewallRuleData))
	for ruleName, rule := range p.firewallRuleData {
		ruleObject, diags := types.ObjectValueFrom(ctx, LanFirewallRuleTypes, LanFirewallRule{
			ID:          types.StringValue(rule.id),
			NetRuleName: types.StringValue(rule.netRuleName),
			IndexInRule: types.Int64Value(int64(rule.indexInRule)),
		})
		if diags.HasError() {
			return &LanFwRulesIndex{}
		}
		firewallRuleObjects[ruleName] = ruleObject
	}
	firewallRuleData, firewallRuleDiags := types.MapValueFrom(
		ctx,
		types.ObjectType{AttrTypes: LanFirewallRuleTypes},
		firewallRuleObjects,
	)
	if firewallRuleDiags.HasError() {
		return &LanFwRulesIndex{}
	}

	return &LanFwRulesIndex{
		SectionData:   sectionData,
		NetworkRules:  ruleData,
		FirewallRules: firewallRuleData,
	}
}

// LAN Firewall Policy returned by the API
// Note: the slices are in "random" order, so that we can test the reordering
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
								ID:   "sub1.3",
								Name: "sub_policy_1.3",
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
							RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
							Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
								ID:   "r1.1.2.1",
								Name: "rule_1.1.2.1",
								Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
									ID:   "s1.1.2",
									Name: "section_1.1.2",
								},
								Firewall: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall{
									{
										Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule{
											ID:   "f1.1.2.1.1",
											Name: "firewall_rule_1.1.2.1.1",
										},
									},
									{
										Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule{
											ID:   "f1.1.2.1.3",
											Name: "firewall_rule_1.1.2.1.3",
										},
									},
									{
										Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule{
											ID:   "f1.1.2.1.2",
											Name: "firewall_rule_1.1.2.1.2",
										},
									},
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
