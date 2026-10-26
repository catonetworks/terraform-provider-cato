package dhcprelaygroup

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type relayGroupAPI interface {
	GetID() string
	GetName() string
	GetServer() []string
}

func parseRelayGroup(ctx context.Context, plan *RelayGroupModel, item relayGroupAPI, diags *diag.Diagnostics) {
	plan.ID = types.StringValue(item.GetID())
	plan.Name = types.StringValue(item.GetName())
	plan.Server = nc.List(ctx, item.GetServer(), diags)
}

func buildCreateInput(ctx context.Context, plan RelayGroupModel, diags *diag.Diagnostics) models.NetworkConfigDhcpRelayGroupCreateInput {
	return models.NetworkConfigDhcpRelayGroupCreateInput{Name: plan.Name.ValueString(),
		Server: nc.Strings(ctx, plan.Server, diags)}
}

func buildUpdateInput(ctx context.Context, plan RelayGroupModel, diags *diag.Diagnostics) models.NetworkConfigDhcpRelayGroupUpdateInput {
	return models.NetworkConfigDhcpRelayGroupUpdateInput{ID: plan.ID.ValueString(),
		Name:   plan.Name.ValueStringPointer(),
		Server: nc.Strings(ctx, plan.Server, diags)}
}
