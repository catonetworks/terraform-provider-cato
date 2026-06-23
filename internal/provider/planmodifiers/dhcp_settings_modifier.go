package planmodifiers

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var dhcpSettingNull = types.ObjectNull(tf.DhcpSettingsAttrTypes)

// DHCPSettingsModifier returns a plan modifier for DHCP settings objects
// handle ID/Name for relay_group reference
func DHCPSettingsModifier(isRangeResource bool) planmodifier.Object {
	return dhcpSettingsModifier{isRangeResource: isRangeResource}
}

// dhcpSettingsModifier implements the plan modifier.
type dhcpSettingsModifier struct {
	isRangeResource bool // if true, take rangeType from  `/range_type', otherwise '/native_range/range_type'
}

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
func (m dhcpSettingsModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	var cfg, state *tf.DhcpSettings

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &cfg, basetypes.ObjectAsOptions{})) {
		return
	}
	if cfg == nil { // removed from the config, use the default
		resp.PlanValue = m.planDhcpDefault(ctx, req, &resp.Diagnostics)
		return
	}
	if !req.StateValue.IsNull() && utils.CheckErr(&resp.Diagnostics, req.StateValue.As(ctx, &state, basetypes.ObjectAsOptions{})) {
		return
	}

	dhcpType := cato_models.DhcpType(cfg.DhcpType.ValueString())

	// microsegmentation is only relevant for DHCP_RANGE
	if dhcpType != cato_models.DhcpTypeDhcpRange && utils.HasValue(cfg.DhcpMicrosegmentation) {
		resp.Diagnostics.AddError("configuration error in dhcp_settings",
			"'dhcp_microsegmentation' can only be set when 'dhcp_type' is 'DHCP_RANGE'")
		return
	}

	// DHCPSettings configured - different rules depending on the DHCP type
	switch dhcpType {
	case cato_models.DhcpTypeDhcpRelay:
		resp.PlanValue = m.planDhcpRelay(ctx, state, cfg, &resp.Diagnostics)
		return
	case cato_models.DhcpTypeDhcpRange:
		resp.PlanValue = m.planDhcpRange(ctx, state, cfg, &resp.Diagnostics)
		return
	case cato_models.DhcpTypeAccountDefault, cato_models.DhcpTypeDhcpDisabled:
		resp.PlanValue = m.planDhcpEmpty(ctx, dhcpType.String(), &resp.Diagnostics)
		return
	default:
		resp.Diagnostics.AddError("Unsupported DHCP type", "Unknown DHCP type: "+dhcpType.String())
		return
	}
}

func (m dhcpSettingsModifier) planDhcpRelay(ctx context.Context, state, cfg *tf.DhcpSettings, diags *diag.Diagnostics) types.Object {
	// Ensure there is exactly one name or id in the config
	if cfg.RelayGroupName.IsNull() && cfg.RelayGroupID.IsNull() {
		diags.AddError("DHCP configuration error in dhcp_settings", "'relay_group_name' or 'relay_group_id' "+
			"must be defined in the config ")
		return dhcpSettingNull
	}
	if !cfg.RelayGroupName.IsNull() && !cfg.RelayGroupID.IsNull() {
		diags.AddError("DHCP configuration error in dhcp_settings",
			fmt.Sprintf("only one of 'relay_group_name' or 'relay_group_id' can be specified in the config, "+
				"[relay_group_id:%q, relay_group_name:%q]",
				cfg.RelayGroupID.ValueString(), cfg.RelayGroupName.ValueString()))
		return dhcpSettingNull
	}

	plan := tf.DhcpSettings{
		DhcpType:              cfg.DhcpType,       // required
		IPRange:               types.StringNull(), // only for DHCP_RANGE
		DhcpMicrosegmentation: types.BoolNull(),   // only for DHCP_RANGE
		RelayGroupID:          types.StringUnknown(),
		RelayGroupName:        types.StringUnknown(),
	}
	// RelayGroup Name configured
	if !cfg.RelayGroupName.IsNull() {
		plan.RelayGroupName = cfg.RelayGroupName
		// if the name is the same as state, use known ID value (if available)
		if state != nil && utils.HasValue(state.RelayGroupName) &&
			state.RelayGroupName.ValueString() == cfg.RelayGroupName.ValueString() {
			plan.RelayGroupID = state.RelayGroupID
		}
	}
	if !cfg.RelayGroupID.IsNull() {
		plan.RelayGroupID = cfg.RelayGroupID
		// if the id is the same as state, use known Name value (if available)
		if state != nil && utils.HasValue(state.RelayGroupID) &&
			state.RelayGroupID.ValueString() == cfg.RelayGroupID.ValueString() {
			plan.RelayGroupName = state.RelayGroupName
		}
	}

	return m.makePlanObj(ctx, plan, diags)
}

