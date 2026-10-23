package lansection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketLanAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanAddSection, error)
	PolicySocketLanMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanMoveSection, error)
	PolicySocketLanPolicy(
		ctx context.Context,
		accountID string,
		input *cato_models.SocketLanPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPolicy, error)
	PolicySocketLanPublishPolicyRevision(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error)
	PolicySocketLanRemoveSection(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanRemoveSection, error)
	PolicySocketLanUpdateSection(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanUpdateSection, error)
}

func (r *socketLanSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketLanAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanAddSection, error) {
	return c.resource.client.Catov2.PolicySocketLanAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketLanMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanMoveSection, error) {
	return c.resource.client.Catov2.PolicySocketLanMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketLanPolicy(
	ctx context.Context,
	accountID string,
	input *cato_models.SocketLanPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPolicy, error) {
	return c.resource.client.Catov2.PolicySocketLanPolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicySocketLanPublishPolicyRevision(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySocketLanPublishPolicyRevision(
		ctx,
		socketLanPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketLanRemoveSection(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanRemoveSection, error) {
	return c.resource.client.Catov2.PolicySocketLanRemoveSection(
		ctx,
		socketLanPolicyMutationInput,
		policyRemoveSectionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketLanUpdateSection(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanUpdateSection, error) {
	return c.resource.client.Catov2.PolicySocketLanUpdateSection(
		ctx,
		socketLanPolicyMutationInput,
		policyUpdateSectionInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *socketLanSectionResource }

var _ ResourceClient = (*sdkClient)(nil)
