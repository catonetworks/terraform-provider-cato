package diarule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyDynamicIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error)
	PolicyDynamicIPAllocationAddRule(
		ctx context.Context,
		input cato_models.DynamicIPAllocationAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationAddRule, error)
	PolicyDynamicIPAllocationMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationMoveRule, error)
	PolicyDynamicIPAllocationRemoveRule(
		ctx context.Context,
		input cato_models.DynamicIPAllocationRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationRemoveRule, error)
	PolicyDynamicIPAllocationUpdateRule(
		ctx context.Context,
		input cato_models.DynamicIPAllocationUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationUpdateRule, error)
	PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error)
}

func (r *dynamicIpAllocationRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyDynamicIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationAddRule(
	ctx context.Context,
	input cato_models.DynamicIPAllocationAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationAddRule, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationMoveRule, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationRemoveRule(
	ctx context.Context,
	input cato_models.DynamicIPAllocationRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationRemoveRule, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationUpdateRule(
	ctx context.Context,
	input cato_models.DynamicIPAllocationUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationUpdateRule, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *dynamicIpAllocationRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyDynamicIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPublishPolicyRevision(ctx, accountID, interceptors...)
}
