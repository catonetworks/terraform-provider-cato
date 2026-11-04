package diaindex

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
	PolicyDynamicIPAllocationMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationMoveSection, error)
	PolicyDynamicIPAllocationMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationMoveRule, error)
	PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error)
}

func (r *diaRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *diaRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyDynamicIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationMoveSection, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationMoveRule, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyDynamicIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyDynamicIPAllocationPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
