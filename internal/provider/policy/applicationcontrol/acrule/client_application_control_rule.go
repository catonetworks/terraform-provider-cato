package acrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	ApplicationControlPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ApplicationControlPolicy, error)
	PolicyApplicationControlAddRule(
		ctx context.Context,
		input cato_models.ApplicationControlAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlAddRule, error)
	PolicyApplicationControlMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlMoveRule, error)
	PolicyApplicationControlRemoveRule(
		ctx context.Context,
		input cato_models.ApplicationControlRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlRemoveRule, error)
	PolicyApplicationControlUpdateRule(
		ctx context.Context,
		input cato_models.ApplicationControlUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlUpdateRule, error)
	PolicyApplicationControlPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error)
}

func (r *applicationControlRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) ApplicationControlPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ApplicationControlPolicy, error) {
	return c.resource.client.Catov2.ApplicationControlPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlAddRule(
	ctx context.Context,
	input cato_models.ApplicationControlAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlAddRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlMoveRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlRemoveRule(
	ctx context.Context,
	input cato_models.ApplicationControlRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlRemoveRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlUpdateRule(
	ctx context.Context,
	input cato_models.ApplicationControlUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlUpdateRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *applicationControlRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyApplicationControlPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyApplicationControlPublishPolicyRevision(ctx, accountID, interceptors...)
}
