package parulebulk

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestMove(t *testing.T) {
	type tc struct {
		id          string
		newPos      int
		expectedIDs []string
	}
	tcs := []tc{
		{id: "r1", newPos: 0, expectedIDs: []string{"r1", "r2", "r3", "r4"}},
		{id: "r3", newPos: 1, expectedIDs: []string{"r1", "r3", "r2", "r4"}},
		{id: "r4", newPos: 3, expectedIDs: []string{"r1", "r2", "r3", "r4"}},
		{id: "r4", newPos: 1, expectedIDs: []string{"r1", "r4", "r2", "r3"}},
		{id: "r4", newPos: 0, expectedIDs: []string{"r4", "r1", "r2", "r3"}},
		{id: "r1", newPos: 1, expectedIDs: []string{"r1", "r1", "r3", "r4"}}, // moving down is not supported!
	}

	ctx := context.Background()
	r := privAccessRuleBulkResource{}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%s-%d", tc.id, tc.newPos), func(t *testing.T) {
			currentRules := getTestRules()
			if err := r.moveToPosition(ctx, currentRules, tc.id, "some name", tc.newPos); err != nil {
				t.Fatalf("moveToPosition returned error: %v", err)
			}
			if len(currentRules) != len(tc.expectedIDs) {
				t.Errorf("Expected %d rules, got %d", len(tc.expectedIDs), len(currentRules))
				return
			}
			for j, rule := range currentRules {
				if rule.ID.ValueString() != tc.expectedIDs[j] {
					t.Errorf("Expected rule %d to be '%s', got '%s'", j, tc.expectedIDs[j], rule.ID.String())
				}
			}
		})
	}
}

func getTestRules() []*PrivateAccessBulkRule {
	return []*PrivateAccessBulkRule{
		{ID: types.StringValue("r1"), Name: types.StringValue("Rule1")},
		{ID: types.StringValue("r2"), Name: types.StringValue("Rule2")},
		{ID: types.StringValue("r3"), Name: types.StringValue("Rule3")},
		{ID: types.StringValue("r4"), Name: types.StringValue("Rule4")},
	}
}

func TestBulkHydrationWithUnmanagedRules(t *testing.T) {
	ctx := context.Background()
	var response sdk.PolicyReadPrivateAccessPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"privateAccess":{"policy":{"rules":[{"rule":{"id":"external-id","name":"external","index":0}},{"rule":{"id":"managed-id","name":"managed","index":1}}]}}}}`), &response))
	client := mocks.NewPrivateAccessBulkResourceClient(t)
	client.EXPECT().PolicyReadPrivateAccessPolicy(mock.Anything, "account-123").Return(&response, nil).Twice()
	r := &privAccessRuleBulkResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: client}
	rule := types.ObjectValueMust(PrivateAccessBulkRuleTypes, map[string]attr.Value{"id": types.StringUnknown(), "name": types.StringValue("managed"), "index": types.Int64Value(0)})
	plan := &PrivateAccessRuleBulkModel{RuleData: types.MapValueMust(types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes}, map[string]attr.Value{"managed": rule})}
	state, rules, diags, err := r.hydratePrivAccessRuleBulkState(ctx, plan)
	require.NoError(t, err)
	require.False(t, diags.HasError())
	require.Len(t, state.RuleData.Elements(), 1)
	require.Len(t, rules, 2)
	require.Less(t, rules["managed"].Index.ValueInt64(), rules["external"].Index.ValueInt64())
	var moved sdk.PolicyPrivateAccessMoveRule
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"privateAccess":{"moveRule":{"status":"SUCCESS"}}}}`), &moved))
	client.EXPECT().PolicyPrivateAccessMoveRule(mock.Anything, "account-123", mock.MatchedBy(func(input m.PolicyMoveRuleInput) bool {
		return input.ID == "managed-id" && input.To != nil && input.To.Position != nil && *input.To.Position == m.PolicyRulePositionEnumFirstInPolicy
	})).Return(&moved, nil).Once()
	require.False(t, r.moveRules(ctx, rules).HasError())
	response.Policy.PrivateAccess.Policy.Rules = response.Policy.PrivateAccess.Policy.Rules[:1]
	state, rules, diags, err = r.hydratePrivAccessRuleBulkState(ctx, plan)
	require.Nil(t, state)
	require.Nil(t, rules)
	require.False(t, diags.HasError())
	require.ErrorContains(t, err, "managed")
}