func (m dhcpSettingsModifier) planDhcpRange(ctx context.Context, state, cfg *tf.DhcpSettings, diags *diag.Diagnostics) types.Object {
	// Ensure IP Range is provided for DHCP_RANGE type
	if cfg.IPRange.IsNull() {
		diags.AddError("DHCP configuration error in dhcp_settings", "'ip_range' must be defined in the config ")
		return dhcpSettingNull
	}
	myState := state
	if myState == nil {
		myState = &tf.DhcpSettings{}
	}

	plan := tf.DhcpSettings{
		DhcpType:              cfg.DhcpType, // required
		IPRange:               cfg.IPRange,  // required for DHCP_RANGE
		DhcpMicrosegmentation: m.getOptionalBoolValue(cfg.DhcpMicrosegmentation, myState.DhcpMicrosegmentation, state != nil),
		RelayGroupID:          types.StringNull(),
		RelayGroupName:        types.StringNull(),
	}

	return m.makePlanObj(ctx, plan, diags)
}

func (m dhcpSettingsModifier) planDhcpEmpty(ctx context.Context, dhcpType string, diags *diag.Diagnostics) types.Object {
	plan := tf.DhcpSettings{
		DhcpType:              types.StringValue(dhcpType),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
		RelayGroupID:          types.StringNull(),
		RelayGroupName:        types.StringNull(),
	}

	return m.makePlanObj(ctx, plan, diags)
}

func (m dhcpSettingsModifier) makePlanObj(ctx context.Context, plan tf.DhcpSettings, diags *diag.Diagnostics) types.Object {
	planObj, objDiag := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, plan)
	if utils.CheckErr(diags, objDiag) {
		return types.ObjectNull(tf.DhcpSettingsAttrTypes)
	}
	return planObj
}

func (m dhcpSettingsModifier) getOptionalBoolValue(cfgVal, stateVal types.Bool, isState bool) types.Bool {
	// if it is in the config, use the config value
	if !cfgVal.IsNull() {
		return cfgVal
	}
	// if it is not in the config but is in state, use the state value
	if isState && !stateVal.IsNull() {
		return stateVal
	}
	// otherwise return unknown - not in the state yet, but API can return it
	return types.BoolUnknown()
}

func (m dhcpSettingsModifier) planDhcpDefault(ctx context.Context, req planmodifier.ObjectRequest, diags *diag.Diagnostics) types.Object {
	if !m.isRangeResource {
		// Native range: user did not configure dhcp_settings. Keep null so that
		// plan and post-apply state are consistent (parseNativeRange also returns
		// null in this case). Users who want explicit ACCOUNT_DEFAULT must set it.
		return dhcpSettingNull
	}

	// non-native range - default DHCP type depends on the range type
	var rangeType types.String
	if utils.CheckErr(diags, req.Config.GetAttribute(ctx, path.Root("range_type"), &rangeType)) {
		return dhcpSettingNull
	}
	if !utils.HasValue(rangeType) {
		diags.AddError("internal error", "failed to get range_type")
		return dhcpSettingNull
	}

	switch cato_models.SubnetType(rangeType.ValueString()) {
	case cato_models.SubnetTypeNative, cato_models.SubnetTypeVlan:
		return m.planDhcpEmpty(ctx, string(cato_models.DhcpTypeAccountDefault), diags)
	default:
		return m.planDhcpEmpty(ctx, string(cato_models.DhcpTypeDhcpDisabled), diags)
	}
}
