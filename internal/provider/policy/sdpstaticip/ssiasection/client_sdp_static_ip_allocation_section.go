package ssiasection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySdpStaticIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPolicy, error)
	PolicySdpStaticIPAllocationAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationAddSection, error)
	PolicySdpStaticIPAllocationMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveSection, error)
	PolicySdpStaticIPAllocationRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationRemoveSection, error)
	PolicySdpStaticIPAllocationUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationUpdateSection, error)
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
}

func (r *sdpStaticIPAllocationSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySdpStaticIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationAddSection, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveSection, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationRemoveSection, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationUpdateSection, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *sdpStaticIPAllocationSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySdpStaticIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, accountID, interceptors...)
}
