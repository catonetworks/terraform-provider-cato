package networkrange

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// checkDhcpSettingsDefault checks DHCP settings config or state,
// If config is provided,
// return true if it is not defined or if dhcp_type is set to ACCOUNT_DEFAULT, false otherwise
// If config is not provided (nil), check the state with the same logic to determine if DHCP settings are default
//
// Reason: API does not return ACCOUNT_DEFAULT, but some other value based on CMA account configuration
func (r *networkRangeResource) checkDhcpSettingsDefault(ctx context.Context, cfg, state *tf.NetworkRange,
	diags *diag.Diagnostics,
) (isDhcpSettingsDefault bool) {
	// cfg defined -> Create/Update flow; check the config value
	if cfg != nil {
		if !utils.HasValue(cfg.DhcpSettings) {
			return false
		}
		var cfgDhcpSettings dhcp.Settings
		if utils.CheckErr(diags, cfg.DhcpSettings.As(ctx, &cfgDhcpSettings, basetypes.ObjectAsOptions{})) {
			return false
		}
		if utils.HasValue(cfgDhcpSettings.DhcpType) &&
			(cfgDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
			return true
		}
		return false
	}

	// cfg is nil -> called from Read(); check the state
	if state == nil || !utils.HasValue(state.DhcpSettings) {
		return false
	}

	var stateDhcpSettings dhcp.Settings
	if utils.CheckErr(diags, state.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{})) {
		return false
	}
	if utils.HasValue(stateDhcpSettings.DhcpType) &&
		(stateDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
		return true
	}
	return false
}

// convertHydrateNetworkRangeStateValues converts between the Terraform model and SDK values.
func convertHydrateNetworkRangeStateValues(
	state *tf.NetworkRange,
	networkRangeID string,
	responseRange *cato.NetworkRange_Site_NetworkRange,
	interfaceIndex string,
	interfaceID string,
	dhcpSettingsObj types.Object,
) (resultNewState tf.NetworkRange) {
	resultNewState = tf.NetworkRange{
		ID:               types.StringValue(networkRangeID),
		DhcpSettings:     dhcpSettingsObj,
		Gateway:          types.StringPointerValue(responseRange.Gateway),
		InterfaceID:      types.StringValue(interfaceID),
		InterfaceIndex:   types.StringValue(interfaceIndex),
		InternetOnly:     types.BoolValue(responseRange.InternetOnly),
		MdnsReflector:    types.BoolValue(responseRange.MdnsReflector),
		LocalIP:          types.StringPointerValue(responseRange.LocalIP), // TODO: HA
		Name:             types.StringValue(responseRange.Name),
		RangeType:        types.StringValue(string(responseRange.RangeType)),
		SiteID:           state.SiteID,
		Subnet:           types.StringValue(responseRange.Subnet),
		TranslatedSubnet: types.StringPointerValue(responseRange.TranslatedSubnet),
		Vlan:             types.Int64PointerValue(responseRange.Vlan),
	}

	if responseRange.RangeType != cato_models.SubnetTypeVlan {
		resultNewState.Vlan = types.Int64Null()
	}
	if state.MdnsReflector.IsNull() {
		resultNewState.MdnsReflector = types.BoolNull()
	}
	if !state.Gateway.IsUnknown() {
		resultNewState.Gateway = state.Gateway
	}
	return resultNewState
}

func addNetworkRangeInputFromTerraform(
	ctx context.Context,
	client dhcp.SdkClienter,
	cfg *tf.NetworkRange,
	plan *tf.NetworkRange,
	diags *diag.Diagnostics,
	rangeType cato_models.SubnetType,
) cato_models.AddNetworkRangeInput {
	return cato_models.AddNetworkRangeInput{
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, client, rangeType, plan.DhcpSettings, diags),
		Gateway:          convert.KnownStringPointer(plan.Gateway),
		InternetOnly:     convert.KnownBoolPointer(plan.InternetOnly),
		LocalIP:          convert.KnownStringPointer(plan.LocalIP),
		MdnsReflector:    convert.KnownBoolPointer(plan.MdnsReflector),
		Name:             plan.Name.ValueString(),
		RangeType:        rangeType,
		Subnet:           plan.Subnet.ValueString(),
		TranslatedSubnet: convert.TranslatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		Vlan:             convert.KnownInt64Pointer(plan.Vlan),
	}
}

func hydrateUpdateNetworkRangeAPI(
	ctx context.Context,
	client dhcp.SdkClienter,
	cfg, plan *tf.NetworkRange,
	diags *diag.Diagnostics,
) cato_models.UpdateNetworkRangeInput {
	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())

	return cato_models.UpdateNetworkRangeInput{
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, client, rangeType, plan.DhcpSettings, diags),
		Gateway:          convert.KnownStringPointer(plan.Gateway),
		InternetOnly:     convert.KnownBoolPointer(plan.InternetOnly),
		LocalIP:          convert.KnownStringPointer(plan.LocalIP),
		MdnsReflector:    convert.KnownBoolPointer(plan.MdnsReflector),
		Name:             convert.KnownStringPointer(plan.Name),
		RangeType:        (*cato_models.SubnetType)(convert.KnownStringPointer(plan.RangeType)),
		Subnet:           convert.KnownStringPointer(plan.Subnet),
		TranslatedSubnet: convert.TranslatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		Vlan:             convert.KnownInt64Pointer(plan.Vlan),
	}
}

func hydrateAddNetworkRangeAPI(
	ctx context.Context,
	client dhcp.SdkClienter,
	cfg, plan *tf.NetworkRange,
	diags *diag.Diagnostics,
) cato_models.AddNetworkRangeInput {
	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())

	return addNetworkRangeInputFromTerraform(ctx, client, cfg, plan, diags, rangeType)
}
