package validators

import (
	"context"
	"fmt"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	// Schema-level validation: no prior state is available, so all non-null values are treated as
	// explicitly set by the user.
	v.ValidateNetworkRange(ctx, &networkRange, nil, &resp.Diagnostics)
}

// ValidateNetworkRange validates the NetworkRange config.
// priorState is the prior Terraform state used to detect state-propagated Optional+Computed
// attribute values. Terraform Core propagates prior-state values into req.Config for
// Optional+Computed attributes, so without a state-aware check the interface_id/index conflict
// validator would fire even when the user only set one of the two fields.
// Pass nil when no prior state is available (schema-level ValidateObject).
func (v NetworkRangeValidator) ValidateNetworkRange(
	ctx context.Context, networkRange *tf.NetworkRange, priorState *tf.NetworkRange, diags *diag.Diagnostics,
) {
	// Validate that local_ip is within network_network_range
	if checkLocalIP(diags, networkRange.LocalIP, networkRange.Subnet) != nil {
		return
	}

	// Validate DHCP settings
	if DHCPChecker.Check(ctx, diags, networkRange.DhcpSettings) != nil {
		return
	}

	// Validate that interface_id and interface_index cannot be set simultaneously.
	// When priorState is provided, only values that differ from state are treated as
	// user-explicitly-set (others are Terraform Core state propagation for Optional+Computed
	// attributes and must not trigger a conflict error).
	idExplicit := interfaceFieldIsExplicit(networkRange.InterfaceID, priorStateInterfaceID(priorState))
	indexExplicit := interfaceFieldIsExplicit(networkRange.InterfaceIndex, priorStateInterfaceIndex(priorState))
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

// interfaceFieldIsExplicit returns true when a cfg value is non-null AND differs from the
// corresponding prior-state value, indicating the user explicitly set it in their configuration.
// When priorState is nil (schema-level validation), any non-null value is treated as explicit.
func interfaceFieldIsExplicit(cfgVal, stateVal types.String) bool {
	if cfgVal.IsNull() {
		return false
	}
	// stateVal is null → no prior state → treat cfgVal as explicit
	return stateVal.IsNull() || cfgVal.ValueString() != stateVal.ValueString()
}

func priorStateInterfaceID(state *tf.NetworkRange) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.InterfaceID
}

func priorStateInterfaceIndex(state *tf.NetworkRange) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.InterfaceIndex
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

	if !isAllowed(diags, rangeType, "vlan", networkRange.Vlan, []cato_models.SubnetType{
		cato_models.SubnetTypeVlan}) {
		return ErrConfig
	}

	// Validate that mDNS reflector is not set to true when rangeType is "Routed"
	if rangeType == cato_models.SubnetTypeRouted &&
		utils.HasValue(networkRange.MdnsReflector) && networkRange.MdnsReflector.ValueBool() {
		diags.AddError("Invalid network range configuration", "mDNS cannot be enabled when rangeType is 'Routed'")
		return ErrConfig
	}

	// Validate that vlan is set for "VLAN" ranges
	if !utils.HasValue(networkRange.Vlan) && rangeType == cato_models.SubnetTypeVlan {
		diags.AddError("Invalid network range configuration", "vlan number must be set for VLAN range type")
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
