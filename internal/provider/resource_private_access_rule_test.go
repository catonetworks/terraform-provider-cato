package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMove(t *testing.T) {
	ctx := context.Background()

	currentRules := []*PrivateAccessBulkRule{
		{ID: types.StringValue("r1"), Name: types.StringValue("Rule1"), Index: types.Int64Value(0), CMAIndex: 1},
		{ID: types.StringValue("r2"), Name: types.StringValue("Rule2"), Index: types.Int64Value(1), CMAIndex: 2},
		{ID: types.StringValue("r3"), Name: types.StringValue("Rule3"), Index: types.Int64Value(2), CMAIndex: 3},
		{ID: types.StringValue("r4"), Name: types.StringValue("Rule4"), Index: types.Int64Value(3), CMAIndex: 4},
	}

	r := privAccessRuleBulkResource{}
	r.moveToPosition(ctx, currentRules, "r3", "Rule3", 0)
}
