package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// DynamicIpAllocation policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicyDynamicIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationPolicy, error) {
	return c.Client.PolicyDynamicIPAllocationPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationAddRule(
	ctx context.Context,
	input cato_models.DynamicIPAllocationAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationAddRule, error) {
	return c.Client.PolicyDynamicIPAllocationAddRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationUpdateRule(
	ctx context.Context,
	input cato_models.DynamicIPAllocationUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationUpdateRule, error) {
	return c.Client.PolicyDynamicIPAllocationUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationRemoveRule(
	ctx context.Context,
	input cato_models.DynamicIPAllocationRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationRemoveRule, error) {
	return c.Client.PolicyDynamicIPAllocationRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationMoveRule, error) {
	return c.Client.PolicyDynamicIPAllocationMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationAddSection, error) {
	return c.Client.PolicyDynamicIPAllocationAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationUpdateSection, error) {
	return c.Client.PolicyDynamicIPAllocationUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationRemoveSection, error) {
	return c.Client.PolicyDynamicIPAllocationRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationMoveSection, error) {
	return c.Client.PolicyDynamicIPAllocationMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationUpdatePolicy(
	ctx context.Context,
	input cato_models.DynamicIPAllocationPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationUpdatePolicy, error) {
	return c.Client.PolicyDynamicIPAllocationUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicyDynamicIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationPublishPolicyRevision, error) {
	return c.Client.PolicyDynamicIPAllocationPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
