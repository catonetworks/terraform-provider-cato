package rpfrule

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
	PolicyRemotePortFwdAddRule(
		ctx context.Context,
		input cato_models.RemotePortFwdAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdAddRule, error)
	PolicyRemotePortFwdMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdMoveRule, error)
	PolicyRemotePortFwdRemoveRule(
		ctx context.Context,
		input cato_models.RemotePortFwdRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdRemoveRule, error)
	PolicyRemotePortFwdUpdateRule(
		ctx context.Context,
		input cato_models.RemotePortFwdUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdUpdateRule, error)
	PolicyRemotePortFwdPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error)
}

func (r *remotePortFwdRuleResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) PolicyRemotePortFwdAddRule(
	ctx context.Context,
	input cato_models.RemotePortFwdAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdAddRule, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdMoveRule, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdRemoveRule(
	ctx context.Context,
	input cato_models.RemotePortFwdRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdRemoveRule, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyRemotePortFwdUpdateRule(
	ctx context.Context,
	input cato_models.RemotePortFwdUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdUpdateRule, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *remotePortFwdRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyRemotePortFwdPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyRemotePortFwdPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyRemotePortFwdPublishPolicyRevision(ctx, accountID, interceptors...)
}
