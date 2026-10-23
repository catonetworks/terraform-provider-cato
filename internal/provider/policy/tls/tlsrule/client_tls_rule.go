package tlsrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTLSInspectAddRule(
		ctx context.Context,
		input cato_models.TLSInspectAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectAddRule, error)
	PolicyTLSInspectMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectMoveRule, error)
	PolicyTLSInspectPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error)
	PolicyTLSInspectRemoveRule(
		ctx context.Context,
		input cato_models.TLSInspectRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectRemoveRule, error)
	PolicyTLSInspectUpdateRule(
		ctx context.Context,
		input cato_models.TLSInspectUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectUpdateRule, error)
	Tlsinspectpolicy(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.Tlsinspectpolicy, error)
}

func (r *tlsInspectionRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyTLSInspectAddRule(
	ctx context.Context,
	input cato_models.TLSInspectAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectAddRule, error) {
	return c.resource.client.Catov2.PolicyTLSInspectAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectMoveRule, error) {
	return c.resource.client.Catov2.PolicyTLSInspectMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTLSInspectPublishPolicyRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectRemoveRule(
	ctx context.Context,
	input cato_models.TLSInspectRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectRemoveRule, error) {
	return c.resource.client.Catov2.PolicyTLSInspectRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectUpdateRule(
	ctx context.Context,
	input cato_models.TLSInspectUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectUpdateRule, error) {
	return c.resource.client.Catov2.PolicyTLSInspectUpdateRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) Tlsinspectpolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Tlsinspectpolicy, error) {
	return c.resource.client.Catov2.Tlsinspectpolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *tlsInspectionRuleResource }

var _ ResourceClient = (*sdkClient)(nil)
