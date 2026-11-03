package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// SiteWebProxy policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicySiteWebProxyPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyPolicy, error) {
	return c.Client.PolicySiteWebProxyPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyAddRule(
	ctx context.Context,
	input cato_models.SiteWebProxyAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyAddRule, error) {
	return c.Client.PolicySiteWebProxyAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyUpdateRule(
	ctx context.Context,
	input cato_models.SiteWebProxyUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyUpdateRule, error) {
	return c.Client.PolicySiteWebProxyUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyRemoveRule(
	ctx context.Context,
	input cato_models.SiteWebProxyRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyRemoveRule, error) {
	return c.Client.PolicySiteWebProxyRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyMoveRule, error) {
	return c.Client.PolicySiteWebProxyMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyAddSection, error) {
	return c.Client.PolicySiteWebProxyAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyUpdateSection, error) {
	return c.Client.PolicySiteWebProxyUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyRemoveSection, error) {
	return c.Client.PolicySiteWebProxyRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyMoveSection, error) {
	return c.Client.PolicySiteWebProxyMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyUpdatePolicy(
	ctx context.Context,
	input cato_models.SiteWebProxyPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyUpdatePolicy, error) {
	return c.Client.PolicySiteWebProxyUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySiteWebProxyPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyPublishPolicyRevision, error) {
	return c.Client.PolicySiteWebProxyPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
