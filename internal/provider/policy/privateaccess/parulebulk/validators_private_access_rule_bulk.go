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

	plan.Publish = types.Int64Value(plan.Publish.ValueInt64() + 1)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
