package alwaysonindex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ModifyPlan keeps observed ordering stable when neither ordering inputs nor
// publication intent changed. A new publication may observe dependent rule or
// section changes, so omitted maps must remain unknown until its read-back.
func (r *alwaysOnRuleBulkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var config, state, plan RulesIndex
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	unchanged := state.PublishTrigger.Equal(plan.PublishTrigger) && state.Changes.Equal(plan.Changes)
	unchanged = unchanged && (config.SectionData.IsNull() || state.SectionData.Equal(plan.SectionData)) &&
		(config.RuleData.IsNull() || state.RuleData.Equal(plan.RuleData))
	if config.SectionData.IsNull() {
		value := types.MapUnknown(SectionObjectType)
		if unchanged {
			value = state.SectionData
		}
		plan.SectionData = value
	}
	if config.RuleData.IsNull() {
		value := types.MapUnknown(RuleObjectType)
		if unchanged {
			value = state.RuleData
		}
		plan.RuleData = value
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

type modTrigger struct{}

func newModChanges() *modTrigger { return &modTrigger{} }

func (m *modTrigger) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	resp.PlanValue = types.Int64Value(0)
}

func (m *modTrigger) Description(_ context.Context) string {
	return "Modify policy trigger"
}
func (m *modTrigger) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
