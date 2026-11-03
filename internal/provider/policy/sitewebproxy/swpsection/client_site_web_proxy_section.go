package swpsection

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
	PolicySiteWebProxyAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyAddSection, error)
	PolicySiteWebProxyMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyMoveSection, error)
	PolicySiteWebProxyRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyRemoveSection, error)
	PolicySiteWebProxyUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyUpdateSection, error)
	PolicySiteWebProxyPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error)
}

func (r *siteWebProxySectionResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) PolicySiteWebProxyAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyAddSection, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyMoveSection, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyRemoveSection, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyUpdateSection, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *siteWebProxySectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySiteWebProxyPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyPublishPolicyRevision(ctx, accountID, interceptors...)
}
