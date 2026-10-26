package dhcpsettings

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	NetworkConfigDhcpSettings(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDhcpSettings, error)
	NetworkConfigDhcpUpdateSettings(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkConfigDhcpUpdateSettingsInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDhcpUpdateSettings, error)
}
type sdkClient struct{ *nc.DHCPClient }

var _ ResourceClient = (*sdkClient)(nil)

func (r *settingsResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{nc.NewDHCPClient(r.client.Catov2)}
}

func completeRelay(ctx context.Context,
	client ResourceClient,
	account string,
	input *cato_models.NetworkConfigDhcpRelayInput,
	diags *diag.Diagnostics) bool {
	if input.Timeout != nil && input.Group != nil {
		return true
	}

	current, err := client.NetworkConfigDhcpSettings(ctx, account)
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
		input.Group = &cato_models.NetworkConfigDhcpRelayGroupRefInput{By: cato_models.ObjectRefByID, Input: relay.Group.ID}
	}
	if input.Timeout == nil || input.Group == nil {
		diags.AddError("Invalid DHCP relay configuration", "Enabled relay requires a timeout and relay group.")
		return false
	}

	return true
}
