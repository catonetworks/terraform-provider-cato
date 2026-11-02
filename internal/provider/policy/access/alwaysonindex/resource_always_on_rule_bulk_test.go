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

func TestAlwaysOnRuleBulkCreateMovesSection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockClient := mocks.NewAlwaysOnRuleBulkClient(t)
	r := &alwaysOnRuleBulkResource{
		client:     &catoclient.Service{AccountId: "account-123"},
		mockClient: mockClient,
	}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	planned := sectionOrderModel(t, map[string]Section{
		"Alpha": {ID: types.StringValue("s1"), Index: types.Int64Value(2)},
		"Beta":  {ID: types.StringValue("s2"), Index: types.Int64Value(1)},
	})
	planned.RuleData = types.MapNull(RuleObjectType)
	planned.Changes = types.Int64Unknown()
	planned.PublishTrigger = types.StringValue("initial-policy-hash")
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	require.False(t, plan.Set(ctx, &planned).HasError())

	var currentPolicy cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{
		"revision":{"changes":0},
		"sections":[{"section":{"id":"s1","name":"Alpha"}},{"section":{"id":"s2","name":"Beta"}}],
		"rules":[]
	}}}}`), &currentPolicy))
	var readInput *cato_models.ZtnaAlwaysOnPolicyInput
	initialReadCall := mockClient.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", readInput).
		Return(&currentPolicy, nil).Once()

	var policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput
	moveCall := mockClient.EXPECT().PolicyZtnaAlwaysOnMoveSection(ctx, "account-123", policyInput, cato_models.PolicyMoveSectionInput{
		ID: "s1",
		To: &cato_models.PolicySectionPositionInput{Position: cato_models.PolicySectionPositionEnumLastInPolicy},
	}).Return(&cato.PolicyZtnaAlwaysOnMoveSection{}, nil).Once()

	var publishResult cato.PolicyZtnaAlwaysOnPublishPolicyRevision
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"SUCCESS"}}}}`),
		&publishResult))
	var publishInput *cato_models.PolicyPublishRevisionInput
	publishCall := mockClient.EXPECT().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, "account-123", policyInput, publishInput).
		Return(&publishResult, nil).Once()

	var policyResult cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{
		"revision":{"changes":0},
		"sections":[{"section":{"id":"s2","name":"Beta"}},{"section":{"id":"s1","name":"Alpha"}}],
		"rules":[]
	}}}}`), &policyResult))
	readCall := mockClient.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", readInput).
		Return(&policyResult, nil).Once()
	mock.InOrder(initialReadCall, moveCall, publishCall, readCall)

	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: plan.Raw},
		Plan:   plan,
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var state RulesIndex
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, planned.SectionData, state.SectionData)
	require.Equal(t, planned.PublishTrigger, state.PublishTrigger)
	require.Equal(t, types.Int64Value(0), state.Changes)
	require.Empty(t, state.RuleData.Elements())
}

func TestAlwaysOnRuleBulkReadPreservesPublishTrigger(t *testing.T) {
	t.Parallel()
	for _, trigger := range []types.String{types.StringNull(), types.StringValue(""), types.StringValue("policy-hash")} {
		t.Run(trigger.String(), func(t *testing.T) {
			t.Parallel()
			r, client, state := publishTriggerFixture(t)
			var prior RulesIndex
			require.False(t, state.Get(ctx, &prior).HasError())
			prior.PublishTrigger = trigger
			require.False(t, state.Set(ctx, prior).HasError())
			client.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(publishTriggerPolicy(t), nil).Once()

			resp := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var actual RulesIndex
			require.False(t, resp.State.Get(ctx, &actual).HasError())
			require.Equal(t, trigger, actual.PublishTrigger)
		})
	}
}

func TestAlwaysOnRuleBulkUpdatePublishTriggerOnly(t *testing.T) {
	t.Parallel()
	for _, trigger := range []types.String{types.StringNull(), types.StringValue(""), types.StringValue("updated-policy-hash")} {
		t.Run(trigger.String(), func(t *testing.T) {
			t.Parallel()
			r, client, state := publishTriggerFixture(t)
			var planned RulesIndex
			require.False(t, state.Get(ctx, &planned).HasError())
			planned.PublishTrigger = trigger
			plan := tfsdk.Plan{Schema: state.Schema}
			require.False(t, plan.Set(ctx, planned).HasError())

			initialRead := client.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(publishTriggerPolicy(t), nil).Once()
			var published cato.PolicyZtnaAlwaysOnPublishPolicyRevision
			require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"SUCCESS"}}}}`), &published))
			publish := client.EXPECT().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, "account-123",
				(*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil), (*cato_models.PolicyPublishRevisionInput)(nil)).
				Return(&published, nil).Once()
			finalRead := client.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(publishTriggerPolicy(t), nil).Once()
			mock.InOrder(initialRead, publish, finalRead)

			resp := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{
				State: state, Plan: plan, Config: tfsdk.Config{Schema: state.Schema, Raw: plan.Raw},
			}, &resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var actual RulesIndex
			require.False(t, resp.State.Get(ctx, &actual).HasError())
			require.Equal(t, planned, actual)
		})
	}
}

