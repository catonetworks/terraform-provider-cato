package ssiarule

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
	PolicySdpStaticIPAllocationAddRule(
		ctx context.Context,
		input cato_models.SdpStaticIPAllocationAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationAddRule, error)
	PolicySdpStaticIPAllocationMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveRule, error)
	PolicySdpStaticIPAllocationRemoveRule(
		ctx context.Context,
		input cato_models.SdpStaticIPAllocationRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationRemoveRule, error)
	PolicySdpStaticIPAllocationUpdateRule(
		ctx context.Context,
		input cato_models.SdpStaticIPAllocationUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationUpdateRule, error)
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
}

func (r *sdpStaticIPAllocationRuleResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) PolicySdpStaticIPAllocationAddRule(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationAddRule, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveRule, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationRemoveRule(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationRemoveRule, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationUpdateRule(
	ctx context.Context,
	input cato_models.SdpStaticIPAllocationUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationUpdateRule, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *sdpStaticIPAllocationRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySdpStaticIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, accountID, interceptors...)
}
