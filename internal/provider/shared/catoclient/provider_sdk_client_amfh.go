package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// AntiMalwareFileHash policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashPolicy, error) {
	return c.Client.PolicyAntiMalwareFileHashPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashAddRule(
	ctx context.Context,
	input cato_models.AntiMalwareFileHashAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashAddRule, error) {
	return c.Client.PolicyAntiMalwareFileHashAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashUpdateRule(
	ctx context.Context,
	input cato_models.AntiMalwareFileHashUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashUpdateRule, error) {
	return c.Client.PolicyAntiMalwareFileHashUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashRemoveRule(
	ctx context.Context,
	input cato_models.AntiMalwareFileHashRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashRemoveRule, error) {
	return c.Client.PolicyAntiMalwareFileHashRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashMoveRule, error) {
	return c.Client.PolicyAntiMalwareFileHashMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashAddSection, error) {
	return c.Client.PolicyAntiMalwareFileHashAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashUpdateSection, error) {
	return c.Client.PolicyAntiMalwareFileHashUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashRemoveSection, error) {
	return c.Client.PolicyAntiMalwareFileHashRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashMoveSection, error) {
	return c.Client.PolicyAntiMalwareFileHashMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashUpdatePolicy(
	ctx context.Context,
	input cato_models.AntiMalwareFileHashPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashUpdatePolicy, error) {
	return c.Client.PolicyAntiMalwareFileHashUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyAntiMalwareFileHashPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAntiMalwareFileHashPublishPolicyRevision, error) {
	return c.Client.PolicyAntiMalwareFileHashPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
