package wnwsection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyWanNetworkAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkAddSection, error)
	PolicyWanNetworkMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkMoveSection, error)
	PolicyWanNetworkPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error)
	PolicyWanNetworkRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkRemoveSection, error)
	PolicyWanNetworkUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanNetworkUpdateSection, error)
	WanNetworkPolicy(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.WanNetworkPolicy, error)
}

func (r *wanNetworkSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyWanNetworkAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkAddSection, error) {
	return c.resource.client.Catov2.PolicyWanNetworkAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkMoveSection, error) {
	return c.resource.client.Catov2.PolicyWanNetworkMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkRemoveSection, error) {
	return c.resource.client.Catov2.PolicyWanNetworkRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanNetworkUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanNetworkUpdateSection, error) {
	return c.resource.client.Catov2.PolicyWanNetworkUpdateSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) WanNetworkPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanNetworkPolicy, error) {
	return c.resource.client.Catov2.WanNetworkPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *wanNetworkSectionResource }

var _ ResourceClient = (*sdkClient)(nil)
