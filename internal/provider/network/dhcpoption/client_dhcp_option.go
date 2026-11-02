package dhcpoption

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	NetworkConfigDhcpCreateOption(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkConfigDhcpCreateOptionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDhcpCreateOption, error)
	NetworkConfigDhcpDeleteOption(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkConfigDhcpDeleteOptionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDhcpDeleteOption, error)
	NetworkConfigDhcpOption(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkConfigDhcpOptionRefInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDhcpOption, error)
	NetworkConfigDhcpUpdateOption(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkConfigDhcpUpdateOptionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDhcpUpdateOption, error)
}
type sdkClient struct{ *nc.DHCPClient }

var _ ResourceClient = (*sdkClient)(nil)

func (r *optionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{nc.NewDHCPClient(r.client.Catov2)}
}
