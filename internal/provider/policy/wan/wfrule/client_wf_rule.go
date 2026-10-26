package wfrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyWanFirewall(
		ctx context.Context,
		wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyWanFirewallAddRule(
		ctx context.Context,
		input cato_models.WanFirewallAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallAddRule, error)
	PolicyWanFirewallMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallMoveRule, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		input *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
	PolicyWanFirewallRemoveRule(
		ctx context.Context,
		input cato_models.WanFirewallRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallRemoveRule, error)
	PolicyWanFirewallUpdateRule(
		ctx context.Context,
		input cato_models.WanFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallUpdateRule, error)
}

func (r *WanFwRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyWanFirewall(
	ctx context.Context,
	wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Policy, error) {
	return c.resource.client.Catov2.PolicyWanFirewall(ctx, wanFirewallPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallAddRule(
	ctx context.Context,
	input cato_models.WanFirewallAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallAddRule, error) {
	return c.resource.client.Catov2.PolicyWanFirewallAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallMoveRule, error) {
	return c.resource.client.Catov2.PolicyWanFirewallMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallPublishPolicyRevision(
	ctx context.Context,
	input *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanFirewallPublishPolicyRevision(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallRemoveRule(
	ctx context.Context,
	input cato_models.WanFirewallRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallRemoveRule, error) {
	return c.resource.client.Catov2.PolicyWanFirewallRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallUpdateRule(
	ctx context.Context,
	input cato_models.WanFirewallUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallUpdateRule, error) {
	return c.resource.client.Catov2.PolicyWanFirewallUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *WanFwRuleResource }

var _ ResourceClient = (*sdkClient)(nil)
