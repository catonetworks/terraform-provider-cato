package papolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyPrivateAccessUpdatePolicy(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessPolicyUpdateInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessUpdatePolicy, error)
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error)
}

func (r *privAccessPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyPrivateAccessUpdatePolicy(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessPolicyUpdateInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessUpdatePolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error) {
	return c.resource.client.Catov2.PolicyReadPrivateAccessPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *privAccessPolicyResource }

var _ ResourceClient = (*sdkClient)(nil)
