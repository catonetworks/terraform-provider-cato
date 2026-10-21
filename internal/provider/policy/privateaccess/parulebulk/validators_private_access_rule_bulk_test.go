package parulebulk

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPrivateAccessBulkPlanPublish(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &privAccessRuleBulkResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	for _, destroy := range []bool{false, true} {
		t.Run(map[bool]string{false: "increment publish", true: "destroy"}[destroy], func(t *testing.T) {
			plan := tfsdk.Plan{Schema: schemaResp.Schema}
			model := &PrivateAccessRuleBulkModel{ID: types.StringValue("0"), Publish: types.Int64Value(7), RuleData: types.MapNull(types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes})}
			if destroy {
				model = nil
			}
			require.False(t, plan.Set(ctx, model).HasError())
			resp := &resource.ModifyPlanResponse{Plan: plan}
			r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			if destroy {
				require.True(t, resp.Plan.Raw.IsNull())
				return
			}
			var got PrivateAccessRuleBulkModel
			require.False(t, resp.Plan.Get(ctx, &got).HasError())
			require.Equal(t, int64(8), got.Publish.ValueInt64())
			require.Equal(t, model.ID, got.ID)
			require.Equal(t, model.RuleData, got.RuleData)
		})
	}
}
