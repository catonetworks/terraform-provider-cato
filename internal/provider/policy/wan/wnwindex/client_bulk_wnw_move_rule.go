package wnwindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyWanNetworkMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkMoveRule, error)
	PolicyWanNetworkMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkMoveSection, error)
	PolicyWanNetworkPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error)
	WanNetworkPolicy(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.WanNetworkPolicy, error)
}

func (r *wanNetworkRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyWanNetworkMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkMoveRule, error) {
	return c.resource.client.Catov2.PolicyWanNetworkMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkMoveSection, error) {
	return c.resource.client.Catov2.PolicyWanNetworkMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) WanNetworkPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanNetworkPolicy, error) {
	return c.resource.client.Catov2.WanNetworkPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *wanNetworkRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)
