package tsrule

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
	PolicyTerminalServerAddRule(
		ctx context.Context,
		input cato_models.TerminalServerAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerAddRule, error)
	PolicyTerminalServerMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerMoveRule, error)
	PolicyTerminalServerRemoveRule(
		ctx context.Context,
		input cato_models.TerminalServerRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerRemoveRule, error)
	PolicyTerminalServerUpdateRule(
		ctx context.Context,
		input cato_models.TerminalServerUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerUpdateRule, error)
	PolicyTerminalServerPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error)
}

func (r *terminalServerRuleResource) resourceAPI() ResourceClient {
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
	return c.resource.client.Catov2.PolicyTerminalServerPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerAddRule(
	ctx context.Context,
	input cato_models.TerminalServerAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerAddRule, error) {
	return c.resource.client.Catov2.PolicyTerminalServerAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerMoveRule, error) {
	return c.resource.client.Catov2.PolicyTerminalServerMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerRemoveRule(
	ctx context.Context,
	input cato_models.TerminalServerRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerRemoveRule, error) {
	return c.resource.client.Catov2.PolicyTerminalServerRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerUpdateRule(
	ctx context.Context,
	input cato_models.TerminalServerUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerUpdateRule, error) {
	return c.resource.client.Catov2.PolicyTerminalServerUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *terminalServerRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyTerminalServerPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPublishPolicyRevision(ctx, accountID, interceptors...)
}
