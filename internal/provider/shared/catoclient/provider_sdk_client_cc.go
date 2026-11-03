package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ClientConnectivity policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicyClientConnectivityPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityPolicy, error) {
	return c.Client.PolicyClientConnectivityPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityAddRule(
	ctx context.Context,
	input cato_models.ClientConnectivityAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityAddRule, error) {
	return c.Client.PolicyClientConnectivityAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityUpdateRule(
	ctx context.Context,
	input cato_models.ClientConnectivityUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityUpdateRule, error) {
	return c.Client.PolicyClientConnectivityUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityRemoveRule(
	ctx context.Context,
	input cato_models.ClientConnectivityRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityRemoveRule, error) {
	return c.Client.PolicyClientConnectivityRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityMoveRule, error) {
	return c.Client.PolicyClientConnectivityMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityAddSection, error) {
	return c.Client.PolicyClientConnectivityAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityUpdateSection, error) {
	return c.Client.PolicyClientConnectivityUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityRemoveSection, error) {
	return c.Client.PolicyClientConnectivityRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityMoveSection, error) {
	return c.Client.PolicyClientConnectivityMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityUpdatePolicy(
	ctx context.Context,
	input cato_models.ClientConnectivityPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityUpdatePolicy, error) {
	return c.Client.PolicyClientConnectivityUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyClientConnectivityPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityPublishPolicyRevision, error) {
	return c.Client.PolicyClientConnectivityPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
