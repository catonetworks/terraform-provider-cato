package ssiapolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySdpStaticIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPolicy, error)
	PolicySdpStaticIPAllocationUpdatePolicy(
		ctx context.Context,
		input cato_models.SdpStaticIPAllocationPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationUpdatePolicy, error)
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
}

func (r *sdpStaticIpAllocationPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySdpStaticIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationUpdatePolicy(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationUpdatePolicy(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *sdpStaticIpAllocationPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySdpStaticIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, accountID, interceptors...)
}
