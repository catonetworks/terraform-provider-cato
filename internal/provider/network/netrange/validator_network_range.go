package netrange

import (
	"context"
	"fmt"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/netrange"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/utils"
)

func GetNetworkRangeValidator() NetworkRangeValidator {
	return NetworkRangeValidator{}
}

// NetworkRangeValidator validates the NetworkRange settings
type NetworkRangeValidator struct{}

func (v NetworkRangeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var networkRange NetworkRange

	if !utils.HasValue(req.ConfigValue) {
		return
	}

	// get network range
	if apperr.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &networkRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if v.isUnknown(networkRange) {
		return
	}
	v.ValidateNetworkRange(ctx, &networkRange, &resp.Diagnostics)
}

func (v NetworkRangeValidator) ValidateNetworkRange(ctx context.Context, networkRange *NetworkRange, diags *diag.Diagnostics) {
	v.validateNetworkRange(ctx, networkRange, nil, diags)
}

// ValidateNetworkRangeWithPriorState validates ModifyPlan input. Terraform can populate Optional+Computed
// attributes from prior state before resource ModifyPlan runs, so only values that differ from
// prior state are treated as user-explicit for mutually-exclusive interface fields.
func (v NetworkRangeValidator) ValidateNetworkRangeWithPriorState(
	ctx context.Context, networkRange *NetworkRange, priorState *NetworkRange, diags *diag.Diagnostics,
) {
	v.validateNetworkRange(ctx, networkRange, priorState, diags)
}

func (v NetworkRangeValidator) validateNetworkRange(
	ctx context.Context, networkRange *NetworkRange, priorState *NetworkRange, diags *diag.Diagnostics,
) {
	// Validate that local_ip is within network_network_range
	if netrange.CheckLocalIP(diags, networkRange.LocalIP, networkRange.Subnet) != nil {
		return
	}

	// Validate DHCP settings
	if dhcp.DHCPChecker.CheckWithPriorState(ctx, diags, networkRange.DhcpSettings, priorStateDhcpSettings(priorState)) != nil {
		return
	}

	// Validate that interface_id and interface_index cannot be set simultaneously
	idExplicit := NetworkRangeInterfaceFieldIsExplicit(networkRange.InterfaceID, priorStateInterfaceID(priorState))
	indexExplicit := NetworkRangeInterfaceFieldIsExplicit(networkRange.InterfaceIndex, priorStateInterfaceIndex(priorState))
	if idExplicit && indexExplicit {
		diags.AddError("Invalid network range Configuration",
			fmt.Sprintf("interface_id '%s' and interface_index '%s' cannot be set simultaneously.",
				networkRange.InterfaceID.ValueString(), networkRange.InterfaceIndex.ValueString()))
		return
	}

	// Validate that InternetOnly and MdnsReflector cannot be set simultaneously
	if utils.HasValue(networkRange.InternetOnly) && utils.HasValue(networkRange.MdnsReflector) &&
		networkRange.InternetOnly.ValueBool() && networkRange.MdnsReflector.ValueBool() {
		diags.AddError(
			"Invalid network range Configuration",
			"InternetOnly and MdnsReflector cannot be set simultaneously",
		)
		return
	}

	// Validate attributes based on rangeType
	if v.checkRangeTypeAttributes(diags, networkRange) != nil {
		return
	}
}

func NetworkRangeInterfaceFieldIsExplicit(cfgVal, stateVal types.String) bool {
	if !utils.HasValue(cfgVal) {
		return false
	}
	return !utils.HasValue(stateVal) || cfgVal.ValueString() != stateVal.ValueString()
}

func priorStateInterfaceID(state *NetworkRange) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.InterfaceID
}

func priorStateInterfaceIndex(state *NetworkRange) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.InterfaceIndex
}

func priorStateDhcpSettings(state *NetworkRange) types.Object {
	if state == nil {
		return types.ObjectNull(dhcp.SettingsAttrTypes)
	}
	return state.DhcpSettings
}

func (v NetworkRangeValidator) checkRangeTypeAttributes(diags *diag.Diagnostics, networkRange *NetworkRange) error {
	rangeType := cato_models.SubnetType(networkRange.RangeType.ValueString())

	isAllowed := func(diags *diag.Diagnostics, rangeType cato_models.SubnetType, attributeName string,
		attributeValue utils.HasValuer, allowedTypes []cato_models.SubnetType) bool {
		if !utils.HasValue(attributeValue) {
			return true
		}
		if slices.Contains(allowedTypes, rangeType) {
			return true
		}
		diags.AddError("Invalid network range Configuration",
			fmt.Sprintf("'%s' is not allowed for rangeType '%s'", attributeName, rangeType))
		return false
	}

	if !isAllowed(diags, rangeType, "local_ip", networkRange.LocalIP, []cato_models.SubnetType{
		cato_models.SubnetTypeDirect,
		cato_models.SubnetTypeNative,
		cato_models.SubnetTypeSecondaryNative,
		cato_models.SubnetTypeVlan}) {
		return apperr.ErrConfig
	}

	if !isAllowed(diags, rangeType, "dhcp_settings", networkRange.DhcpSettings, []cato_models.SubnetType{
		cato_models.SubnetTypeNative,
		cato_models.SubnetTypeSecondaryNative,
		cato_models.SubnetTypeVlan}) {
		return apperr.ErrConfig
	}

	if !isAllowed(diags, rangeType, "gateway", networkRange.Gateway, []cato_models.SubnetType{
		cato_models.SubnetTypeRouted}) {
		return apperr.ErrConfig
	}

	if !isAllowed(diags, rangeType, "vlan", networkRange.Vlan, []cato_models.SubnetType{
		cato_models.SubnetTypeVlan}) {
		return apperr.ErrConfig
	}

	// Validate that mDNS reflector is not set to true when rangeType is "Routed"
	if rangeType == cato_models.SubnetTypeRouted &&
		utils.HasValue(networkRange.MdnsReflector) && networkRange.MdnsReflector.ValueBool() {
		diags.AddError("Invalid network range configuration", "mDNS cannot be enabled when rangeType is 'Routed'")
		return apperr.ErrConfig
	}

	// Validate that vlan is set for "VLAN" ranges
	if !utils.HasValue(networkRange.Vlan) && rangeType == cato_models.SubnetTypeVlan {
		diags.AddError("Invalid network range configuration", "vlan number must be set for VLAN range type")
		return apperr.ErrConfig
	}

	return nil
}

func (v NetworkRangeValidator) Description(_ context.Context) string {
	return "network range settings validation" // TODO: provide more detailed description
}

func (v NetworkRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v NetworkRangeValidator) isUnknown(networkRange NetworkRange) bool {
	return networkRange.DhcpSettings.IsUnknown() ||
		networkRange.Gateway.IsUnknown() ||
		networkRange.InterfaceID.IsUnknown() ||
		networkRange.InterfaceIndex.IsUnknown() ||
		networkRange.InternetOnly.IsUnknown() ||
		networkRange.MdnsReflector.IsUnknown() ||
		networkRange.LocalIP.IsUnknown() ||
		networkRange.Name.IsUnknown() ||
		networkRange.RangeType.IsUnknown() ||
		networkRange.SiteID.IsUnknown() ||
		networkRange.Subnet.IsUnknown() ||
		networkRange.TranslatedSubnet.IsUnknown() ||
		networkRange.Vlan.IsUnknown()
}
