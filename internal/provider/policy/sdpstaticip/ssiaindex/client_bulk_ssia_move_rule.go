package ssiaindex

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
	PolicySdpStaticIPAllocationMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveSection, error)
	PolicySdpStaticIPAllocationMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveRule, error)
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
}

func (r *ssiaRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *ssiaRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySdpStaticIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPolicy, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveSection, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationMoveSection(ctx, accountID, input, nil, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationMoveRule, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationMoveRule(ctx, accountID, input, nil, interceptors...)
}

func (c *sdkClient) PolicySdpStaticIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
