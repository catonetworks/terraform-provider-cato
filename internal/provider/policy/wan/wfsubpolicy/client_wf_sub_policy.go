package wfsubpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func (r *wfSubPolicyResource) getClient() WanFirewallSubPolicyClient {
	if r.subPolyClient != nil {
		return r.subPolyClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

// WanFirewallSubPolicyClient is the narrow SDK surface used by the
// cato_wf_sub_policy resource.
type WanFirewallSubPolicyClient interface {
	PolicyWanFirewall(
		ctx context.Context,
		wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyWanFirewallAddSubPolicy(
		ctx context.Context,
		wanFirewallAddSubPolicyInput cato_models.WanFirewallAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallAddSubPolicy, error)
	PolicyWanFirewallRemoveSubPolicy(
		ctx context.Context,
		wanFirewallRemoveSubPolicyInput cato_models.WanFirewallRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallRemoveSubPolicy, error)
	PolicyWanFirewallUpdateRule(
		ctx context.Context,
		wanFirewallUpdateRuleInput cato_models.WanFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallUpdateRule, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyWanFirewall(
		ctx context.Context,
		wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyWanFirewallAddSubPolicy(
		ctx context.Context,
		wanFirewallAddSubPolicyInput cato_models.WanFirewallAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallAddSubPolicy, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
	PolicyWanFirewallRemoveSubPolicy(
		ctx context.Context,
		wanFirewallRemoveSubPolicyInput cato_models.WanFirewallRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallRemoveSubPolicy, error)
	PolicyWanFirewallUpdateRule(
		ctx context.Context,
		wanFirewallUpdateRuleInput cato_models.WanFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallUpdateRule, error)
}

func (r *wfSubPolicyResource) resourceAPI() ResourceClient {
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
	return c.resource.getClient().PolicyWanFirewall(ctx, wanFirewallPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallAddSubPolicy(
	ctx context.Context,
	wanFirewallAddSubPolicyInput cato_models.WanFirewallAddSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallAddSubPolicy, error) {
	return c.resource.getClient().PolicyWanFirewallAddSubPolicy(ctx, wanFirewallAddSubPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallPublishPolicyRevision(
	ctx context.Context,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error) {
	return c.resource.getClient().PolicyWanFirewallPublishPolicyRevision(ctx, policyPublishRevisionInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallRemoveSubPolicy(
	ctx context.Context,
	wanFirewallRemoveSubPolicyInput cato_models.WanFirewallRemoveSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallRemoveSubPolicy, error) {
	return c.resource.getClient().PolicyWanFirewallRemoveSubPolicy(ctx, wanFirewallRemoveSubPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallUpdateRule(
	ctx context.Context,
	wanFirewallUpdateRuleInput cato_models.WanFirewallUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallUpdateRule, error) {
	return c.resource.getClient().PolicyWanFirewallUpdateRule(ctx, wanFirewallUpdateRuleInput, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *wfSubPolicyResource }

var _ ResourceClient = (*sdkClient)(nil)
