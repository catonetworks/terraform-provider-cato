package rpfpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	RemotePortFwdPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.RemotePortFwdPolicy, error)
	PolicyRemotePortFwdUpdatePolicy(
		ctx context.Context,
		input cato_models.RemotePortFwdPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdUpdatePolicy, error)
	PolicyRemotePortFwdPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error)
}

func (r *remotePortFwdPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) RemotePortFwdPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.RemotePortFwdPolicy, error) {
	return c.resource.client.Catov2.RemotePortFwdPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdUpdatePolicy(
	ctx context.Context,
	input cato_models.RemotePortFwdPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *remotePortFwdPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyRemotePortFwdPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
