package alwaysonindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyZtnaAlwaysOnPolicy(
		ctx context.Context,
		accountID string,
		input *cato_models.ZtnaAlwaysOnPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnMoveSection, error)
	PolicyZtnaAlwaysOnMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnMoveRule, error)
	PolicyZtnaAlwaysOnPublishPolicyRevision(
		ctx context.Context,
		input *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnPublishPolicyRevision, error)
}

func (r *aoRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct{ resource *aoRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyZtnaAlwaysOnPolicy(
	ctx context.Context,
	accountID string,
	input *cato_models.ZtnaAlwaysOnPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnPolicy, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnMoveSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnMoveRule, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnMoveRule(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnPublishPolicyRevision(
	ctx context.Context,
	input *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, accountID, nil, input, interceptors...)
}
