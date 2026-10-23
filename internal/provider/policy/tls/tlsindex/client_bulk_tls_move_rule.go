package tlsindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTLSInspectMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectMoveRule, error)
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
	Tlsinspectpolicy(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.Tlsinspectpolicy, error)
}

func (r *tlsRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyTLSInspectMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectMoveRule, error) {
	return c.resource.client.Catov2.PolicyTLSInspectMoveRule(ctx, input, accountID, interceptors...)
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

func (c *sdkClient) Tlsinspectpolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Tlsinspectpolicy, error) {
	return c.resource.client.Catov2.Tlsinspectpolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *tlsRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)
