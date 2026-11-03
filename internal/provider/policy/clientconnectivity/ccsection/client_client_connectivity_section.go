package ccsection

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
	PolicyClientConnectivityAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityAddSection, error)
	PolicyClientConnectivityMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityMoveSection, error)
	PolicyClientConnectivityRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityRemoveSection, error)
	PolicyClientConnectivityUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityUpdateSection, error)
	PolicyClientConnectivityPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error)
}

func (r *clientConnectivitySectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyClientConnectivityPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPolicy, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityAddSection, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityMoveSection, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityRemoveSection, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyClientConnectivityUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityUpdateSection, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *clientConnectivitySectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyClientConnectivityPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyClientConnectivityPublishPolicyRevision(ctx, accountID, interceptors...)
}
