package sbindex

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
	PolicySocketBypassMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassMoveSection, error)
	PolicySocketBypassMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassMoveRule, error)
	PolicySocketBypassPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error)
}

func (r *sbRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *sbRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySocketBypassPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPolicy, error) {
	return c.resource.client.Catov2.PolicySocketBypassPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicySocketBypassMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassMoveSection, error) {
	return c.resource.client.Catov2.PolicySocketBypassMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicySocketBypassMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassMoveRule, error) {
	return c.resource.client.Catov2.PolicySocketBypassMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicySocketBypassPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySocketBypassPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
