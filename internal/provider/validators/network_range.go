package validators

import (
	"context"
	"fmt"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

func GetNetworkRangeValidator() NetworkRangeValidator {
	return NetworkRangeValidator{}
}

// NetworkRangeValidator validates the NetworkRange settings
type NetworkRangeValidator struct{}

func (v NetworkRangeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var networkRange tf.NetworkRange

	if !utils.HasValue(req.ConfigValue) {
		return
	}

	// get network range
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &networkRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if v.isUnknown(networkRange) {
		return
	}
	v.ValidateNetworkRange(ctx, &networkRange, &resp.Diagnostics)
}

func (v NetworkRangeValidator) ValidateNetworkRange(ctx context.Context, networkRange *tf.NetworkRange, diags *diag.Diagnostics) {
	// Validate that local_ip is within network_network_range
	if checkLocalIP(diags, networkRange.LocalIP, networkRange.Subnet) != nil {
		return
	}

	// Validate DHCP settings
	if DHCPChecker.Check(ctx, diags, networkRange.DhcpSettings) != nil {
		return
	}

	// Validate that interface_id and interface_index cannot be set simultaneously
	if utils.HasValue(networkRange.InterfaceID) && utils.HasValue(networkRange.InterfaceIndex) {
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

func (v NetworkRangeValidator) checkRangeTypeAttributes(diags *diag.Diagnostics, networkRange *tf.NetworkRange) error {
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
		return ErrConfig
	}

	if !isAllowed(diags, rangeType, "dhcp_settings", networkRange.DhcpSettings, []cato_models.SubnetType{
		cato_models.SubnetTypeNative,
		cato_models.SubnetTypeSecondaryNative,
		cato_models.SubnetTypeVlan}) {
		return ErrConfig
	}

	if !isAllowed(diags, rangeType, "gateway", networkRange.Gateway, []cato_models.SubnetType{
		cato_models.SubnetTypeRouted}) {
		return ErrConfig
	}

	// Validate that mDNS reflector is not set to true when rangeType is "Routed"
	if utils.HasValue(networkRange.MdnsReflector) && networkRange.MdnsReflector.ValueBool() {
		diags.AddError("Invalid network range configuration", "mDNS cannot be enabled when rangeType is 'Routed'")
		return ErrConfig
	}

	// gateway is required for Routed range
	if !utils.HasValue(networkRange.Gateway) && rangeType == cato_models.SubnetTypeRouted {
		diags.AddError("Invalid network range configuration", "gateway is required for rangeType 'Routed'")
		return ErrConfig
	}

	return nil
}

func (v NetworkRangeValidator) Description(_ context.Context) string {
	return "network range settings validation" // TODO: provide more detailed description
}

func (v NetworkRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v NetworkRangeValidator) isUnknown(networkRange tf.NetworkRange) bool {
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
