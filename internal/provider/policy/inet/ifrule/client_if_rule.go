package ifrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type InternetFirewallPolicyClient interface {
	PolicyInternetFirewallAddRule(
		ctx context.Context,
		internetFirewallAddRuleInput cato_models.InternetFirewallAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddRule, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallMoveRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallMoveRule, error)
	PolicyInternetFirewallUpdateRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput cato_models.InternetFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdateRule, error)
	PolicyInternetFirewallRemoveRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallRemoveRuleInput cato_models.InternetFirewallRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallRemoveRule, error)
}

func (r *InternetFwRuleResource) getIfwClient() InternetFirewallPolicyClient {
	if r.IfwClient != nil {
		return r.IfwClient
	}

	if r.Client == nil {
		return nil
	}

	return r.Client.Catov2
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallAddRule(
		ctx context.Context,
		internetFirewallAddRuleInput cato_models.InternetFirewallAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddRule, error)
	PolicyInternetFirewallMoveRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallMoveRule, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
	PolicyInternetFirewallRemoveRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallRemoveRuleInput cato_models.InternetFirewallRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallRemoveRule, error)
	PolicyInternetFirewallUpdateRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput cato_models.InternetFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdateRule, error)
}

func (r *InternetFwRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyInternetFirewall(
	ctx context.Context,
	internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Policy, error) {
	return c.resource.getIfwClient().PolicyInternetFirewall(ctx, internetFirewallPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallAddRule(
	ctx context.Context,
	internetFirewallAddRuleInput cato_models.InternetFirewallAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallAddRule, error) {
	return c.resource.getIfwClient().PolicyInternetFirewallAddRule(ctx, internetFirewallAddRuleInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallMoveRule(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallMoveRule, error) {
	return c.resource.getIfwClient().PolicyInternetFirewallMoveRule(
		ctx,
		internetFirewallPolicyMutationInput,
		policyMoveRuleInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallPublishPolicyRevision(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error) {
	return c.resource.getIfwClient().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		internetFirewallPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallRemoveRule(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	internetFirewallRemoveRuleInput cato_models.InternetFirewallRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallRemoveRule, error) {
	return c.resource.getIfwClient().PolicyInternetFirewallRemoveRule(
		ctx,
		internetFirewallPolicyMutationInput,
		internetFirewallRemoveRuleInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallUpdateRule(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	internetFirewallUpdateRuleInput cato_models.InternetFirewallUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallUpdateRule, error) {
	return c.resource.getIfwClient().PolicyInternetFirewallUpdateRule(
		ctx,
		internetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *InternetFwRuleResource }

var _ ResourceClient = (*sdkClient)(nil)
