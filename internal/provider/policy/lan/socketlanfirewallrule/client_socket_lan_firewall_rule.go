package socketlanfirewallrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketLanFirewallAddRule(
		ctx context.Context,
		accountID string,
		policyInput *cato_models.SocketLanPolicyMutationInput,
		input cato_models.SocketLanFirewallAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanFirewallAddRule, error)
	PolicySocketLanFirewallMoveRule(
		ctx context.Context,
		accountID string,
		policyInput *cato_models.SocketLanPolicyMutationInput,
		input cato_models.PolicyMoveSubRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanFirewallMoveRule, error)
	PolicySocketLanFirewallRemoveRule(
		ctx context.Context,
		accountID string,
		policyInput *cato_models.SocketLanPolicyMutationInput,
		input cato_models.SocketLanFirewallRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanFirewallRemoveRule, error)
	PolicySocketLanFirewallUpdateRule(
		ctx context.Context,
		accountID string,
		policyInput *cato_models.SocketLanPolicyMutationInput,
		input cato_models.SocketLanFirewallUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanFirewallUpdateRule, error)
	PolicySocketLanPolicy(
		ctx context.Context,
		accountID string,
		input *cato_models.SocketLanPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPolicy, error)
	PolicySocketLanPublishPolicyRevision(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error)
}

func (r *socketLanFirewallRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketLanFirewallAddRule(
	ctx context.Context,
	accountID string,
	policyInput *cato_models.SocketLanPolicyMutationInput,
	input cato_models.SocketLanFirewallAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanFirewallAddRule, error) {
	return c.resource.client.Catov2.PolicySocketLanFirewallAddRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicySocketLanFirewallMoveRule(
	ctx context.Context,
	accountID string,
	policyInput *cato_models.SocketLanPolicyMutationInput,
	input cato_models.PolicyMoveSubRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanFirewallMoveRule, error) {
	return c.resource.client.Catov2.PolicySocketLanFirewallMoveRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicySocketLanFirewallRemoveRule(
	ctx context.Context,
	accountID string,
	policyInput *cato_models.SocketLanPolicyMutationInput,
	input cato_models.SocketLanFirewallRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanFirewallRemoveRule, error) {
	return c.resource.client.Catov2.PolicySocketLanFirewallRemoveRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicySocketLanFirewallUpdateRule(
	ctx context.Context,
	accountID string,
	policyInput *cato_models.SocketLanPolicyMutationInput,
	input cato_models.SocketLanFirewallUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanFirewallUpdateRule, error) {
	return c.resource.client.Catov2.PolicySocketLanFirewallUpdateRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicySocketLanPolicy(
	ctx context.Context,
	accountID string,
	input *cato_models.SocketLanPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPolicy, error) {
	return c.resource.client.Catov2.PolicySocketLanPolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicySocketLanPublishPolicyRevision(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySocketLanPublishPolicyRevision(
		ctx,
		socketLanPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *socketLanFirewallRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)
