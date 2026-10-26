package parule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyPrivateAccessAddRule(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessAddRule, error)
	PolicyPrivateAccessDeleteRule(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessDeleteRule, error)
	PolicyPrivateAccessPublishRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessPublishRevision, error)
	PolicyPrivateAccessUpdateRule(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessUpdateRule, error)
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error)
}

func (r *privAccessRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyPrivateAccessAddRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessAddRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessAddRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessDeleteRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessDeleteRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessDeleteRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessPublishRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessPublishRevision, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessPublishRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessUpdateRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessUpdateRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessUpdateRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error) {
	return c.resource.client.Catov2.PolicyReadPrivateAccessPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *privAccessRuleResource }

var _ ResourceClient = (*sdkClient)(nil)
