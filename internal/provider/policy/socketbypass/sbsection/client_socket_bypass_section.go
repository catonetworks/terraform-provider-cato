package sbsection

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
	PolicySocketBypassAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassAddSection, error)
	PolicySocketBypassMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassMoveSection, error)
	PolicySocketBypassRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassRemoveSection, error)
	PolicySocketBypassUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassUpdateSection, error)
	PolicySocketBypassPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error)
}

func (r *socketBypassSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketBypassPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPolicy, error) {
	return c.resource.client.Catov2.PolicySocketBypassPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassAddSection, error) {
	return c.resource.client.Catov2.PolicySocketBypassAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassMoveSection, error) {
	return c.resource.client.Catov2.PolicySocketBypassMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassRemoveSection, error) {
	return c.resource.client.Catov2.PolicySocketBypassRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketBypassUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassUpdateSection, error) {
	return c.resource.client.Catov2.PolicySocketBypassUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *socketBypassSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySocketBypassPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketBypassPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySocketBypassPublishPolicyRevision(ctx, accountID, interceptors...)
}
