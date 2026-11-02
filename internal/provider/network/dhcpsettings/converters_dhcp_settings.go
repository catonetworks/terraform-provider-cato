package dhcpsettings

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildSettings(ctx context.Context, plan SettingsModel, diags *diag.Diagnostics) models.NetworkConfigDhcpUpdateSettingsInput {
	input := models.NetworkConfigDhcpUpdateSettingsInput{}
	if !plan.LeaseTime.IsUnknown() {
		input.LeaseTime = plan.LeaseTime.ValueInt64Pointer()
	}
	if plan.Relay.IsNull() || plan.Relay.IsUnknown() {
		return input
	}
	var relay RelayModel
	diags.Append(plan.Relay.As(ctx, &relay, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return input
	}
	input.Relay = &models.NetworkConfigDhcpRelayInput{Enabled: relay.Enabled.ValueBoolPointer()}
	if !relay.Timeout.IsUnknown() {
		input.Relay.Timeout = relay.Timeout.ValueInt64Pointer()
	}
	if !relay.Group.IsUnknown() {
		input.Relay.Group = idname.PrepareIDName[models.NetworkConfigDhcpRelayGroupRefInput](ctx, relay.Group, diags)
	}
	return input
}

func parseRelay(ctx context.Context, enabled bool, timeout *int64, id, name string, diags *diag.Diagnostics) types.Object {
	group := types.ObjectNull(idname.RefAttrTypes)
	if id != "" {
		var d diag.Diagnostics
		group, d = types.ObjectValueFrom(ctx, idname.RefAttrTypes, idname.Ref{ID: types.StringValue(id), Name: types.StringValue(name)})
		diags.Append(d...)
	}
	value,
		d := types.ObjectValueFrom(ctx,
		RelayAttrTypes,
		RelayModel{Enabled: types.BoolValue(enabled),
			Timeout: types.Int64PointerValue(timeout),
			Group:   group})
	diags.Append(d...)
	return value
}
