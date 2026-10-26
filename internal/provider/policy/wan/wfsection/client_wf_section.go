package wfsection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyWanFirewall(
		ctx context.Context,
		wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyWanFirewallAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallAddSection, error)
	PolicyWanFirewallMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallMoveSection, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		input *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
	PolicyWanFirewallRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallRemoveSection, error)
	PolicyWanFirewallUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallUpdateSection, error)
}

func (r *wanFwSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyWanFirewall(
	ctx context.Context,
	wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Policy, error) {
	return c.resource.client.Catov2.PolicyWanFirewall(ctx, wanFirewallPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallAddSection, error) {
	return c.resource.client.Catov2.PolicyWanFirewallAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallMoveSection, error) {
	return c.resource.client.Catov2.PolicyWanFirewallMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallPublishPolicyRevision(
	ctx context.Context,
	input *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyWanFirewallPublishPolicyRevision(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallRemoveSection, error) {
	return c.resource.client.Catov2.PolicyWanFirewallRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallUpdateSection, error) {
	return c.resource.client.Catov2.PolicyWanFirewallUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *wanFwSectionResource }

var _ ResourceClient = (*sdkClient)(nil)
