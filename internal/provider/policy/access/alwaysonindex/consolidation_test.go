package alwaysonindex

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func twoSectionOrdering(t *testing.T, swapped bool) RulesIndex {
	t.Helper()
	state := sectionOrderModel(t, map[string]Section{
		"Alpha": {ID: types.StringValue("s1"), Index: types.Int64Value(1)},
		"Beta":  {ID: types.StringValue("s2"), Index: types.Int64Value(2)},
	})
	a, b := "Alpha", "Beta"
	if swapped {
		a, b = b, a
	}
	rules, diagnostics := types.MapValueFrom(context.Background(), RuleObjectType, map[string]Rule{
		"First":  {ID: types.StringValue("r1"), IndexInSection: types.Int64Value(1), SectionName: types.StringValue(a)},
		"Second": {ID: types.StringValue("r2"), IndexInSection: types.Int64Value(1), SectionName: types.StringValue(b)},
	})
	require.False(t, diagnostics.HasError())
	state.RuleData = rules
	return state
}

func TestOmittedOrderingPreservesLiveOrder(t *testing.T) {
	t.Parallel()
	state := twoSectionOrdering(t, false)
	config := state
	config.SectionData = types.MapNull(SectionObjectType)
	config.RuleData = types.MapNull(RuleObjectType)
	var diagnostics diag.Diagnostics
	sections, ids := buildSectionOrder(context.Background(), state, config, &diagnostics)
	rules := buildRuleOrder(context.Background(), state, config, ids, &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.Equal(t, sections.current, sections.target)
	for _, order := range rules {
		require.Equal(t, order.current, order.target)
	}
}

func TestBulkMovesRulesBetweenSections(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diagnostics diag.Diagnostics
	config := twoSectionOrdering(t, true)
	config.SectionData = types.MapNull(SectionObjectType)
	orders := buildRuleOrder(ctx, twoSectionOrdering(t, false), config,
		map[string]string{"Alpha": "s1", "Beta": "s2"}, &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	client := mocks.NewAlwaysOnRuleBulkClient(t)
	var response cato.PolicyZtnaAlwaysOnMoveRule
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"moveRule":{"status":"SUCCESS"}}}}`), &response))
	for _, move := range []struct{ id, section string }{{"r2", "s1"}, {"r1", "s2"}} {
		client.EXPECT().PolicyZtnaAlwaysOnMoveRule(ctx, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
			mock.MatchedBy(func(input cato_models.PolicyMoveRuleInput) bool {
				return input.ID == move.id && input.To != nil && input.To.Ref != nil && *input.To.Ref == move.section &&
					input.To.Position != nil && *input.To.Position == cato_models.PolicyRulePositionEnumLastInSection
			})).Return(&response, nil).Once()
	}
	r := &alwaysOnRuleBulkResource{client: &catoclient.Service{AccountId: "123"}, mockClient: client}
	r.moveRules(ctx, orders, &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
}

func TestRelocationRejectsMissingDuplicateAndUnknownRules(t *testing.T) {
	t.Parallel()
	for _, target := range []map[string]map[string]int64{
		{"Beta": {"First": 1}},
		{"Alpha": {"First": 1}, "Beta": {"First": 1, "Second": 2}},
		{"Missing": {"First": 1, "Second": 2}},
		{"Beta": {"First": 1, "Unknown": 2}},
	} {
		var diagnostics diag.Diagnostics
		validateRuleRelocation(map[string]string{"Alpha": "s1", "Beta": "s2"},
			map[string]map[string]int64{"Alpha": {"First": 1}, "Beta": {"Second": 1}}, target, &diagnostics)
		require.True(t, diagnostics.HasError())
	}
}

func TestAlwaysOnBulkRetriesTransientRead(t *testing.T) {
	t.Parallel()
	r, client, _ := publishTriggerFixture(t)
	client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "account-123", mock.Anything).
		Return(nil, errors.New("DOWNSTREAM_SERVICE_ERROR")).Once()
	client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "account-123", mock.Anything).
		Return(publishTriggerPolicy(t), nil).Once()
	var diagnostics diag.Diagnostics
	require.NotNil(t, r.readPolicy(context.Background(), &diagnostics))
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
}

func TestAlwaysOnBulkPublicationRetryHonorsCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, client, _ := publishTriggerFixture(t)
	client.On("PolicyZtnaAlwaysOnPublishPolicyRevision", mock.Anything, "account-123", mock.Anything, mock.Anything).
		Return(nil, errors.New("reorderPolicyBlockedByActiveSessions")).Run(func(mock.Arguments) { cancel() }).Once()
	var diagnostics diag.Diagnostics
	r.publish(ctx, &diagnostics)
	require.True(t, diagnostics.HasError())
	require.Contains(t, diagnostics.Errors()[0].Detail(), "context canceled")
}

func TestAlwaysOnOrderingMapsAreOptionalComputed(t *testing.T) {
	t.Parallel()
	var response resource.SchemaResponse
	NewAlwaysOnRuleBulkResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	for _, name := range []string{"section_data", "rule_data"} {
		attribute := response.Schema.Attributes[name]
		require.True(t, attribute.IsOptional())
		require.True(t, attribute.IsComputed())
	}
}

func TestRuleOrderingChangeKeepsOmittedSectionsUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	r := &alwaysOnRuleBulkResource{}
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	before, after := twoSectionOrdering(t, false), twoSectionOrdering(t, true)
	before.Changes, after.Changes = types.Int64Value(0), types.Int64Value(0)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	require.False(t, state.Set(ctx, before).HasError())
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	after.SectionData = types.MapUnknown(SectionObjectType)
	require.False(t, plan.Set(ctx, after).HasError())
	after.SectionData = types.MapNull(SectionObjectType)
	config := tfsdk.State{Schema: schemaResponse.Schema}
	require.False(t, config.Set(ctx, after).HasError())
	response := resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan: plan, State: state, Config: tfsdk.Config(config),
	}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var actual RulesIndex
	require.False(t, response.Plan.Get(ctx, &actual).HasError())
	require.True(t, actual.SectionData.IsUnknown(), "changed ordering may observe dependent section changes after apply")
}
