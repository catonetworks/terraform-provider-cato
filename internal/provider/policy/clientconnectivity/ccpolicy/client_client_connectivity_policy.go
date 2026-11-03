package ccpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyClientConnectivityPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityPolicy, error)
	PolicyClientConnectivityUpdatePolicy(
		ctx context.Context,
		input cato_models.ClientConnectivityPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityUpdatePolicy, error)
	PolicyClientConnectivityPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error)
}

func (r *clientConnectivityPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyClientConnectivityPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPolicy, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityUpdatePolicy(
	ctx context.Context,
	input cato_models.ClientConnectivityPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *clientConnectivityPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyClientConnectivityPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
