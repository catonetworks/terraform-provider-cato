package networkconfig

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// DHCPClient adds DHCP response validation to the provider SDK.
type DHCPClient struct{ *catoclient.ProviderSDKClient }

// NewDHCPClient returns the DHCP-specific SDK adapter.
func NewDHCPClient(client *catoclient.ProviderSDKClient) *DHCPClient {
	return &DHCPClient{ProviderSDKClient: client}
}

func (c *DHCPClient) NetworkConfigDhcpSettings(ctx context.Context, accountID string,
	interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpSettings, error) {
	return c.ProviderSDKClient.NetworkConfigDhcpSettings(ctx, accountID,
		append(interceptors, validateReadResponse("dhcp", "settings"))...)
}

func (c *DHCPClient) NetworkConfigDhcpOption(ctx context.Context, accountID string, input cato_models.NetworkConfigDhcpOptionRefInput,
	interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpOption, error) {
	return c.ProviderSDKClient.NetworkConfigDhcpOption(ctx, accountID, input,
		append(interceptors, validateReadResponse("dhcp", "option"))...)
}

func (c *DHCPClient) NetworkConfigDhcpRelayGroup(ctx context.Context, accountID string,
	input cato_models.NetworkConfigDhcpRelayGroupRefInput,
	interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpRelayGroup, error) {
	return c.ProviderSDKClient.NetworkConfigDhcpRelayGroup(ctx, accountID, input,
		append(interceptors, validateReadResponse("dhcp", "relayGroup"))...)
}
