package idname

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// PlanModifier returns a plan modifier for ID/Name reference objects
func PlanModifier() planmodifier.Object {
	return idNamePlanModifier{}
}

// idNamePlanModifier implements the plan modifier.
type idNamePlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m idNamePlanModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m idNamePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString implements the plan modification logic.
//
//nolint:gocyclo
func (m idNamePlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	var cfg, state *Model
	var plan Model

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &cfg, basetypes.ObjectAsOptions{})) {
		return
	}
	if cfg == nil { // removed from the config
		if req.StateValue.IsNull() {
			return
		}
		resp.PlanValue = types.ObjectNull(ModelTypes)
		return
	}

	// Ensure there is exactly one name or id in the config
	if cfg.Name.IsNull() && cfg.ID.IsNull() {
		resp.Diagnostics.AddError("idName reference error in "+req.Path.String(), "'name' or 'id' must be defined in the config ")
	}
	if !cfg.Name.IsNull() && !cfg.ID.IsNull() {
		resp.Diagnostics.AddError("idName reference error in "+req.Path.String(),
			fmt.Sprintf("only one of 'name' or 'id' can be specified in the config, [id:%q, name:%q]", cfg.ID.ValueString(), cfg.Name.ValueString()))
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	if utils.CheckErr(&resp.Diagnostics, req.StateValue.As(ctx, &state, basetypes.ObjectAsOptions{})) {
		return
	}

	// Name is configured
	if !cfg.Name.IsNull() {
		// if Name is in the state and it is the same, use the known ID value (if available)
		if state != nil && utils.HasValue(state.Name) && state.Name.ValueString() == cfg.Name.ValueString() {
			resp.PlanValue = req.StateValue
			return
		}
		// Name is different -> set ID as unknown
		plan.Name = cfg.Name
		plan.ID = types.StringUnknown()
		planObj, diag := types.ObjectValueFrom(ctx, ModelTypes, plan)
		if utils.CheckErr(&resp.Diagnostics, diag) {
			return
		}
		resp.PlanValue = planObj
		return
	}

	// ID is configured
	// if ID is in the state and it is the same, use the known Name value (if available)
	if state != nil && utils.HasValue(state.ID) && state.ID.ValueString() == cfg.ID.ValueString() {
		resp.PlanValue = req.StateValue
		return
	}
	// ID is different -> set Name as unknown
	plan.Name = types.StringUnknown()
	plan.ID = cfg.ID
	planObj, diag := types.ObjectValueFrom(ctx, ModelTypes, plan)
	if utils.CheckErr(&resp.Diagnostics, diag) {
		return
	}
	resp.PlanValue = planObj
}
