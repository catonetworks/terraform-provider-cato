package ccrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyClientConnectivityPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityPolicy, error)
	PolicyClientConnectivityAddRule(
		ctx context.Context,
		input cato_models.ClientConnectivityAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityAddRule, error)
	PolicyClientConnectivityMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityMoveRule, error)
	PolicyClientConnectivityRemoveRule(
		ctx context.Context,
		input cato_models.ClientConnectivityRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityRemoveRule, error)
	PolicyClientConnectivityUpdateRule(
		ctx context.Context,
		input cato_models.ClientConnectivityUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityUpdateRule, error)
	PolicyClientConnectivityPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error)
}

func (r *clientConnectivityRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyClientConnectivityPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPolicy, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityAddRule(
	ctx context.Context,
	input cato_models.ClientConnectivityAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityAddRule, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityMoveRule, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityRemoveRule(
	ctx context.Context,
	input cato_models.ClientConnectivityRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityRemoveRule, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityUpdateRule(
	ctx context.Context,
	input cato_models.ClientConnectivityUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityUpdateRule, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *clientConnectivityRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyClientConnectivityPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPublishPolicyRevision(ctx, accountID, interceptors...)
}
