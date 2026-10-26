package socketlannetworkrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketLanAddRule(
		ctx context.Context,
		input cato_models.SocketLanAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanAddRule, error)
	PolicySocketLanMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanMoveRule, error)
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
	PolicySocketLanRemoveRule(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		socketLanRemoveRuleInput cato_models.SocketLanRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanRemoveRule, error)
	PolicySocketLanUpdateRule(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		socketLanUpdateRuleInput cato_models.SocketLanUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanUpdateRule, error)
}

func (r *socketLanNetworkRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketLanAddRule(
	ctx context.Context,
	input cato_models.SocketLanAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanAddRule, error) {
	return c.resource.client.Catov2.PolicySocketLanAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketLanMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanMoveRule, error) {
	return c.resource.client.Catov2.PolicySocketLanMoveRule(ctx, input, accountID, interceptors...)
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

func (c *sdkClient) PolicySocketLanRemoveRule(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	socketLanRemoveRuleInput cato_models.SocketLanRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanRemoveRule, error) {
	return c.resource.client.Catov2.PolicySocketLanRemoveRule(
		ctx,
		socketLanPolicyMutationInput,
		socketLanRemoveRuleInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketLanUpdateRule(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	socketLanUpdateRuleInput cato_models.SocketLanUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanUpdateRule, error) {
	return c.resource.client.Catov2.PolicySocketLanUpdateRule(
		ctx,
		socketLanPolicyMutationInput,
		socketLanUpdateRuleInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *socketLanNetworkRuleResource }

var _ ResourceClient = (*sdkClient)(nil)
