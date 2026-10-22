package acsection

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
	PolicyApplicationControlAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlAddSection, error)
	PolicyApplicationControlMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlMoveSection, error)
	PolicyApplicationControlRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlRemoveSection, error)
	PolicyApplicationControlUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlUpdateSection, error)
	PolicyApplicationControlPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error)
}

func (r *applicationControlSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) ApplicationControlPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ApplicationControlPolicy, error) {
	return c.resource.client.Catov2.ApplicationControlPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlAddSection, error) {
	return c.resource.client.Catov2.PolicyApplicationControlAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlMoveSection, error) {
	return c.resource.client.Catov2.PolicyApplicationControlMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlRemoveSection, error) {
	return c.resource.client.Catov2.PolicyApplicationControlRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlUpdateSection, error) {
	return c.resource.client.Catov2.PolicyApplicationControlUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *applicationControlSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyApplicationControlPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyApplicationControlPublishPolicyRevision(ctx, accountID, interceptors...)
}
