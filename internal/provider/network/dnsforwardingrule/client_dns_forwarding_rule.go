package dnsforwardingrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	NetworkConfigDNSCreateForwardingRule(ctx context.Context,
		accountID string,
		networkConfigDNSCreateForwardingRuleInput cato_models.NetworkConfigDNSCreateForwardingRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSCreateForwardingRule,
		error)
	NetworkConfigDNSDeleteForwardingRule(ctx context.Context,
		accountID string,
		networkConfigDNSDeleteForwardingRuleInput cato_models.NetworkConfigDNSDeleteForwardingRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSDeleteForwardingRule,
		error)
	NetworkConfigDNSForwardingRule(ctx context.Context,
		accountID string,
		networkConfigDNSForwardingRuleRefInput cato_models.NetworkConfigDNSForwardingRuleRefInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSForwardingRule,
		error)
	NetworkConfigDNSUpdateForwardingRule(ctx context.Context,
		accountID string,
		networkConfigDNSUpdateForwardingRuleInput cato_models.NetworkConfigDNSUpdateForwardingRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSUpdateForwardingRule,
		error)
}
type sdkClient struct{ *nc.DNSClient }

var _ ResourceClient = (*sdkClient)(nil)

func (r *forwardingRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{nc.NewDNSClient(r.client.Catov2)}
}
