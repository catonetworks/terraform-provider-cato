package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
			r.moveToPosition(ctx, currentRules, tc.id, "some name", tc.newPos)
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
