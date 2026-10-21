package networkrange

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

// ModifyPlan ensures that exactly one of interface_id or interface_index is set,
// and if one changes in the config, the other one is marked as unknown.
func (r *networkRangeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, cfg *tf.NetworkRange
	state := &tf.NetworkRange{} // avoid nil pointer dereference
	stateDefined := !req.State.Raw.IsNull()

	if req.Plan.Raw.IsNull() { // resource destruction
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if stateDefined {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Validate config - ensure there is exactly one interface_index or interface_id.
	nrValidator := NetworkRangeValidator{}
	nrValidator.ValidateNetworkRangeWithPriorState(ctx, cfg, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())

	// mdns reflector is only relevant for native, direct and vlan ranges
	if rangeType == cato_models.SubnetTypeRouted {
		plan.MdnsReflector = types.BoolNull()
		if utils.HasValue(cfg.MdnsReflector) {
			resp.Diagnostics.AddError("Invalid network range configuration",
				"mdns_reflector cannot be used when rangeType is 'Routed'")
			return
		}
	}

	// set interfaceIndex and interfaceID
	r.planInterfaceIDIndex(cfg, plan, state, stateDefined)

	// Gateway is only relevant for Routed range type
	plan.Gateway = types.StringNull()
	if rangeType == cato_models.SubnetTypeRouted {
		plan.Gateway = planmodifiers.DefaultPlanValue(cfg.Gateway, state.Gateway, stateDefined)
	}

	// Local IP is only relevant for Direct, Native and VLAN range types
	plan.LocalIP = types.StringNull()
	if rangeType != cato_models.SubnetTypeRouted {
		plan.LocalIP = planmodifiers.DefaultPlanValue(cfg.LocalIP, state.LocalIP, stateDefined)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// planInterfaceIDIndex - plan mutex interface fields
// - Ensures that exactly one of interface_id or interface_index is set in the plan.
// - If one changes in the config, the other one is marked as unknown.
// - If neither changes in the config, the known value from state is used in the plan (if available)
func (r *networkRangeResource) planInterfaceIDIndex(cfg, plan, state *tf.NetworkRange, stateDefined bool) {
	if !stateDefined {
		return
	}

	indexExplicit := validators.StringFieldIsExplicit(cfg.InterfaceIndex, state.InterfaceIndex)
	idExplicit := validators.StringFieldIsExplicit(cfg.InterfaceID, state.InterfaceID)

	// if interfaceIndex is configured and is different from the state, mark interfaceID as unknown
	if indexExplicit {
		plan.InterfaceIndex = cfg.InterfaceIndex
		plan.InterfaceID = types.StringUnknown()
		return
	}
	// if interfaceID is configured and is different from the state, mark interfaceIndex as unknown
	if idExplicit {
		plan.InterfaceID = cfg.InterfaceID
		plan.InterfaceIndex = types.StringUnknown()
		return
	}

	if utils.HasValue(cfg.InterfaceIndex) {
		plan.InterfaceIndex = cfg.InterfaceIndex
		plan.InterfaceID = state.InterfaceID
	}
	if utils.HasValue(cfg.InterfaceID) {
		plan.InterfaceID = cfg.InterfaceID
		plan.InterfaceIndex = state.InterfaceIndex
	}
}
