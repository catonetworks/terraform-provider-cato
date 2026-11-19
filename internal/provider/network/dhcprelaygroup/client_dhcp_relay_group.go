package dhcprelaygroup

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	NetworkConfigDhcpCreateRelayGroup(ctx context.Context,
		accountID string,
		networkConfigDhcpCreateRelayGroupInput cato_models.NetworkConfigDhcpCreateRelayGroupInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpCreateRelayGroup,
		error)
	NetworkConfigDhcpDeleteRelayGroup(ctx context.Context,
		accountID string,
		networkConfigDhcpDeleteRelayGroupInput cato_models.NetworkConfigDhcpDeleteRelayGroupInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpDeleteRelayGroup,
		error)
	NetworkConfigDhcpRelayGroup(ctx context.Context,
		accountID string,
		networkConfigDhcpRelayGroupRefInput cato_models.NetworkConfigDhcpRelayGroupRefInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpRelayGroup,
		error)
	NetworkConfigDhcpUpdateRelayGroup(ctx context.Context,
		accountID string,
		networkConfigDhcpUpdateRelayGroupInput cato_models.NetworkConfigDhcpUpdateRelayGroupInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDhcpUpdateRelayGroup,
		error)
}
type sdkClient struct{ *nc.DHCPClient }

var _ ResourceClient = (*sdkClient)(nil)

func (r *relayGroupResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{nc.NewDHCPClient(r.client.Catov2)}
}

// A downstream delete failure is not absence proof; confirm against the exact scoped object.
func (r *relayGroupResource) confirmDeleted(ctx context.Context, id string, operationErr error) bool {
	if !nc.IsBackendFailure(operationErr, "deleteRelayGroup") {
		return false
	}
	result,
		err := r.resourceAPI().NetworkConfigDhcpRelayGroup(ctx,
		r.client.AccountId,
		cato_models.NetworkConfigDhcpRelayGroupRefInput{By: cato_models.ObjectRefByID,
			Input: id})
	if err != nil {
		return nc.IsNotFound(err, "relayGroup")
	}
	return result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.RelayGroup == nil
}

func createdID(result *cato.NetworkConfigDhcpCreateRelayGroup) string {
	items := result.GetNetworkConfig().GetDhcp().GetCreateRelayGroup().GetRelayGroup()
	if len(items) != 1 {
		return ""
	}
	return items[0].GetID()
}
