package ifpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	InternetFirewallPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.InternetFirewallPolicy, error)
	PolicyInternetFirewallUpdatePolicy(
		ctx context.Context,
		input cato_models.InternetFirewallPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdatePolicy, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
}

func (r *internetFirewallPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct {
	resource *internetFirewallPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) InternetFirewallPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.InternetFirewallPolicy, error) {
	return c.resource.client.Catov2.InternetFirewallPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallUpdatePolicy(
	ctx context.Context,
	input cato_models.InternetFirewallPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallUpdatePolicy(ctx, nil, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, nil, nil, accountID, interceptors...)
}
