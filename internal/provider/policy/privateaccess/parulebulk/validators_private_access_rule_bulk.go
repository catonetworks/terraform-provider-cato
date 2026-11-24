package parulebulk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *privAccessRuleBulkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *PrivateAccessRuleBulkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan == nil {
		return
	}

	// Only bump the publish counter when something actually changes. Bumping it
	// unconditionally made every plan show "publish = N -> N+1", a perpetual
	// in-place update on a resource whose rules had not changed.
	var state *PrivateAccessRuleBulkModel
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if state != nil && sameOrdering(ctx, plan.RuleData, state.RuleData) {
		plan.Publish = state.Publish
	} else {
		plan.Publish = types.Int64Value(plan.Publish.ValueInt64() + 1)
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// sameOrdering compares only what the configuration controls (rule name ->
// index); the computed rule ids can be unknown at plan time.
func sameOrdering(ctx context.Context, a, b types.Map) bool {
	if a.IsUnknown() || b.IsUnknown() || a.IsNull() != b.IsNull() {
		return false
	}
	var am, bm map[string]PrivateAccessBulkRule
	if a.ElementsAs(ctx, &am, false).HasError() || b.ElementsAs(ctx, &bm, false).HasError() {
		return false
	}
	if len(am) != len(bm) {
		return false
	}
	for k, av := range am {
		bv, ok := bm[k]
		if !ok || av.Index.IsUnknown() || !av.Index.Equal(bv.Index) {
			return false
		}
	}
	return true
}
