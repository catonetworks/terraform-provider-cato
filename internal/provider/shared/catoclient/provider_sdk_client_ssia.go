package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// SdpStaticIpAllocation policy family wrappers (draft-targeted: nil mutation inputs).

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationPolicy, error) {
	return c.Client.PolicySdpStaticIPAllocationPolicy(ctx, accountID, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationAddRule(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationAddRule, error) {
	return c.Client.PolicySdpStaticIPAllocationAddRule(ctx, accountID, input, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationUpdateRule(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationUpdateRule, error) {
	return c.Client.PolicySdpStaticIPAllocationUpdateRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationRemoveRule(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationRemoveRule, error) {
	return c.Client.PolicySdpStaticIPAllocationRemoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationMoveRule, error) {
	return c.Client.PolicySdpStaticIPAllocationMoveRule(ctx, accountID, input, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationAddSection, error) {
	return c.Client.PolicySdpStaticIPAllocationAddSection(ctx, accountID, input, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationUpdateSection, error) {
	return c.Client.PolicySdpStaticIPAllocationUpdateSection(ctx, accountID, input, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationRemoveSection, error) {
	return c.Client.PolicySdpStaticIPAllocationRemoveSection(ctx, accountID, input, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationMoveSection, error) {
	return c.Client.PolicySdpStaticIPAllocationMoveSection(ctx, accountID, input, nil, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationUpdatePolicy(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationUpdatePolicy, error) {
	return c.Client.PolicySdpStaticIPAllocationUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *ProviderSDKClient) PolicySdpStaticIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationPublishPolicyRevision, error) {
	return c.Client.PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
