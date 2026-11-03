package acindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	ApplicationControlPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ApplicationControlPolicy, error)
	PolicyApplicationControlMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlMoveSection, error)
	PolicyApplicationControlMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlMoveRule, error)
	PolicyApplicationControlPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error)
}

func (r *apcRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *apcRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) ApplicationControlPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ApplicationControlPolicy, error) {
	return c.resource.client.Catov2.ApplicationControlPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlMoveSection, error) {
	return c.resource.client.Catov2.PolicyApplicationControlMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlMoveRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyApplicationControlPublishPolicyRevision(ctx, accountID, interceptors...)
}
