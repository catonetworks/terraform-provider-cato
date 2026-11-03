package tsindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTerminalServerPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPolicy, error)
	PolicyTerminalServerMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerMoveSection, error)
	PolicyTerminalServerMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerMoveRule, error)
	PolicyTerminalServerPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error)
}

func (r *tsRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *tsRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyTerminalServerPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPolicy, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerMoveSection, error) {
	return c.resource.client.Catov2.PolicyTerminalServerMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerMoveRule, error) {
	return c.resource.client.Catov2.PolicyTerminalServerMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
