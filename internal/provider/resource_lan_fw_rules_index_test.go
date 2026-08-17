package provider

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"strings"
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

func TestBulkLfMoveRuleExampleData(t *testing.T) {
	t.Parallel()

	payload, err := os.ReadFile("../../examples/resources/cato_bulk_lf_move_rule/data/lanFirewall.json")
	require.NoError(t, err)
	var fixture struct {
		Sections      map[string]any `json:"sections"`
		NetworkRules  map[string]any `json:"network_rules"`
		FirewallRules map[string]any `json:"firewall_rules"`
		Policies      map[string]any `json:"policies"`
	}
	require.NoError(t, json.Unmarshal(payload, &fixture))
	require.NotEmpty(t, fixture.Sections)
	require.NotEmpty(t, fixture.NetworkRules)
	require.NotEmpty(t, fixture.FirewallRules)
	require.NotEmpty(t, fixture.Policies)
}

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
		"section_1":     {ID: types.StringValue("s1"), SectionName: types.StringValue("section_1"), SectionIndex: types.Int64Value(1), SubPolicyName: types.StringValue("")},
		"section_3":     {ID: types.StringValue("s3"), SectionName: types.StringValue("section_3"), SectionIndex: types.Int64Value(2), SubPolicyName: types.StringValue("")},
		"section_2":     {ID: types.StringValue("s2"), SectionName: types.StringValue("section_2"), SectionIndex: types.Int64Value(3), SubPolicyName: types.StringValue("")},
		"section_1.1.3": {ID: types.StringValue("s1.1.3"), SectionName: types.StringValue("section_1.1.3"), SectionIndex: types.Int64Value(1), SubPolicyName: types.StringValue("sub_policy_1.1")},
		"section_1.1.1": {ID: types.StringValue("s1.1.1"), SectionName: types.StringValue("section_1.1.1"), SectionIndex: types.Int64Value(2), SubPolicyName: types.StringValue("sub_policy_1.1")},
		"section_1.1.2": {ID: types.StringValue("s1.1.2"), SectionName: types.StringValue("section_1.1.2"), SectionIndex: types.Int64Value(3), SubPolicyName: types.StringValue("sub_policy_1.1")},
	}, sectionData)

	var networkRules map[string]LanNetworkRule
	networkDiags := newState.NetworkRules.ElementsAs(ctx, &networkRules, false)
	require.False(t, networkDiags.HasError(), "unexpected network rule state diagnostics: %+v", networkDiags)
	require.Equal(t, map[string]LanNetworkRule{
		"rule_1.2":       {ID: types.StringValue("r1.2"), RuleType: types.StringValue("POLICY_RULE"), RuleName: types.StringValue("rule_1.2"), SectionName: types.StringValue("section_1"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(1)},
		"rule_1.1":       {ID: types.StringValue("r1.1"), RuleType: types.StringValue("POLICY_RULE"), RuleName: types.StringValue("rule_1.1"), SectionName: types.StringValue("section_1"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(2)},
		"sub_policy_1.1": {ID: types.StringValue("sub1.1"), RuleType: types.StringValue("SUB_POLICY_SCOPE"), RuleName: types.StringValue("sub_policy_1.1"), SectionName: types.StringValue("section_1"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(3)},
		"sub_policy_1.2": {ID: types.StringValue("sub1.2"), RuleType: types.StringValue("SUB_POLICY_SCOPE"), RuleName: types.StringValue("sub_policy_1.2"), SectionName: types.StringValue("section_1"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(5)},
		"sub_policy_1.3": {ID: types.StringValue("sub1.3"), RuleType: types.StringValue("SUB_POLICY_SCOPE"), RuleName: types.StringValue("sub_policy_1.3"), SectionName: types.StringValue("section_1"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(4)},
		"rule_1.3":       {ID: types.StringValue("r1.3"), RuleType: types.StringValue("POLICY_RULE"), RuleName: types.StringValue("rule_1.3"), SectionName: types.StringValue("section_1"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(6)},
		"rule_1.1.2.1":   {ID: types.StringValue("r1.1.2.1"), RuleType: types.StringValue("POLICY_RULE"), RuleName: types.StringValue("rule_1.1.2.1"), SectionName: types.StringValue("section_1.1.2"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(1)},
		"rule_1.1.2.3":   {ID: types.StringValue("r1.1.2.3"), RuleType: types.StringValue("POLICY_RULE"), RuleName: types.StringValue("rule_1.1.2.3"), SectionName: types.StringValue("section_1.1.2"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(2)},
		"rule_1.1.2.2":   {ID: types.StringValue("r1.1.2.2"), RuleType: types.StringValue("POLICY_RULE"), RuleName: types.StringValue("rule_1.1.2.2"), SectionName: types.StringValue("section_1.1.2"), SectionKey: types.StringNull(), IndexInSection: types.Int64Value(3)},
	}, networkRules)

	var firewallRules map[string]LanFirewallRule
	firewallDiags := newState.FirewallRules.ElementsAs(ctx, &firewallRules, false)
	require.False(t, firewallDiags.HasError(), "unexpected firewall rule state diagnostics: %+v", firewallDiags)
	require.Equal(t, map[string]LanFirewallRule{
		"firewall_rule_1.1.2.1.1": {
			ID:               types.StringValue("f1.1.2.1.1"),
			FirewallRuleName: types.StringValue("firewall_rule_1.1.2.1.1"),
			NetRuleName:      types.StringValue("rule_1.1.2.1"),
			NetRuleKey:       types.StringNull(),
			IndexInRule:      types.Int64Value(1),
		},
		"firewall_rule_1.1.2.1.3": {
			ID:               types.StringValue("f1.1.2.1.3"),
			FirewallRuleName: types.StringValue("firewall_rule_1.1.2.1.3"),
			NetRuleName:      types.StringValue("rule_1.1.2.1"),
			NetRuleKey:       types.StringNull(),
			IndexInRule:      types.Int64Value(2),
		},
		"firewall_rule_1.1.2.1.2": {
			ID:               types.StringValue("f1.1.2.1.2"),
			FirewallRuleName: types.StringValue("firewall_rule_1.1.2.1.2"),
			NetRuleName:      types.StringValue("rule_1.1.2.1"),
			NetRuleKey:       types.StringNull(),
			IndexInRule:      types.Int64Value(3),
		},
	}, firewallRules)

	require.Equal(t, &lfIndexMap{
		sections: map[string]itemOrder{
			"": {
				parentID:   "",
				parentName: "",
				current:    []nameID{{name: "section_1", id: "s1"}, {name: "section_3", id: "s3"}, {name: "section_2", id: "s2"}},
				target:     []nameID{{name: "section_1", id: "s1"}, {name: "section_2", id: "s2"}, {name: "section_3", id: "s3"}},
			},
			"sub1.1": {
				parentID:   "sub1.1",
				parentName: "sub_policy_1.1",
				current:    []nameID{{name: "section_1.1.3", id: "s1.1.3"}, {name: "section_1.1.1", id: "s1.1.1"}, {name: "section_1.1.2", id: "s1.1.2"}},
				target:     []nameID{{name: "section_1.1.1", id: "s1.1.1"}, {name: "section_1.1.2", id: "s1.1.2"}, {name: "section_1.1.3", id: "s1.1.3"}},
			},
		},
		rulesOrSubPols: map[string]itemOrderType{
			"s1": {
				parentID:   "s1",
				parentName: "section_1",
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
			"s1.1.2": {
				parentID:   "s1.1.2",
				parentName: "section_1.1.2",
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
			"r1.2": {parentID: "r1.2", parentName: "rule_1.2"},
			"r1.1": {parentID: "r1.1", parentName: "rule_1.1"},
			"r1.3": {parentID: "r1.3", parentName: "rule_1.3"},
			"r1.1.2.1": {
				parentID:   "r1.1.2.1",
				parentName: "rule_1.1.2.1",
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
			"r1.1.2.3": {parentID: "r1.1.2.3", parentName: "rule_1.1.2.3"},
			"r1.1.2.2": {parentID: "r1.1.2.2", parentName: "rule_1.1.2.2"},
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

func TestHydrateLanFwDuplicateNamesWithAliases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics
	plan := lPMockClient.createPlan(duplicateLanPolicyPlan(true, true))
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{policy: duplicateLanPolicy()},
	}

	state, indexMap := res.hydrateLanFwRulesIndex(ctx, plan, &diags)

	require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
	require.Equal(t, []string{"child-section", "main-section"}, mapKeys[LanFwSectionData](ctx, t, state.SectionData))
	require.Equal(t, []string{"child-net", "main-net"}, mapKeys[LanNetworkRule](ctx, t, state.NetworkRules))
	require.Equal(t, []string{"child-fw", "main-fw"}, mapKeys[LanFirewallRule](ctx, t, state.FirewallRules))
	require.Contains(t, indexMap.sections, "")
	require.Contains(t, indexMap.sections, "sp-child")
	require.Contains(t, indexMap.rulesOrSubPols, "s-main")
	require.Contains(t, indexMap.rulesOrSubPols, "s-child")
	require.Contains(t, indexMap.firewallRules, "r-main")
	require.Contains(t, indexMap.firewallRules, "r-child")

	var sections map[string]LanFwSectionData
	require.False(t, state.SectionData.ElementsAs(ctx, &sections, false).HasError())
	require.Equal(t, "Shared section", sections["main-section"].SectionName.ValueString())
	require.Equal(t, "Shared section", sections["child-section"].SectionName.ValueString())

	var rules map[string]LanNetworkRule
	require.False(t, state.NetworkRules.ElementsAs(ctx, &rules, false).HasError())
	require.Equal(t, "Shared network rule", rules["main-net"].RuleName.ValueString())
	require.Equal(t, "main-section", rules["main-net"].SectionKey.ValueString())
	require.Equal(t, "child-section", rules["child-net"].SectionKey.ValueString())

	var firewallRules map[string]LanFirewallRule
	require.False(t, state.FirewallRules.ElementsAs(ctx, &firewallRules, false).HasError())
	require.Equal(t, "Shared firewall rule", firewallRules["main-fw"].FirewallRuleName.ValueString())
	require.Equal(t, "main-net", firewallRules["main-fw"].NetRuleKey.ValueString())
	require.Equal(t, "child-net", firewallRules["child-fw"].NetRuleKey.ValueString())
}

func TestReadLanFwPreservesAliasKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var hydrateDiags diag.Diagnostics
	plan := lPMockClient.createPlan(duplicateLanPolicyPlan(true, true))
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{policy: duplicateLanPolicy()},
	}
	initialState, _ := res.hydrateLanFwRulesIndex(ctx, plan, &hydrateDiags)
	require.False(t, hydrateDiags.HasError(), "unexpected hydration diagnostics: %+v", hydrateDiags)

	resourceSchema := getLanFwRulesIndexSchema(ctx, t)
	requestState := tfsdk.State{Schema: resourceSchema}
	require.False(t, requestState.Set(ctx, initialState).HasError())
	response := &resource.ReadResponse{State: tfsdk.State{Schema: resourceSchema}}

	res.Read(ctx, resource.ReadRequest{State: requestState}, response)

	require.False(t, response.Diagnostics.HasError(), "unexpected diagnostics: %+v", response.Diagnostics)
	var state LanFwRulesIndex
	require.False(t, response.State.Get(ctx, &state).HasError())
	require.Equal(t, []string{"child-section", "main-section"}, mapKeys[LanFwSectionData](ctx, t, state.SectionData))
	require.Equal(t, []string{"child-net", "main-net"}, mapKeys[LanNetworkRule](ctx, t, state.NetworkRules))
	require.Equal(t, []string{"child-fw", "main-fw"}, mapKeys[LanFirewallRule](ctx, t, state.FirewallRules))
}

func TestReadLanFwUpdatesParentKeyWhenNetworkRuleMoves(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var hydrateDiags diag.Diagnostics
	mockClient := &lanPolicyMockClient{policy: duplicateLanPolicy()}
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: mockClient,
	}
	initialState, _ := res.hydrateLanFwRulesIndex(
		ctx, lPMockClient.createPlan(duplicateLanPolicyPlan(true, true)), &hydrateDiags,
	)
	require.False(t, hydrateDiags.HasError(), "unexpected hydration diagnostics: %+v", hydrateDiags)

	mockClient.policy = networkRulesMovedLanPolicy()
	state := readLanFwState(ctx, t, &res, initialState)

	require.Equal(t, []string{"child-net", "main-net"}, mapKeys[LanNetworkRule](ctx, t, state.NetworkRules))
	var rules map[string]LanNetworkRule
	require.False(t, state.NetworkRules.ElementsAs(ctx, &rules, false).HasError())
	require.Equal(t, "child-section", rules["main-net"].SectionKey.ValueString())
	require.Equal(t, "main-section", rules["child-net"].SectionKey.ValueString())
}

func TestReadLanFwUpdatesParentKeyWhenFirewallRuleMoves(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var hydrateDiags diag.Diagnostics
	mockClient := &lanPolicyMockClient{policy: duplicateLanPolicy()}
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: mockClient,
	}
	initialState, _ := res.hydrateLanFwRulesIndex(
		ctx, lPMockClient.createPlan(duplicateLanPolicyPlan(true, true)), &hydrateDiags,
	)
	require.False(t, hydrateDiags.HasError(), "unexpected hydration diagnostics: %+v", hydrateDiags)

	mockClient.policy = firewallRulesMovedLanPolicy()
	state := readLanFwState(ctx, t, &res, initialState)

	require.Equal(t, []string{"child-fw", "main-fw"}, mapKeys[LanFirewallRule](ctx, t, state.FirewallRules))
	var rules map[string]LanFirewallRule
	require.False(t, state.FirewallRules.ElementsAs(ctx, &rules, false).HasError())
	require.Equal(t, "child-net", rules["main-fw"].NetRuleKey.ValueString())
	require.Equal(t, "main-net", rules["child-fw"].NetRuleKey.ValueString())
}

func TestReadLanFwIgnoresStaleAliasesAndHydratesAPIDrift(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var hydrateDiags diag.Diagnostics
	mockClient := &lanPolicyMockClient{policy: duplicateLanPolicy()}
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: mockClient,
	}
	initialState, _ := res.hydrateLanFwRulesIndex(
		ctx, lPMockClient.createPlan(duplicateLanPolicyPlan(true, true)), &hydrateDiags,
	)
	require.False(t, hydrateDiags.HasError(), "unexpected hydration diagnostics: %+v", hydrateDiags)

	mockClient.policy = addedAndRemovedLanPolicy()
	state := readLanFwState(ctx, t, &res, initialState)

	require.Equal(t, []string{"New section", "main-section"}, mapKeys[LanFwSectionData](ctx, t, state.SectionData))
	require.Equal(t, []string{"New network rule", "main-net"}, mapKeys[LanNetworkRule](ctx, t, state.NetworkRules))
	require.Equal(t, []string{"New firewall rule", "main-fw"}, mapKeys[LanFirewallRule](ctx, t, state.FirewallRules))

	var rules map[string]LanNetworkRule
	require.False(t, state.NetworkRules.ElementsAs(ctx, &rules, false).HasError())
	require.Equal(t, "New section", rules["New network rule"].SectionKey.ValueString())
	var firewallRules map[string]LanFirewallRule
	require.False(t, state.FirewallRules.ElementsAs(ctx, &firewallRules, false).HasError())
	require.Equal(t, "New network rule", firewallRules["New firewall rule"].NetRuleKey.ValueString())
}

func TestReadLanFwAssignsCollisionSafeKeysToDiscoveredDuplicates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var hydrateDiags diag.Diagnostics
	mockClient := &lanPolicyMockClient{policy: singleLanPolicy()}
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: mockClient,
	}
	initialState, _ := res.hydrateLanFwRulesIndex(
		ctx, lPMockClient.createPlan(singleLanPolicyPlan()), &hydrateDiags,
	)
	require.False(t, hydrateDiags.HasError(), "unexpected hydration diagnostics: %+v", hydrateDiags)

	mockClient.policy = duplicateLanPolicy()
	state := readLanFwState(ctx, t, &res, initialState)

	require.Equal(t,
		[]string{"Shared section", "Shared section__s-child"},
		mapKeys[LanFwSectionData](ctx, t, state.SectionData),
	)
	require.Equal(t,
		[]string{"Shared network rule", "Shared network rule__r-child"},
		mapKeys[LanNetworkRule](ctx, t, state.NetworkRules),
	)
	require.Equal(t,
		[]string{"Shared firewall rule", "Shared firewall rule__f-child"},
		mapKeys[LanFirewallRule](ctx, t, state.FirewallRules),
	)

	var rules map[string]LanNetworkRule
	require.False(t, state.NetworkRules.ElementsAs(ctx, &rules, false).HasError())
	require.Equal(t, "Shared section__s-child",
		rules["Shared network rule__r-child"].SectionKey.ValueString())
	var firewallRules map[string]LanFirewallRule
	require.False(t, state.FirewallRules.ElementsAs(ctx, &firewallRules, false).HasError())
	require.Equal(t, "Shared network rule__r-child",
		firewallRules["Shared firewall rule__f-child"].NetRuleKey.ValueString())
}

func readLanFwState(ctx context.Context, t *testing.T, res *lanRulesIndexResource,
	stateModel *LanFwRulesIndex,
) LanFwRulesIndex {
	t.Helper()
	resourceSchema := getLanFwRulesIndexSchema(ctx, t)
	requestState := tfsdk.State{Schema: resourceSchema}
	require.False(t, requestState.Set(ctx, stateModel).HasError())
	response := &resource.ReadResponse{State: tfsdk.State{Schema: resourceSchema}}

	res.Read(ctx, resource.ReadRequest{State: requestState}, response)

	require.False(t, response.Diagnostics.HasError(), "unexpected diagnostics: %+v", response.Diagnostics)
	var state LanFwRulesIndex
	require.False(t, response.State.Get(ctx, &state).HasError())
	return state
}

func TestHydrateLanFwAmbiguousSectionNameRequiresKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics
	plan := lPMockClient.createPlan(duplicateLanPolicyPlan(false, true))
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{policy: duplicateLanPolicy()},
	}

	res.hydrateLanFwRulesIndex(ctx, plan, &diags)

	requireDiagnosticContains(t, diags,
		`section name "Shared section" is ambiguous; set section_key to the parent map key`)
}

func TestHydrateLanFwAmbiguousNetworkRuleNameRequiresKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics
	plan := lPMockClient.createPlan(duplicateLanPolicyPlan(true, false))
	res := lanRulesIndexResource{
		client:       &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{policy: duplicateLanPolicy()},
	}

	res.hydrateLanFwRulesIndex(ctx, plan, &diags)

	requireDiagnosticContains(t, diags,
		`network rule name "Shared network rule" is ambiguous; set net_rule_key to the parent map key`)
}

func mapKeys[T any](ctx context.Context, t *testing.T, value types.Map) []string {
	t.Helper()
	var elements map[string]T
	require.False(t, value.ElementsAs(ctx, &elements, false).HasError())
	keys := make([]string, 0, len(elements))
	for key := range elements {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func requireDiagnosticContains(t *testing.T, diags diag.Diagnostics, want string) {
	t.Helper()
	require.True(t, diags.HasError(), "expected error diagnostic containing %q", want)
	var details []string
	for _, diagnostic := range diags {
		details = append(details, diagnostic.Detail())
	}
	require.True(t, strings.Contains(strings.Join(details, "\n"), want),
		"diagnostic details did not contain %q: %v", want, details)
}

func TestLanPolicyMutationStatusError(t *testing.T) {
	t.Parallel()

	success := cato_models.PolicyMutationStatusSuccess
	failure := cato_models.PolicyMutationStatusFailure
	missing := cato_models.PolicyMutationStatus("")
	tests := []struct {
		name    string
		status  *cato_models.PolicyMutationStatus
		wantErr string
	}{
		{name: "success", status: &success},
		{name: "failure", status: &failure, wantErr: `API mutation returned status "FAILURE"`},
		{name: "missing pointer", wantErr: "API mutation returned no status"},
		{name: "missing value", status: &missing, wantErr: `API mutation returned status ""`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := lanPolicyMutationStatusError(tc.status)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestLanFwMoveMutationsRejectFailureStatusWithoutErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(context.Context, *lanRulesIndexResource) error
	}{
		{
			name: "section",
			call: func(ctx context.Context, r *lanRulesIndexResource) error {
				return r.moveSectionToPosition(ctx, "",
					[]nameID{{name: "first", id: "first"}, {name: "second", id: "second"}},
					"first", "first", 1)
			},
		},
		{
			name: "network rule",
			call: func(ctx context.Context, r *lanRulesIndexResource) error {
				return r.moveNetRuleToPosition(ctx, "section",
					[]nameIDType{{name: "first", id: "first"}, {name: "second", id: "second"}},
					"first", "first", 1)
			},
		},
		{
			name: "sub-policy",
			call: func(ctx context.Context, r *lanRulesIndexResource) error {
				return r.moveSubPolicyToPosition(ctx, "section",
					[]nameIDType{{name: "first", id: "first"}, {name: "second", id: "second"}},
					"first", "first", 1)
			},
		},
		{
			name: "firewall rule",
			call: func(ctx context.Context, r *lanRulesIndexResource) error {
				return r.moveFwRuleToPosition(ctx, "network-rule",
					[]nameID{{name: "first", id: "first"}, {name: "second", id: "second"}},
					"first", "first", 1)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := lanRulesIndexResource{
				client: &catoClientData{AccountId: "testID"},
				catov2Client: &lanPolicyMockClient{
					mutationStatus: cato_models.PolicyMutationStatusFailure,
				},
			}

			err := tc.call(context.Background(), &r)

			require.ErrorContains(t, err, `API mutation returned status "FAILURE"`)
		})
	}
}

func TestLanFwPublishRejectsFailureStatusWithoutErrors(t *testing.T) {
	t.Parallel()

	r := lanRulesIndexResource{
		client: &catoClientData{AccountId: "testID"},
		catov2Client: &lanPolicyMockClient{
			mutationStatus: cato_models.PolicyMutationStatusFailure,
		},
	}
	var diags diag.Diagnostics

	r.publish(context.Background(), &diags)

	requireDiagnosticContains(t, diags, `API mutation returned status "FAILURE"`)
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
	mutationStatus        cato_models.PolicyMutationStatus
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

func (m *lanPolicyMockClient) responseMutationStatus() cato_models.PolicyMutationStatus {
	if m.mutationStatus == "" {
		return cato_models.PolicyMutationStatusSuccess
	}
	return m.mutationStatus
}

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
			SocketLan: &cato_go_sdk.PolicySocketLanMoveSection_Policy_SocketLan{
				MoveSection: cato_go_sdk.PolicySocketLanMoveSection_Policy_SocketLan_MoveSection{
					Status: m.responseMutationStatus(),
				},
			},
		},
	}, nil
}

func (m *lanPolicyMockClient) PolicySocketLanMoveRule(_ context.Context, input cato_models.PolicyMoveRuleInput,
	accountID string, _ ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveRule, error) {
	m.moveRuleCalls = append(m.moveRuleCalls, lanPolicyMoveRuleCall{input: input, accountID: accountID})
	return &cato_go_sdk.PolicySocketLanMoveRule{
		Policy: &cato_go_sdk.PolicySocketLanMoveRule_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanMoveRule_Policy_SocketLan{
				MoveRule: cato_go_sdk.PolicySocketLanMoveRule_Policy_SocketLan_MoveRule{
					Status: m.responseMutationStatus(),
				},
			},
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
			SocketLan: &cato_go_sdk.PolicySocketLanFirewallMoveRule_Policy_SocketLan{
				Firewall: cato_go_sdk.PolicySocketLanFirewallMoveRule_Policy_SocketLan_Firewall{
					MoveRule: cato_go_sdk.PolicySocketLanFirewallMoveRule_Policy_SocketLan_Firewall_MoveRule{
						Status: m.responseMutationStatus(),
					},
				},
			},
		},
	}, nil
}

func (m *lanPolicyMockClient) PolicySocketLanPublishPolicyRevision(_ context.Context,
	_ *cato_models.SocketLanPolicyMutationInput, _ *cato_models.PolicyPublishRevisionInput, _ string,
	_ ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error) {
	m.publishCalls++
	return &cato_go_sdk.PolicySocketLanPublishPolicyRevision{
		Policy: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan{
				PublishPolicyRevision: cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan_PublishPolicyRevision{
					Status: m.responseMutationStatus(),
				},
			},
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
	sectionName   string
	subPolicyName string
	sectionIndex  int
}
type lanPolocyNetRuleData struct {
	id             string
	ruleType       string
	ruleName       string
	sectionName    string
	sectionKey     string
	indexInSection int
}

type lanPolocyFwRuleData struct {
	id               string
	firewallRuleName string
	netRuleName      string
	netRuleKey       string
	indexInRule      int
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

func duplicateLanPolicyPlan(withSectionKeys, withNetRuleKeys bool) lanPolicyPlanItem {
	mainSectionKey := ""
	childSectionKey := ""
	if withSectionKeys {
		mainSectionKey = "main-section"
		childSectionKey = "child-section"
	}
	mainNetRuleKey := ""
	childNetRuleKey := ""
	if withNetRuleKeys {
		mainNetRuleKey = "main-net"
		childNetRuleKey = "child-net"
	}
	return lanPolicyPlanItem{
		sectionData: map[string]lanPolicySectionData{
			"main-section": {
				sectionName:  "Shared section",
				sectionIndex: 1,
			},
			"child-section": {
				sectionName:   "Shared section",
				subPolicyName: "Child policy",
				sectionIndex:  1,
			},
		},
		networkRuleData: map[string]lanPolocyNetRuleData{
			"main-net": {
				ruleName:       "Shared network rule",
				sectionName:    "Shared section",
				sectionKey:     mainSectionKey,
				indexInSection: 1,
			},
			"child-net": {
				ruleName:       "Shared network rule",
				sectionName:    "Shared section",
				sectionKey:     childSectionKey,
				indexInSection: 1,
			},
		},
		firewallRuleData: map[string]lanPolocyFwRuleData{
			"main-fw": {
				firewallRuleName: "Shared firewall rule",
				netRuleName:      "Shared network rule",
				netRuleKey:       mainNetRuleKey,
				indexInRule:      1,
			},
			"child-fw": {
				firewallRuleName: "Shared firewall rule",
				netRuleName:      "Shared network rule",
				netRuleKey:       childNetRuleKey,
				indexInRule:      1,
			},
		},
	}
}

func singleLanPolicyPlan() lanPolicyPlanItem {
	return lanPolicyPlanItem{
		sectionData: map[string]lanPolicySectionData{
			"Shared section": {sectionIndex: 1},
		},
		networkRuleData: map[string]lanPolocyNetRuleData{
			"Shared network rule": {
				sectionName:    "Shared section",
				indexInSection: 1,
			},
		},
		firewallRuleData: map[string]lanPolocyFwRuleData{
			"Shared firewall rule": {
				netRuleName: "Shared network rule",
				indexInRule: 1,
			},
		},
	}
}

func (m *lanPolicyMockClient) createPlan(p lanPolicyPlanItem) *LanFwRulesIndex {
	ctx := context.Background()
	sectionObjects := make(map[string]types.Object, len(p.sectionData))
	for sectionName, section := range p.sectionData {
		sectionObject, diags := types.ObjectValueFrom(ctx, LanFwSectionDataTypes, LanFwSectionData{
			ID:            types.StringValue(section.id),
			SectionName:   optionalStringValue(section.sectionName),
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
			RuleName:       optionalStringValue(rule.ruleName),
			SectionName:    types.StringValue(rule.sectionName),
			SectionKey:     optionalStringValue(rule.sectionKey),
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
			ID:               types.StringValue(rule.id),
			FirewallRuleName: optionalStringValue(rule.firewallRuleName),
			NetRuleName:      types.StringValue(rule.netRuleName),
			NetRuleKey:       optionalStringValue(rule.netRuleKey),
			IndexInRule:      types.Int64Value(int64(rule.indexInRule)),
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

func duplicateLanPolicy() *cato_go_sdk.PolicySocketLanPolicy {
	return &cato_go_sdk.PolicySocketLanPolicy{
		Policy: &cato_go_sdk.PolicySocketLanPolicy_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan{
				Policy: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy{
					Enabled: true,
					Sections: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections{
						duplicateLanSection("s-main", "Shared section", nil),
						duplicateLanSection("s-child", "Shared section", ptr("sp-child")),
					},
					Rules: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules{
						duplicateLanRule("r-main", "Shared network rule", "s-main", "Shared section",
							"f-main", "Shared firewall rule"),
						duplicateLanRule("r-child", "Shared network rule", "s-child", "Shared section",
							"f-child", "Shared firewall rule"),
					},
					SubPolicies: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies{
						{
							Policy: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies_Policy{
								ID:   "sp-child",
								Name: "Child policy",
							},
						},
					},
				},
			},
		},
	}
}

func singleLanPolicy() *cato_go_sdk.PolicySocketLanPolicy {
	policy := duplicateLanPolicy()
	policyBase := &policy.Policy.SocketLan.Policy
	policyBase.Sections = policyBase.Sections[:1]
	policyBase.Rules = policyBase.Rules[:1]
	policyBase.SubPolicies = nil
	return policy
}

func addedAndRemovedLanPolicy() *cato_go_sdk.PolicySocketLanPolicy {
	policy := singleLanPolicy()
	policyBase := &policy.Policy.SocketLan.Policy
	policyBase.Sections = append(policyBase.Sections,
		duplicateLanSection("s-new", "New section", nil))
	policyBase.Rules = append(policyBase.Rules,
		duplicateLanRule("r-new", "New network rule", "s-new", "New section",
			"f-new", "New firewall rule"))
	return policy
}

func networkRulesMovedLanPolicy() *cato_go_sdk.PolicySocketLanPolicy {
	policy := duplicateLanPolicy()
	policyBase := &policy.Policy.SocketLan.Policy
	policyBase.Rules[0].Rule.Section.ID = "s-child"
	policyBase.Rules[1].Rule.Section.ID = "s-main"
	return policy
}

func firewallRulesMovedLanPolicy() *cato_go_sdk.PolicySocketLanPolicy {
	policy := duplicateLanPolicy()
	policyBase := &policy.Policy.SocketLan.Policy
	policyBase.Rules[0].Rule.Firewall, policyBase.Rules[1].Rule.Firewall =
		policyBase.Rules[1].Rule.Firewall, policyBase.Rules[0].Rule.Firewall
	return policy
}

func duplicateLanSection(id, name string,
	subPolicyID *string,
) *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections {
	return &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections{
		Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section{
			ID:          id,
			Name:        name,
			SubPolicyID: subPolicyID,
		},
	}
}

func duplicateLanRule(id, name, sectionID, sectionName, firewallID, firewallName string,
) *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules {
	return &cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules{
		RuleType: cato_models.PolicyRuleTypeEnumPolicyRule,
		Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule{
			ID:   id,
			Name: name,
			Section: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Section{
				ID:   sectionID,
				Name: sectionName,
			},
			Firewall: []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall{
				{
					Rule: cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule{
						ID:   firewallID,
						Name: firewallName,
					},
				},
			},
		},
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
