package swppolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySiteWebProxyPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyPolicy, error)
	PolicySiteWebProxyUpdatePolicy(
		ctx context.Context,
		input cato_models.SiteWebProxyPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyUpdatePolicy, error)
	PolicySiteWebProxyPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error)
}

func (r *siteWebProxyPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySiteWebProxyPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyPolicy, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyUpdatePolicy(
	ctx context.Context,
	input cato_models.SiteWebProxyPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyUpdatePolicy(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *siteWebProxyPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySiteWebProxyPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyPublishPolicyRevision(ctx, accountID, interceptors...)
}
