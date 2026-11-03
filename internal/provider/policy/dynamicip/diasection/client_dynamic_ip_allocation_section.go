package diasection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyDynamicIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error)
	PolicyDynamicIPAllocationAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationAddSection, error)
	PolicyDynamicIPAllocationMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationMoveSection, error)
	PolicyDynamicIPAllocationRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationRemoveSection, error)
	PolicyDynamicIPAllocationUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationUpdateSection, error)
	PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error)
}

func (r *dynamicIPAllocationSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyDynamicIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationAddSection, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationMoveSection, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationRemoveSection, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationUpdateSection, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *dynamicIPAllocationSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyDynamicIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPublishPolicyRevision(ctx, accountID, interceptors...)
}
