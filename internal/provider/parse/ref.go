package parse

import (
	"context"
	"fmt"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type IdNameRefModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var IdNameRefModelTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

func SchemaNameID(prefix string) map[string]schema.Attribute {
	if prefix != "" {
		prefix += " "
	}
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: prefix + "name",
			Optional:    true,
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: prefix + "ID",
			Optional:    true,
			Computed:    true,
		},
	}
}

// UseStateForUnknown returns a plan modifier that copies a known prior state
// value into the planned value. Use this when it is known that an unconfigured
// value will remain the same after a resource update.
//
// To prevent Terraform errors, the framework automatically sets unconfigured
// and Computed attributes to an unknown value "(known after apply)" on update.
// Using this plan modifier will instead display the prior state value in the
// plan, unless a prior plan modifier adjusts the value.
func IdNameModifier() planmodifier.Object {
	return idNamePlanModifier{}
}

// idNamePlanModifier implements the plan modifier.
type idNamePlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m idNamePlanModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m idNamePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyString implements the plan modification logic.
func (m idNamePlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	var cfg, state, plan IdNameRefModel
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &cfg, basetypes.ObjectAsOptions{})) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, req.StateValue.As(ctx, &state, basetypes.ObjectAsOptions{})) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, req.PlanValue.As(ctx, &plan, basetypes.ObjectAsOptions{})) {
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

	// Name is configured
	if !cfg.Name.IsNull() {
		// if Name is in the state and it is the same, use the known ID value (if available)
		if utils.HasValue(state.Name) && state.Name.ValueString() == cfg.Name.ValueString() {
			resp.PlanValue = req.StateValue
			return
		}
		// Name is different -> set ID as unknown
		plan.Name = cfg.Name
		plan.ID = types.StringUnknown()
		planObj, diag := types.ObjectValueFrom(ctx, IdNameRefModelTypes, plan)
		if utils.CheckErr(&resp.Diagnostics, diag) {
			return
		}
		resp.PlanValue = planObj
		return
	}

	// ID is configured
	// if ID is in the state and it is the same, use the known Name value (if available)
	if utils.HasValue(state.ID) && state.ID.ValueString() == cfg.ID.ValueString() {
		resp.PlanValue = req.StateValue
		return
	}
	// ID is different -> set Name as unknown
	plan.Name = types.StringUnknown()
	plan.ID = cfg.ID
	planObj, diag := types.ObjectValueFrom(ctx, IdNameRefModelTypes, plan)
	if utils.CheckErr(&resp.Diagnostics, diag) {
		return
	}
	resp.PlanValue = planObj
}
