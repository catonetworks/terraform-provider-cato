package wfpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	WanFirewallPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanFirewallPolicy, error)
	PolicyWanFirewallUpdatePolicy(
		ctx context.Context,
		input cato_models.WanFirewallPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallUpdatePolicy, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
}

func (r *wfPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct {
	resource *wfPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) WanFirewallPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanFirewallPolicy, error) {
	return c.resource.client.Catov2.WanFirewallPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallUpdatePolicy(
	ctx context.Context,
	input cato_models.WanFirewallPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyWanFirewallUpdatePolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanFirewallPublishPolicyRevision(ctx, nil, accountID, interceptors...)
}