func TestAlwaysOnRuleBulkPublishFailurePreservesTrigger(t *testing.T) {
	t.Parallel()
	r, client, state := publishTriggerFixture(t)
	var planned RulesIndex
	require.False(t, state.Get(ctx, &planned).HasError())
	planned.PublishTrigger = types.StringValue("updated-policy-hash")
	plan := tfsdk.Plan{Schema: state.Schema}
	require.False(t, plan.Set(ctx, planned).HasError())
	initialRead := client.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
		Return(publishTriggerPolicy(t), nil).Once()
	publish := client.EXPECT().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, "account-123",
		(*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil), (*cato_models.PolicyPublishRevisionInput)(nil)).
		Return(nil, errors.New("publish failed")).Once()
	mock.InOrder(initialRead, publish)

	resp := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{
		State: state, Plan: plan, Config: tfsdk.Config{Schema: state.Schema, Raw: plan.Raw},
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "publish failed")
	require.True(t, state.Raw.Equal(resp.State.Raw), "failed publishing must not acknowledge the new trigger")
}

func publishTriggerFixture(t *testing.T) (*alwaysOnRuleBulkResource, *mocks.AlwaysOnRuleBulkClient, tfsdk.State) {
	t.Helper()
	client := mocks.NewAlwaysOnRuleBulkClient(t)
	r := &alwaysOnRuleBulkResource{client: &catoclient.Service{AccountId: "account-123"}, mockClient: client}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	var diags diag.Diagnostics
	model := parseRulesIndex(ctx, publishTriggerPolicy(t), &diags)
	require.False(t, diags.HasError(), "%v", diags)
	model.PublishTrigger = types.StringValue("previous-policy-hash")
	state := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, state.Set(ctx, model).HasError())
	return r, client, state
}

func publishTriggerPolicy(t *testing.T) *cato.PolicyZtnaAlwaysOnPolicy {
	t.Helper()
	var policy cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{
		"revision":{"changes":0},
		"sections":[{"section":{"id":"s1","name":"Alpha"}}],
		"rules":[{"rule":{"id":"r1","name":"First","section":{"id":"s1","name":"Alpha"}}}]
	}}}}`), &policy))
	return &policy
}

func TestAlwaysOnRuleBulkCreateMovesRule(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockClient := mocks.NewAlwaysOnRuleBulkClient(t)
	r := &alwaysOnRuleBulkResource{
		client:     &catoclient.Service{AccountId: "account-123"},
		mockClient: mockClient,
	}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	current := sectionOrderModel(t, map[string]Section{
		"Alpha": {ID: types.StringValue("s1"), Index: types.Int64Value(1)},
	})
	currentRules, ruleDiags := types.MapValueFrom(ctx, RuleObjectType, map[string]Rule{
		"First": {
			ID: types.StringValue("r1"), IndexInSection: types.Int64Value(1), SectionName: types.StringValue("Alpha"),
		},
		"Second": {
			ID: types.StringValue("r2"), IndexInSection: types.Int64Value(2), SectionName: types.StringValue("Alpha"),
		},
	})
	require.False(t, ruleDiags.HasError(), "%v", ruleDiags)
	current.RuleData = currentRules
	current.Changes = types.Int64Null()
	planned := current
	planned.RuleData, ruleDiags = types.MapValueFrom(ctx, RuleObjectType, map[string]Rule{
		"First": {
			ID: types.StringValue("r1"), IndexInSection: types.Int64Value(2), SectionName: types.StringValue("Alpha"),
		},
		"Second": {
			ID: types.StringValue("r2"), IndexInSection: types.Int64Value(1), SectionName: types.StringValue("Alpha"),
		},
	})
	require.False(t, ruleDiags.HasError(), "%v", ruleDiags)
	planned.Changes = types.Int64Unknown()
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	require.False(t, plan.Set(ctx, &planned).HasError())

	var currentPolicy cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{
		"revision":{"changes":0},
		"sections":[{"section":{"id":"s1","name":"Alpha"}}],
		"rules":[
			{"rule":{"id":"r1","name":"First","section":{"id":"s1","name":"Alpha"}}},
			{"rule":{"id":"r2","name":"Second","section":{"id":"s1","name":"Alpha"}}}
		]
	}}}}`), &currentPolicy))
	var readInput *cato_models.ZtnaAlwaysOnPolicyInput
	initialReadCall := mockClient.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", readInput).
		Return(&currentPolicy, nil).Once()

	var policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput
	moveCall := mockClient.EXPECT().PolicyZtnaAlwaysOnMoveRule(ctx, "account-123", policyInput, cato_models.PolicyMoveRuleInput{
		ID: "r1",
		To: &cato_models.PolicyRulePositionInput{
			Position: new(cato_models.PolicyRulePositionEnumLastInSection),
			Ref:      new("s1"),
		},
	}).Return(&cato.PolicyZtnaAlwaysOnMoveRule{}, nil).Once()

	var publishResult cato.PolicyZtnaAlwaysOnPublishPolicyRevision
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"SUCCESS"}}}}`),
		&publishResult))
	var publishInput *cato_models.PolicyPublishRevisionInput
	publishCall := mockClient.EXPECT().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, "account-123", policyInput, publishInput).
		Return(&publishResult, nil).Once()

	var policyResult cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{
		"revision":{"changes":0},
		"sections":[{"section":{"id":"s1","name":"Alpha"}}],
		"rules":[
			{"rule":{"id":"r2","name":"Second","section":{"id":"s1","name":"Alpha"}}},
			{"rule":{"id":"r1","name":"First","section":{"id":"s1","name":"Alpha"}}}
		]
	}}}}`), &policyResult))
	readCall := mockClient.EXPECT().PolicyZtnaAlwaysOnPolicy(ctx, "account-123", readInput).
		Return(&policyResult, nil).Once()
	mock.InOrder(initialReadCall, moveCall, publishCall, readCall)

	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: plan.Raw},
		Plan:   plan,
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var state RulesIndex
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, planned.SectionData, state.SectionData)
	require.Equal(t, planned.RuleData, state.RuleData)
	require.Equal(t, types.Int64Value(0), state.Changes)
}
