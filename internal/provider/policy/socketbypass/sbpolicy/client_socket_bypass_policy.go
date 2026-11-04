package sbpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketBypassPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassPolicy, error)
	PolicySocketBypassUpdatePolicy(
		ctx context.Context,
		input cato_models.SocketBypassPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassUpdatePolicy, error)
	PolicySocketBypassPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error)
}

func (r *socketBypassPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketBypassPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPolicy, error) {
	return c.resource.client.Catov2.PolicySocketBypassPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicySocketBypassUpdatePolicy(
	ctx context.Context,
	input cato_models.SocketBypassPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicySocketBypassUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *socketBypassPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySocketBypassPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySocketBypassPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
