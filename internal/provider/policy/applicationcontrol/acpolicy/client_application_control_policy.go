package acpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	ApplicationControlPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ApplicationControlPolicy, error)
	PolicyApplicationControlUpdatePolicy(
		ctx context.Context,
		input cato_models.ApplicationControlPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlUpdatePolicy, error)
	PolicyApplicationControlPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error)
}

func (r *applicationControlPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) ApplicationControlPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ApplicationControlPolicy, error) {
	return c.resource.client.Catov2.ApplicationControlPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlUpdatePolicy(
	ctx context.Context,
	input cato_models.ApplicationControlPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyApplicationControlUpdatePolicy(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *applicationControlPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyApplicationControlPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyApplicationControlPublishPolicyRevision(ctx, accountID, interceptors...)
}
