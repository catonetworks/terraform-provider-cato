package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// SocketBypass policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicySocketBypassPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassPolicy, error) {
	return c.Client.PolicySocketBypassPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassAddRule(
	ctx context.Context,
	input cato_models.SocketBypassAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassAddRule, error) {
	return c.Client.PolicySocketBypassAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassUpdateRule(
	ctx context.Context,
	input cato_models.SocketBypassUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassUpdateRule, error) {
	return c.Client.PolicySocketBypassUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassRemoveRule(
	ctx context.Context,
	input cato_models.SocketBypassRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassRemoveRule, error) {
	return c.Client.PolicySocketBypassRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassMoveRule, error) {
	return c.Client.PolicySocketBypassMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassAddSection, error) {
	return c.Client.PolicySocketBypassAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassUpdateSection, error) {
	return c.Client.PolicySocketBypassUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassRemoveSection, error) {
	return c.Client.PolicySocketBypassRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassMoveSection, error) {
	return c.Client.PolicySocketBypassMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassUpdatePolicy(
	ctx context.Context,
	input cato_models.SocketBypassPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassUpdatePolicy, error) {
	return c.Client.PolicySocketBypassUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySocketBypassPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassPublishPolicyRevision, error) {
	return c.Client.PolicySocketBypassPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
