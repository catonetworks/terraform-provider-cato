package strule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySplitTunnelPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelPolicy, error)
	PolicySplitTunnelAddRule(
		ctx context.Context,
		input cato_models.SplitTunnelAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelAddRule, error)
	PolicySplitTunnelMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelMoveRule, error)
	PolicySplitTunnelRemoveRule(
		ctx context.Context,
		input cato_models.SplitTunnelRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelRemoveRule, error)
	PolicySplitTunnelUpdateRule(
		ctx context.Context,
		input cato_models.SplitTunnelUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelUpdateRule, error)
	PolicySplitTunnelPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelPublishPolicyRevision, error)
}

func (r *splitTunnelRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySplitTunnelPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelPolicy, error) {
	return c.resource.client.Catov2.PolicySplitTunnelPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelAddRule(
	ctx context.Context,
	input cato_models.SplitTunnelAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelAddRule, error) {
	return c.resource.client.Catov2.PolicySplitTunnelAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelMoveRule, error) {
	return c.resource.client.Catov2.PolicySplitTunnelMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelRemoveRule(
	ctx context.Context,
	input cato_models.SplitTunnelRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelRemoveRule, error) {
	return c.resource.client.Catov2.PolicySplitTunnelRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelUpdateRule(
	ctx context.Context,
	input cato_models.SplitTunnelUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelUpdateRule, error) {
	return c.resource.client.Catov2.PolicySplitTunnelUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *splitTunnelRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySplitTunnelPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySplitTunnelPublishPolicyRevision(ctx, accountID, interceptors...)
}
