package tlspolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	Tlsinspectpolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Tlsinspectpolicy, error)
	PolicyTLSInspectUpdatePolicy(
		ctx context.Context,
		input cato_models.TLSInspectPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectUpdatePolicy, error)
	PolicyTLSInspectPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error)
}

func (r *tlsInspectPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct {
	resource *tlsInspectPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) Tlsinspectpolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Tlsinspectpolicy, error) {
	return c.resource.client.Catov2.Tlsinspectpolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectUpdatePolicy(
	ctx context.Context,
	input cato_models.TLSInspectPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyTLSInspectUpdatePolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyTLSInspectPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTLSInspectPublishPolicyRevision(ctx, accountID, interceptors...)
}
