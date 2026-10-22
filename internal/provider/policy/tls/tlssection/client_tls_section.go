package tlssection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTLSInspectAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectAddSection, error)
	PolicyTLSInspectMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectMoveSection, error)
	PolicyTLSInspectPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error)
	PolicyTLSInspectRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectRemoveSection, error)
	PolicyTLSInspectUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectUpdateSection, error)
	Tlsinspectpolicy(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.Tlsinspectpolicy, error)
}

func (r *tlsInspectionSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyTLSInspectAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectAddSection, error) {
	return c.resource.client.Catov2.PolicyTLSInspectAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectMoveSection, error) {
	return c.resource.client.Catov2.PolicyTLSInspectMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTLSInspectPublishPolicyRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectRemoveSection, error) {
	return c.resource.client.Catov2.PolicyTLSInspectRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectUpdateSection, error) {
	return c.resource.client.Catov2.PolicyTLSInspectUpdateSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) Tlsinspectpolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Tlsinspectpolicy, error) {
	return c.resource.client.Catov2.Tlsinspectpolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *tlsInspectionSectionResource }

var _ ResourceClient = (*sdkClient)(nil)
