package dhcp

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperr"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// dhcpChecker validates that the provided DHCP settings are consistent and valid.
type dhcpChecker struct{}

// DHCPChecker validates that the provided DHCP settings are consistent and valid.
var DHCPChecker dhcpChecker

// Check validates the DHCP settings object.
// Returns error and updates diags if the settings are invalid, otherwise returns nil.
func (d dhcpChecker) Check(ctx context.Context, diags *diag.Diagnostics, dhcp types.Object) error {
	return d.check(ctx, diags, dhcp, types.ObjectNull(tf.DhcpSettingsAttrTypes))
}

// CheckWithPriorState validates DHCP settings during ModifyPlan. Terraform can populate
// Optional+Computed nested attributes from prior state into req.Config, so prior state is used
// to distinguish user-explicit relay fields from propagated values.
func (d dhcpChecker) CheckWithPriorState(ctx context.Context, diags *diag.Diagnostics, dhcp, priorState types.Object) error {
	return d.check(ctx, diags, dhcp, priorState)
}

func (d dhcpChecker) check(ctx context.Context, diags *diag.Diagnostics, dhcp, priorState types.Object) error {
	var dhcpSettings *Settings

	if !utils.HasValue(dhcp) {
		return nil
	}

	// get dhcp config
	if utils.CheckErr(diags, dhcp.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})) {
		return apperr.ErrConfig
	}
	if d.isUnknown(dhcpSettings) {
		return nil
	}

	var priorDhcpSettings *Settings
	if utils.HasValue(priorState) {
		if utils.CheckErr(diags, priorState.As(ctx, &priorDhcpSettings, basetypes.ObjectAsOptions{})) {
			return apperr.ErrConfig
		}
		if d.isUnknown(priorDhcpSettings) {
			priorDhcpSettings = nil
		}
	}

	// validate dhcp-relay group setting
	if relayErr := d.checkDHCPRelay(diags, *dhcpSettings, priorDhcpSettings); relayErr != nil {
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
func (d dhcpChecker) checkDHCPRelay(diags *diag.Diagnostics, dhcpSettings Settings, priorState *Settings) error {
	relayGroupNameSet := utils.HasValue(dhcpSettings.RelayGroupName)
	relayGroupIDSet := utils.HasValue(dhcpSettings.RelayGroupID)
	relayGroupNameExplicit := utils.StringFieldIsExplicit(dhcpSettings.RelayGroupName, priorStateRelayGroupName(priorState))
	relayGroupIDExplicit := utils.StringFieldIsExplicit(dhcpSettings.RelayGroupID, priorStateRelayGroupID(priorState))
	dhcpType := cato_models.DhcpType(dhcpSettings.DhcpType.ValueString())

	// If DHCP type is not DHCP_RELAY, relay group name/id must not be configured
	if dhcpType != cato_models.DhcpTypeDhcpRelay {
		if relayGroupIDExplicit || relayGroupNameExplicit {
			diags.AddError("Invalid DHCP Configuration",
				fmt.Sprintf("relay_group_id or relay_group_name can only be configured when DHCP type is 'DHCP_RELAY' (have %q)", dhcpType))
			return apperr.ErrConfig
		}
		return nil
	}

	// If DHCP type is DHCP_RELAY, exactly on of relay_group_name relay_group_id must not be configured
	if !relayGroupNameSet && !relayGroupIDSet {
		diags.AddError("Invalid DHCP Configuration",
			"either relay_group_id or relay_group_name must be configured when DHCP type is 'DHCP_RELAY'")
		return apperr.ErrConfig
	}
	if relayGroupNameExplicit && relayGroupIDExplicit {
		diags.AddError("Invalid DHCP Configuration",
			"only one of relay_group_id or relay_group_name can be configured when DHCP type is 'DHCP_RELAY'")
		return apperr.ErrConfig
	}
	return nil
}

func priorStateRelayGroupName(state *Settings) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.RelayGroupName
}

func priorStateRelayGroupID(state *Settings) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.RelayGroupID
}

// checkDHCPRange validates the consistency of DHCP range settings
// - if DHCP type is DHCP_RANGE, ip_range must be set, otherwise it must not be set.
// - if DHCP type is not  DHCP_RANGE, dhcp_microsegmentation must not be set to true
func (d dhcpChecker) checkDHCPRange(diags *diag.Diagnostics, dhcpSettings Settings) error {
	dhcpType := cato_models.DhcpType(dhcpSettings.DhcpType.ValueString())
	ipRangeSet := utils.HasValue(dhcpSettings.IPRange)

	if dhcpType == cato_models.DhcpTypeDhcpRange {
		if !ipRangeSet {
			diags.AddError("Invalid DHCP Configuration",
				"ip_range must be configured when DHCP type is 'DHCP_RANGE'")
			return apperr.ErrConfig
		}
		return nil
	}

	// dhcpType != DHCP_RANGE
	if ipRangeSet {
		diags.AddError("Invalid DHCP Configuration",
			"ip_range can only be configured when DHCP type is 'DHCP_RANGE'")
		return apperr.ErrConfig
	}

	if utils.HasValue(dhcpSettings.DhcpMicrosegmentation) && dhcpSettings.DhcpMicrosegmentation.ValueBool() {
		diags.AddError("Invalid DHCP Configuration",
			"dhcp_microsegmentation can only be seto to true, when DHCP type is 'DHCP_RANGE'")
		return apperr.ErrConfig
	}

	return nil
}

// isUnknown returns true if dhcpSettings or any of its components have unknown values.
func (d dhcpChecker) isUnknown(dhcpSettings *Settings) bool {
	if dhcpSettings == nil {
		return true
	}
	return dhcpSettings.DhcpType.IsUnknown() ||
		dhcpSettings.IPRange.IsUnknown() ||
		dhcpSettings.RelayGroupID.IsUnknown() ||
		dhcpSettings.RelayGroupName.IsUnknown() ||
		dhcpSettings.DhcpMicrosegmentation.IsUnknown()
}

// TypeValidator validates that the provided string is a valid DHCP type
type (
	TypeValidator struct{}
)

func (v TypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	dhcpType := cato_models.DhcpType(value)
	if !dhcpType.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid DHCP type (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllDhcpType))
		return
	}
}

func (v TypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("DHCP type must be one of: %v", cato_models.AllDhcpType)
}
func (v TypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
