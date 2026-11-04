package ccindex

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
	PolicyClientConnectivityMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityMoveSection, error)
	PolicyClientConnectivityMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityMoveRule, error)
	PolicyClientConnectivityPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error)
}

func (r *ccRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *ccRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyClientConnectivityPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPolicy, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityMoveSection, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityMoveRule, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
