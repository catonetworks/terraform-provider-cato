package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// dhcpChecker validates that the provided DHCP settings are consistent and valid.
type dhcpChecker struct{}

// DHCPChecker validates that the provided DHCP settings are consistent and valid.
var DHCPChecker dhcpChecker

// Check validates the DHCP settings object.
// Returns error and updates diags if the settings are invalid, otherwise returns nil.
func (d dhcpChecker) Check(ctx context.Context, diags *diag.Diagnostics, dhcp types.Object) error {
	var dhcpSettings *tf.DhcpSettings

	if !utils.HasValue(dhcp) {
		return nil
	}

	// get dhcp config
	if utils.CheckErr(diags, dhcp.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})) {
		return ErrConfig
	}
	if d.isUnknown(dhcpSettings) {
		return nil
	}

	// validate dhcp-relay group setting
	if relayErr := d.checkDHCPRelay(diags, *dhcpSettings); relayErr != nil {
		return relayErr
	}

	// validate dhcp ip range setting
	if rangeErr := d.checkDHCPRange(diags, *dhcpSettings); rangeErr != nil {
		return rangeErr
	}

	return nil
}

// checkDHCPRelay validates the consistency of DHCP relay settings
// if DHCP type is DHCP_RELAY, exactly one of relay_group_name or relay_group_id must be set,
// otherwise, neither can be set.
func (d dhcpChecker) checkDHCPRelay(diags *diag.Diagnostics, dhcpSettings tf.DhcpSettings) error {
	relayGroupNameSet := utils.HasValue(dhcpSettings.RelayGroupName)
	relayGroupIDSet := utils.HasValue(dhcpSettings.RelayGroupID)
	dhcpType := cato_models.DhcpType(dhcpSettings.DhcpType.ValueString())

	// If DHCP type is not DHCP_RELAY, relay group name/id must not be configured
	if dhcpType != cato_models.DhcpTypeDhcpRelay {
		if relayGroupIDSet || relayGroupNameSet {
			diags.AddError("Invalid DHCP Configuration",
				fmt.Sprintf("relay_group_id or relay_group_name can only be configured when DHCP type is 'DHCP_RELAY' (have %q)", dhcpType))
			return ErrConfig
		}
		return nil
	}

	// If DHCP type is DHCP_RELAY, exactly on of relay_group_name relay_group_id must not be configured
	if !relayGroupNameSet && !relayGroupIDSet {
		diags.AddError("Invalid DHCP Configuration",
			"either relay_group_id or relay_group_name must be configured when DHCP type is 'DHCP_RELAY'")
		return ErrConfig
	}
	if relayGroupNameSet && relayGroupIDSet {
		diags.AddError("Invalid DHCP Configuration",
			"only one of relay_group_id or relay_group_name can be configured when DHCP type is 'DHCP_RELAY'")
		return ErrConfig
	}
	return nil
}

// checkDHCPRange validates the consistency of DHCP range settings
// - if DHCP type is DHCP_RANGE, ip_range must be set, otherwise it must not be set.
// - if DHCP type is not  DHCP_RANGE, dhcp_microsegmentation must not be set to true
func (d dhcpChecker) checkDHCPRange(diags *diag.Diagnostics, dhcpSettings tf.DhcpSettings) error {
	dhcpType := cato_models.DhcpType(dhcpSettings.DhcpType.ValueString())
	ipRangeSet := utils.HasValue(dhcpSettings.IPRange)

	if dhcpType == cato_models.DhcpTypeDhcpRange {
		if !ipRangeSet {
			diags.AddError("Invalid DHCP Configuration",
				"ip_range must be configured when DHCP type is 'DHCP_RANGE'")
			return ErrConfig
		}
		return nil
	}

	// dhcpType != DHCP_RANGE
	if ipRangeSet {
		diags.AddError("Invalid DHCP Configuration",
			"ip_range can only be configured when DHCP type is 'DHCP_RANGE'")
		return ErrConfig
	}

	if utils.HasValue(dhcpSettings.DhcpMicrosegmentation) && dhcpSettings.DhcpMicrosegmentation.ValueBool() {
		diags.AddError("Invalid DHCP Configuration",
			"dhcp_microsegmentation can only be seto to true, when DHCP type is 'DHCP_RANGE'")
		return ErrConfig
	}

	return nil
}

// isUnknown returns true if dhcpSettings or any of its components have unknown values.
func (d dhcpChecker) isUnknown(dhcpSettings *tf.DhcpSettings) bool {
	if dhcpSettings == nil {
		return true
	}
	return dhcpSettings.DhcpType.IsUnknown() ||
		dhcpSettings.IPRange.IsUnknown() ||
		dhcpSettings.RelayGroupID.IsUnknown() ||
		dhcpSettings.RelayGroupName.IsUnknown() ||
		dhcpSettings.DhcpMicrosegmentation.IsUnknown()
}
