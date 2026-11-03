package diapolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyDynamicIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error)
	PolicyDynamicIPAllocationUpdatePolicy(
		ctx context.Context,
		input cato_models.DynamicIPAllocationPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationUpdatePolicy, error)
	PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error)
}

func (r *dynamicIPAllocationPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyDynamicIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationUpdatePolicy(
	ctx context.Context,
	input cato_models.DynamicIPAllocationPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationUpdatePolicy(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *dynamicIPAllocationPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyDynamicIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPublishPolicyRevision(ctx, accountID, interceptors...)
}
