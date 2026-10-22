package ifsubpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// InternetFirewallSubPolicyClient is the narrow SDK surface used by the
// cato_if_sub_policy resource. It exists so the resource can be unit tested with
// generated mocks.
type InternetFirewallSubPolicyClient interface {
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallAddSubPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallAddSubPolicyInput cato_models.InternetFirewallAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSubPolicy, error)
	PolicyInternetFirewallRemoveSubPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallRemoveSubPolicyInput cato_models.InternetFirewallRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy, error)
	PolicyInternetFirewallUpdateRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput cato_models.InternetFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdateRule, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
}

func (r *ifSubPolicyResource) getClient() InternetFirewallSubPolicyClient {
	if r.subPolyClient != nil {
		return r.subPolyClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *ifSubPolicyResource) publish(ctx context.Context) error {
	_, err := r.resourceAPI().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		&cato_models.PolicyPublishRevisionInput{},
		r.client.AccountId,
	)
	return err
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallAddSubPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallAddSubPolicyInput cato_models.InternetFirewallAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSubPolicy, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
	PolicyInternetFirewallRemoveSubPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallRemoveSubPolicyInput cato_models.InternetFirewallRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy, error)
	PolicyInternetFirewallUpdateRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput cato_models.InternetFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdateRule, error)
}

func (r *ifSubPolicyResource) resourceAPI() ResourceClient {
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
	return c.resource.getClient().PolicyInternetFirewall(ctx, internetFirewallPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallAddSubPolicy(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	internetFirewallAddSubPolicyInput cato_models.InternetFirewallAddSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallAddSubPolicy, error) {
	return c.resource.getClient().PolicyInternetFirewallAddSubPolicy(
		ctx,
		internetFirewallPolicyMutationInput,
		internetFirewallAddSubPolicyInput,
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
	return c.resource.getClient().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		internetFirewallPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallRemoveSubPolicy(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	internetFirewallRemoveSubPolicyInput cato_models.InternetFirewallRemoveSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy, error) {
	return c.resource.getClient().PolicyInternetFirewallRemoveSubPolicy(
		ctx,
		internetFirewallPolicyMutationInput,
		internetFirewallRemoveSubPolicyInput,
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
	return c.resource.getClient().PolicyInternetFirewallUpdateRule(
		ctx,
		internetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *ifSubPolicyResource }

var _ ResourceClient = (*sdkClient)(nil)
