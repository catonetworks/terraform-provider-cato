package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// TerminalServer policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicyTerminalServerPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerPolicy, error) {
	return c.Client.PolicyTerminalServerPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerAddRule(
	ctx context.Context,
	input cato_models.TerminalServerAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerAddRule, error) {
	return c.Client.PolicyTerminalServerAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerUpdateRule(
	ctx context.Context,
	input cato_models.TerminalServerUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerUpdateRule, error) {
	return c.Client.PolicyTerminalServerUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerRemoveRule(
	ctx context.Context,
	input cato_models.TerminalServerRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerRemoveRule, error) {
	return c.Client.PolicyTerminalServerRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerMoveRule, error) {
	return c.Client.PolicyTerminalServerMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerAddSection, error) {
	return c.Client.PolicyTerminalServerAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerUpdateSection, error) {
	return c.Client.PolicyTerminalServerUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerRemoveSection, error) {
	return c.Client.PolicyTerminalServerRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerMoveSection, error) {
	return c.Client.PolicyTerminalServerMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerUpdatePolicy(
	ctx context.Context,
	input cato_models.TerminalServerPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerUpdatePolicy, error) {
	return c.Client.PolicyTerminalServerUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyTerminalServerPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerPublishPolicyRevision, error) {
	return c.Client.PolicyTerminalServerPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
