package tspolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTerminalServerPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPolicy, error)
	PolicyTerminalServerUpdatePolicy(
		ctx context.Context,
		input cato_models.TerminalServerPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerUpdatePolicy, error)
	PolicyTerminalServerPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error)
}

func (r *terminalServerPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyTerminalServerPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPolicy, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerUpdatePolicy(
	ctx context.Context,
	input cato_models.TerminalServerPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyTerminalServerUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *terminalServerPolicyResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyTerminalServerPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
