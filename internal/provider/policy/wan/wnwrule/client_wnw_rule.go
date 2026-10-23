package wnwrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyWanNetworkAddRule(
		ctx context.Context,
		input cato_models.WanNetworkAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkAddRule, error)
	PolicyWanNetworkPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error)
	PolicyWanNetworkRemoveRule(
		ctx context.Context,
		input cato_models.WanNetworkRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkRemoveRule, error)
	PolicyWanNetworkUpdateRule(
		ctx context.Context,
		input cato_models.WanNetworkUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkUpdateRule, error)
	WanNetworkPolicy(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.WanNetworkPolicy, error)
}

func (r *wanNetworkRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyWanNetworkAddRule(
	ctx context.Context,
	input cato_models.WanNetworkAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkAddRule, error) {
	return c.resource.client.Catov2.PolicyWanNetworkAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkRemoveRule(
	ctx context.Context,
	input cato_models.WanNetworkRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkRemoveRule, error) {
	return c.resource.client.Catov2.PolicyWanNetworkRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkUpdateRule(
	ctx context.Context,
	input cato_models.WanNetworkUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkUpdateRule, error) {
	return c.resource.client.Catov2.PolicyWanNetworkUpdateRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) WanNetworkPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanNetworkPolicy, error) {
	return c.resource.client.Catov2.WanNetworkPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *wanNetworkRuleResource }

var _ ResourceClient = (*sdkClient)(nil)
