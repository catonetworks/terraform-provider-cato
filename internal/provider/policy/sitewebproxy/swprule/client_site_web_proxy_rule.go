package swprule

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
	PolicySiteWebProxyAddRule(
		ctx context.Context,
		input cato_models.SiteWebProxyAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyAddRule, error)
	PolicySiteWebProxyMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyMoveRule, error)
	PolicySiteWebProxyRemoveRule(
		ctx context.Context,
		input cato_models.SiteWebProxyRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyRemoveRule, error)
	PolicySiteWebProxyUpdateRule(
		ctx context.Context,
		input cato_models.SiteWebProxyUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyUpdateRule, error)
	PolicySiteWebProxyPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error)
}

func (r *siteWebProxyRuleResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) PolicySiteWebProxyAddRule(
	ctx context.Context,
	input cato_models.SiteWebProxyAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyAddRule, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyMoveRule, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyRemoveRule(
	ctx context.Context,
	input cato_models.SiteWebProxyRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyRemoveRule, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySiteWebProxyUpdateRule(
	ctx context.Context,
	input cato_models.SiteWebProxyUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyUpdateRule, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *siteWebProxyRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySiteWebProxyPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySiteWebProxyPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySiteWebProxyPublishPolicyRevision(ctx, accountID, interceptors...)
}
