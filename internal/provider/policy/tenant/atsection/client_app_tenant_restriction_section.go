package atsection

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
	PolicyAppTenantRestrictionAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionAddSection, error)
	PolicyAppTenantRestrictionMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionMoveSection, error)
	PolicyAppTenantRestrictionRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionRemoveSection, error)
	PolicyAppTenantRestrictionUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionUpdateSection, error)
	PolicyAppTenantRestrictionPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error)
}

func (r *appTenantRestrictionSectionResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) PolicyAppTenantRestrictionAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionAddSection, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionMoveSection, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionRemoveSection, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionUpdateSection, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *appTenantRestrictionSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyAppTenantRestrictionPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, accountID, interceptors...)
}
