package rpfindex

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
	PolicyRemotePortFwdMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdMoveSection, error)
	PolicyRemotePortFwdMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdMoveRule, error)
	PolicyRemotePortFwdPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error)
}

func (r *rpfRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *rpfRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) RemotePortFwdPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.RemotePortFwdPolicy, error) {
	return c.resource.client.Catov2.RemotePortFwdPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdMoveSection, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdMoveRule, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
