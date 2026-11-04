package atpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AppTenantRestrictionPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppTenantRestrictionPolicy, error)
	PolicyAppTenantRestrictionUpdatePolicy(
		ctx context.Context,
		input cato_models.AppTenantRestrictionPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionUpdatePolicy, error)
	PolicyAppTenantRestrictionPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error)
}

func (r *appTenantRestrictionPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) AppTenantRestrictionPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppTenantRestrictionPolicy, error) {
	return c.resource.client.Catov2.AppTenantRestrictionPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionUpdatePolicy(
	ctx context.Context,
	input cato_models.AppTenantRestrictionPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionUpdatePolicy(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *appTenantRestrictionPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyAppTenantRestrictionPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, accountID, interceptors...)
}
