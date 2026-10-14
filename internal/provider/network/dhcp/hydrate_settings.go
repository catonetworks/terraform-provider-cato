package dhcp

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

func (r *settingsResource) write(ctx context.Context, plan *SettingsModel, diags *diag.Diagnostics) {
	// Unknown optional+computed fields mean preserve current server values, not clear.
	input := models.NetworkConfigDhcpUpdateSettingsInput{LeaseTime: plan.LeaseTime.ValueInt64Pointer(),
		Relay: &models.NetworkConfigDhcpRelayInput{Enabled: plan.RelayEnabled.ValueBoolPointer()}}
	if !plan.RelayTimeout.IsUnknown() && !plan.RelayTimeout.IsNull() {
		input.Relay.Timeout = plan.RelayTimeout.ValueInt64Pointer()
	}
	if !plan.RelayGroupID.IsUnknown() && !plan.RelayGroupID.IsNull() {
		input.Relay.Group = &models.NetworkConfigDhcpRelayGroupRefInput{By: models.ObjectRefByID, Input: plan.RelayGroupID.ValueString()}
	}
	if plan.RelayEnabled.ValueBool() && !r.completeRelay(ctx, input.Relay, diags) {
		return
	}
	result, err := r.Client.DHCP().NetworkConfigDhcpUpdateSettings(ctx, r.Client.AccountId, input)
	if err != nil {
		nc.APIError(diags, "updateSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.Dhcp.UpdateSettings != nil &&
		result.NetworkConfig.Dhcp.UpdateSettings.Settings.LeaseTime >= 60, diags) {
		return
	}
	item := result.NetworkConfig.Dhcp.UpdateSettings.Settings
	plan.ID = types.StringValue(r.Client.AccountId)
	plan.LeaseTime = types.Int64Value(item.LeaseTime)
	plan.RelayEnabled = types.BoolValue(item.Relay.Enabled)
	plan.RelayTimeout = types.Int64PointerValue(item.Relay.Timeout)
	plan.RelayGroupID = types.StringNull()
	if item.Relay.Group != nil {
		plan.RelayGroupID = types.StringValue(item.Relay.Group.ID)
	}
}

func (r *settingsResource) completeRelay(ctx context.Context, input *models.NetworkConfigDhcpRelayInput, diags *diag.Diagnostics) bool {
	if input.Timeout != nil && input.Group != nil {
		return true
	}

	current, err := r.Client.DHCP().NetworkConfigDhcpSettings(ctx, r.Client.AccountId)
	if err != nil {
		nc.APIError(diags, "read relay configuration", err)
		return false
	}
	if !nc.RequirePayload(current != nil && current.NetworkConfig != nil && current.NetworkConfig.Dhcp.Settings != nil, diags) {
		return false
	}
	relay := current.NetworkConfig.Dhcp.Settings.Relay
	if input.Timeout == nil {
		input.Timeout = relay.Timeout
	}
	if input.Group == nil && relay.Group != nil {
		input.Group = &models.NetworkConfigDhcpRelayGroupRefInput{By: models.ObjectRefByID, Input: relay.Group.ID}
	}
	if input.Timeout == nil || input.Group == nil {
		diags.AddError("Invalid DHCP relay configuration", "Enabled relay requires a timeout and relay group.")
		return false
	}

	return true
}
