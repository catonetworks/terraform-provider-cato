package wnwpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	WanNetworkPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanNetworkPolicy, error)
	PolicyWanNetworkUpdatePolicy(
		ctx context.Context,
		input cato_models.WanNetworkPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkUpdatePolicy, error)
	PolicyWanNetworkPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error)
}

func (r *wanNetworkPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct {
	resource *wanNetworkPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) WanNetworkPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanNetworkPolicy, error) {
	return c.resource.client.Catov2.WanNetworkPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkUpdatePolicy(
	ctx context.Context,
	input cato_models.WanNetworkPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyWanNetworkUpdatePolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, accountID, interceptors...)
}
