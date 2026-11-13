package rpfsection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	RemotePortFwdPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.RemotePortFwdPolicy, error)
	PolicyRemotePortFwdAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdAddSection, error)
	PolicyRemotePortFwdMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdMoveSection, error)
	PolicyRemotePortFwdRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdRemoveSection, error)
	PolicyRemotePortFwdUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdUpdateSection, error)
	PolicyRemotePortFwdPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error)
}

func (r *remotePortFwdSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) RemotePortFwdPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.RemotePortFwdPolicy, error) {
	return c.resource.client.Catov2.RemotePortFwdPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdAddSection, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdMoveSection, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdRemoveSection, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdUpdateSection, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *remotePortFwdSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyRemotePortFwdPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdPublishPolicyRevision(ctx, accountID, interceptors...)
}
