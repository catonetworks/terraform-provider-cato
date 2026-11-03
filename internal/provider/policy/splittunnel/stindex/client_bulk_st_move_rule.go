package stindex

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
	PolicySplitTunnelMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelMoveSection, error)
	PolicySplitTunnelMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelMoveRule, error)
	PolicySplitTunnelPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelPublishPolicyRevision, error)
}

func (r *stRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *stRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySplitTunnelPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelPolicy, error) {
	return c.resource.client.Catov2.PolicySplitTunnelPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelMoveSection, error) {
	return c.resource.client.Catov2.PolicySplitTunnelMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelMoveRule, error) {
	return c.resource.client.Catov2.PolicySplitTunnelMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicySplitTunnelPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySplitTunnelPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySplitTunnelPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
