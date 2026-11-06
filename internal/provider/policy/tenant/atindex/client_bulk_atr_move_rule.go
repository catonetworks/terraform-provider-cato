package atindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AppTenantRestrictionPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppTenantRestrictionPolicy, error)
	PolicyAppTenantRestrictionMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionMoveSection, error)
	PolicyAppTenantRestrictionMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionMoveRule, error)
	PolicyAppTenantRestrictionPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error)
}

func (r *atrRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *atrRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) AppTenantRestrictionPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppTenantRestrictionPolicy, error) {
	return c.resource.client.Catov2.AppTenantRestrictionPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionMoveSection, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionMoveRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, accountID, interceptors...)
}
