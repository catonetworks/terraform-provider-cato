package sbrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketBypassPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassPolicy, error)
	PolicySocketBypassAddRule(
		ctx context.Context,
		input cato_models.SocketBypassAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassAddRule, error)
	PolicySocketBypassMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassMoveRule, error)
	PolicySocketBypassRemoveRule(
		ctx context.Context,
		input cato_models.SocketBypassRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassRemoveRule, error)
	PolicySocketBypassUpdateRule(
		ctx context.Context,
		input cato_models.SocketBypassUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassUpdateRule, error)
	PolicySocketBypassPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error)
}

func (r *socketBypassRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketBypassPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPolicy, error) {
	return c.resource.client.Catov2.PolicySocketBypassPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassAddRule(
	ctx context.Context,
	input cato_models.SocketBypassAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassAddRule, error) {
	return c.resource.client.Catov2.PolicySocketBypassAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassMoveRule, error) {
	return c.resource.client.Catov2.PolicySocketBypassMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassRemoveRule(
	ctx context.Context,
	input cato_models.SocketBypassRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassRemoveRule, error) {
	return c.resource.client.Catov2.PolicySocketBypassRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassUpdateRule(
	ctx context.Context,
	input cato_models.SocketBypassUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassUpdateRule, error) {
	return c.resource.client.Catov2.PolicySocketBypassUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *socketBypassRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySocketBypassPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySocketBypassPublishPolicyRevision(ctx, accountID, interceptors...)
}
