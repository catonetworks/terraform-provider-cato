package swpindex

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
	PolicySiteWebProxyMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyMoveSection, error)
	PolicySiteWebProxyMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyMoveRule, error)
	PolicySiteWebProxyPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error)
}

func (r *swpRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *swpRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySiteWebProxyPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyPolicy, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyMoveSection, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyMoveRule, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyPublishPolicyRevision(ctx, accountID, interceptors...)
}
