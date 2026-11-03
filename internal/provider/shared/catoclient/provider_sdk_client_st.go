package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// SplitTunnel policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicySplitTunnelPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelPolicy, error) {
	return c.Client.PolicySplitTunnelPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelAddRule(
	ctx context.Context,
	input cato_models.SplitTunnelAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelAddRule, error) {
	return c.Client.PolicySplitTunnelAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelUpdateRule(
	ctx context.Context,
	input cato_models.SplitTunnelUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelUpdateRule, error) {
	return c.Client.PolicySplitTunnelUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelRemoveRule(
	ctx context.Context,
	input cato_models.SplitTunnelRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelRemoveRule, error) {
	return c.Client.PolicySplitTunnelRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelMoveRule, error) {
	return c.Client.PolicySplitTunnelMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelAddSection, error) {
	return c.Client.PolicySplitTunnelAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelUpdateSection, error) {
	return c.Client.PolicySplitTunnelUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelRemoveSection, error) {
	return c.Client.PolicySplitTunnelRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelMoveSection, error) {
	return c.Client.PolicySplitTunnelMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelUpdatePolicy(
	ctx context.Context,
	input cato_models.SplitTunnelPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelUpdatePolicy, error) {
	return c.Client.PolicySplitTunnelUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySplitTunnelPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelPublishPolicyRevision, error) {
	return c.Client.PolicySplitTunnelPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
