package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// RemotePortFwd policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) RemotePortFwdPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.RemotePortFwdPolicy, error) {
	return c.Client.RemotePortFwdPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdAddRule(
	ctx context.Context,
	input cato_models.RemotePortFwdAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdAddRule, error) {
	return c.Client.PolicyRemotePortFwdAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdUpdateRule(
	ctx context.Context,
	input cato_models.RemotePortFwdUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdUpdateRule, error) {
	return c.Client.PolicyRemotePortFwdUpdateRule(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdRemoveRule(
	ctx context.Context,
	input cato_models.RemotePortFwdRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdRemoveRule, error) {
	return c.Client.PolicyRemotePortFwdRemoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdMoveRule, error) {
	return c.Client.PolicyRemotePortFwdMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdAddSection, error) {
	return c.Client.PolicyRemotePortFwdAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdUpdateSection, error) {
	return c.Client.PolicyRemotePortFwdUpdateSection(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdRemoveSection, error) {
	return c.Client.PolicyRemotePortFwdRemoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdMoveSection, error) {
	return c.Client.PolicyRemotePortFwdMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdUpdatePolicy(
	ctx context.Context,
	input cato_models.RemotePortFwdPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdUpdatePolicy, error) {
	return c.Client.PolicyRemotePortFwdUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyRemotePortFwdPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdPublishPolicyRevision, error) {
	return c.Client.PolicyRemotePortFwdPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
