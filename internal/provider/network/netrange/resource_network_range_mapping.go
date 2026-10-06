package netrange

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/netrange/application"
)

// planInterfaceIDIndex keeps the unconfigured interface field consistent with prior state.
func (r *networkRangeResource) planInterfaceIDIndex(cfg, plan, state *NetworkRange, stateDefined bool) {
	if !stateDefined {
		return
	}

	indexExplicit := NetworkRangeInterfaceFieldIsExplicit(cfg.InterfaceIndex, state.InterfaceIndex)
	idExplicit := NetworkRangeInterfaceFieldIsExplicit(cfg.InterfaceID, state.InterfaceID)

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

// defaultPlanValue returns the appropriate plan value: cfg -> state -> unknown

func defaultPlanValue(cfgValue, stateValue types.String, stateDefined bool) (planValue types.String) {
	newValue := types.StringUnknown()
	if utils.HasValue(cfgValue) {
		newValue = cfgValue
	} else if stateDefined && utils.HasValue(stateValue) {
		newValue = stateValue
	}
	return newValue
}

func (r *networkRangeResource) checkDhcpSettingsDefault(ctx context.Context, cfg, state *NetworkRange,
	diags *diag.Diagnostics,
) (isDhcpSettingsDefault bool) {
	// cfg defined -> Create/Update flow; check the config value
	if cfg != nil {
		if !utils.HasValue(cfg.DhcpSettings) {
			return false
		}
		var cfgDhcpSettings dhcp.Settings
		if apperr.CheckErr(diags, cfg.DhcpSettings.As(ctx, &cfgDhcpSettings, basetypes.ObjectAsOptions{})) {
			return false
		}
		if utils.HasValue(cfgDhcpSettings.DhcpType) &&
			(cfgDhcpSettings.DhcpType.ValueString() == "ACCOUNT_DEFAULT") {
			return true
		}
		return false
	}

	// cfg is nil -> called from Read(); check the state
	if state == nil || !utils.HasValue(state.DhcpSettings) {
		return false
	}

	var stateDhcpSettings dhcp.Settings
	if apperr.CheckErr(diags, state.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{})) {
		return false
	}
	if utils.HasValue(stateDhcpSettings.DhcpType) &&
		(stateDhcpSettings.DhcpType.ValueString() == "ACCOUNT_DEFAULT") {
		return true
	}
	return false
}

func translatedSubnetForAPIInput(configValue, planValue types.String) *string {
	if configValue.IsNull() || configValue.IsUnknown() {
		return nil
	}
	return parse.StringPointerForOptionalInput(planValue)
}
func (r *networkRangeResource) projectState(
	ctx context.Context,
	cfg, state *NetworkRange,
	responseRange *application.Snapshot,
	diags *diag.Diagnostics,
) NetworkRange {
	// DHCP settings
	isDhcpSettingsDefault := r.checkDhcpSettingsDefault(ctx, cfg, state, diags)
	dhcpSettingsObj := dhcp.SettingsDefault(ctx, diags)
	if responseRange.DhcpSettings != nil && !isDhcpSettingsDefault {
		dhcpSettingsObj = dhcp.ProjectSettings(ctx, responseRange.DhcpSettings, diags)
	}
	if diags.HasError() {
		return NetworkRange{}
	}

	newState := NetworkRange{

		ID:             types.StringValue(responseRange.NetworkRangeID),
		DhcpSettings:   dhcpSettingsObj,
		Gateway:        types.StringPointerValue(responseRange.Gateway),
		InterfaceID:    types.StringValue(responseRange.InterfaceID),
		InterfaceIndex: types.StringValue(responseRange.InterfaceIndex),
		InternetOnly:   types.BoolValue(responseRange.InternetOnly),
		MdnsReflector:  types.BoolValue(responseRange.MdnsReflector),
		LocalIP:        types.StringPointerValue(responseRange.LocalIP),
		// TODO: HA
		Name:             types.StringValue(responseRange.Name),
		RangeType:        types.StringValue(responseRange.RangeType),
		SiteID:           state.SiteID,
		Subnet:           types.StringValue(responseRange.Subnet),
		TranslatedSubnet: types.StringPointerValue(responseRange.TranslatedSubnet),
		Vlan:             types.Int64PointerValue(responseRange.Vlan),
	}

	if responseRange.RangeType != "VLAN" {
		newState.Vlan = types.Int64Null()
	}
	if state.MdnsReflector.IsNull() {
		newState.MdnsReflector = types.BoolNull()
	}
	if !state.Gateway.IsUnknown() {
		newState.Gateway = state.Gateway
	}

	return newState
}

func (r *networkRangeResource) readInput(
	ctx context.Context,
	cfg, state *NetworkRange,
	id string,
	diags *diag.Diagnostics,
) application.ReadInput {
	in := application.ReadInput{

		ID:               id,
		InterfaceID:      state.InterfaceID.ValueString(),
		InterfaceIndex:   state.InterfaceIndex.ValueString(),
		ResolveRelayName: !r.checkDhcpSettingsDefault(ctx, cfg, state, diags),
	}
	if cfg != nil {
		in.SiteID = cfg.SiteID.ValueString()
		in.ResolveInterface = true
		in.InterfaceByID = utils.HasValue(cfg.InterfaceID)
	}
	return in
}
func (r *networkRangeResource) prepareInput(ctx context.Context, cfg, plan *NetworkRange, diags *diag.Diagnostics) application.Input {
	return application.Input{

		ID:               plan.ID.ValueString(),
		SiteID:           plan.SiteID.ValueString(),
		InterfaceID:      parse.KnownStringPointer(plan.InterfaceID),
		InterfaceIndex:   plan.InterfaceIndex.ValueString(),
		Name:             parse.KnownStringPointer(plan.Name),
		RangeType:        parse.KnownStringPointer(plan.RangeType),
		Subnet:           parse.KnownStringPointer(plan.Subnet),
		Gateway:          parse.KnownStringPointer(plan.Gateway),
		LocalIP:          parse.KnownStringPointer(plan.LocalIP),
		TranslatedSubnet: translatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		InternetOnly:     parse.KnownBoolPointer(plan.InternetOnly),
		MdnsReflector:    parse.KnownBoolPointer(plan.MdnsReflector),
		Vlan:             parse.KnownInt64Pointer(plan.Vlan),
		DhcpSettings:     dhcp.DecodeSettings(ctx, plan.DhcpSettings, diags),
	}
}
