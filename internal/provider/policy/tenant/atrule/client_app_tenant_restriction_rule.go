package atrule

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
	PolicyAppTenantRestrictionAddRule(
		ctx context.Context,
		input cato_models.AppTenantRestrictionAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionAddRule, error)
	PolicyAppTenantRestrictionMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionMoveRule, error)
	PolicyAppTenantRestrictionRemoveRule(
		ctx context.Context,
		input cato_models.AppTenantRestrictionRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionRemoveRule, error)
	PolicyAppTenantRestrictionUpdateRule(
		ctx context.Context,
		input cato_models.AppTenantRestrictionUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionUpdateRule, error)
	PolicyAppTenantRestrictionPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error)
}

func (r *appTenantRestrictionRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) AppTenantRestrictionPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppTenantRestrictionPolicy, error) {
	return c.resource.client.Catov2.AppTenantRestrictionPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionAddRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionAddRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionMoveRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionRemoveRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionRemoveRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionUpdateRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionUpdateRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *appTenantRestrictionRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyAppTenantRestrictionPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, accountID, interceptors...)
}
