package dnssettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type settingsAPI interface {
	GetPrimaryServer() *string
	GetSecondaryServer() *string
	GetSuffix() []string
	GetAcceptDNSRequestsOnLanInterfaceIP() bool
}

func buildSettings(ctx context.Context, plan SettingsModel, diags *diag.Diagnostics) nc.DNSSettingsInput {
	input := nc.BuildDNSInput(ctx, plan.PrimaryServer, plan.SecondaryServer, plan.Suffix, diags)
	input.AcceptDNSRequestsOnLanInterfaceIP = plan.AcceptDNSRequestsOnLanInterfaceIP.ValueBool()
	return input
}
func parseSettings(ctx context.Context, item settingsAPI, accountID string, diags *diag.Diagnostics) SettingsModel {
	state := SettingsModel{ID: types.StringValue(accountID),
		AcceptDNSRequestsOnLanInterfaceIP: types.BoolValue(item.GetAcceptDNSRequestsOnLanInterfaceIP())}
	nc.ParseDNS(ctx, item, &state.PrimaryServer, &state.SecondaryServer, &state.Suffix, diags)
	return state
}
