package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type DHCPClient struct{ *ProviderSDKClient }

func NewDHCPClient(client *ProviderSDKClient) *DHCPClient {
	return &DHCPClient{ProviderSDKClient: client}
}

func (s *Service) DHCP() *DHCPClient { return NewDHCPClient(s.Catov2) }

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
