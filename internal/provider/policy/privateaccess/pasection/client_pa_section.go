package pasection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error)
	PolicyPrivateAccessAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessAddSection, error)
	PolicyPrivateAccessUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessUpdateSection, error)
	PolicyPrivateAccessRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessRemoveSection, error)
	PolicyPrivateAccessMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessMoveSection, error)
	PolicyPrivateAccessPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessPublishPolicyRevision, error)
}

func (r *privateAccessSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct {
	resource *privateAccessSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error) {
	return c.resource.client.Catov2.PolicyReadPrivateAccessPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessAddSection, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessUpdateSection, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessUpdateSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessRemoveSection, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessMoveSection, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
