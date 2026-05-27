package planmodifiers

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// DHCPSettingsModifier returns a plan modifier for DHCP settings objects
// handle ID/Name for relay_group reference
func DHCPSettingsModifier() planmodifier.Object {
	return dhcpSettingsModifier{}
}

// dhcpSettingsModifier implements the plan modifier.
type dhcpSettingsModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m dhcpSettingsModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m dhcpSettingsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString implements the plan modification logic.
// If DHCP type is DHCP_RELAY, ensures that the relay group reference is properly handled (ID/Name)
//
//nolint:gocyclo
func (m dhcpSettingsModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	var cfg, state *tf.DhcpSettings
	var plan tf.DhcpSettings

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &cfg, basetypes.ObjectAsOptions{})) {
		return
	}
	if cfg == nil { // removed from the config
		resp.PlanValue = m.makeUnknownValue(ctx, &resp.Diagnostics) // API may return a value even if the config is removed
		return
	}

	// If not DHCP_RELAY type, do nothing
	dhcpType := cato_models.DhcpType(cfg.DhcpType.ValueString())
	if dhcpType != cato_models.DhcpTypeDhcpRelay {
		return
	}

	// Ensure there is exactly one name or id in the config
	if cfg.RelayGroupName.IsNull() && cfg.RelayGroupID.IsNull() {
		resp.Diagnostics.AddError("DHCP configuration error in "+req.Path.String(), "'relay_group_name' or 'relay_group_id' "+
			"must be defined in the config ")
	}
	if !cfg.RelayGroupName.IsNull() && !cfg.RelayGroupID.IsNull() {
		resp.Diagnostics.AddError("DHCP configuration error in "+req.Path.String(),
			fmt.Sprintf("only one of 'relay_group_name' or 'relay_group_id' can be specified in the config, "+
				"[relay_group_id:%q, relay_group_name:%q]",
				cfg.RelayGroupID.ValueString(), cfg.RelayGroupName.ValueString()))
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	if utils.CheckErr(&resp.Diagnostics, req.StateValue.As(ctx, &state, basetypes.ObjectAsOptions{})) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &plan, basetypes.ObjectAsOptions{})) {
		return
	}

	fmt.Println(plan) // TODO: is it filled in?

	// Name is configured
	if !cfg.RelayGroupName.IsNull() {
		// if Relay group name is in the state and it is the same, use the known ID value (if available)
		if state != nil && utils.HasValue(state.RelayGroupName) && state.RelayGroupName.ValueString() == cfg.RelayGroupName.ValueString() {
			resp.PlanValue = req.StateValue
			return
		}
		// Name is different -> set ID as unknown
		plan.RelayGroupName = cfg.RelayGroupName
		plan.RelayGroupID = types.StringUnknown()
		planObj, diag := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, plan)
		if utils.CheckErr(&resp.Diagnostics, diag) {
			return
		}
		resp.PlanValue = planObj
		return
	}

	// ID is configured
	// if ID is in the state and it is the same, use the known Name value (if available)
	if state != nil && utils.HasValue(state.RelayGroupID) && state.RelayGroupID.ValueString() == cfg.RelayGroupID.ValueString() {
		resp.PlanValue = req.StateValue
		return
	}
	// ID is different -> set Name as unknown
	plan.RelayGroupName = types.StringUnknown()
	plan.RelayGroupID = cfg.RelayGroupID
	planObj, diag := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, plan)
	if utils.CheckErr(&resp.Diagnostics, diag) {
		return
	}
	resp.PlanValue = planObj
}
func (m dhcpSettingsModifier) makeUnknownValue(ctx context.Context, diags *diag.Diagnostics) types.Object {
	plan := tf.DhcpSettings{
		DhcpType:              types.StringUnknown(),
		IPRange:               types.StringUnknown(),
		RelayGroupID:          types.StringUnknown(),
		RelayGroupName:        types.StringUnknown(),
		DhcpMicrosegmentation: types.BoolUnknown(),
	}

	planObj, diag := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, plan)
	if utils.CheckErr(diags, diag) {
		return types.ObjectNull(tf.DhcpSettingsAttrTypes)
	}
	return planObj
}
