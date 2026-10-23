package ifsection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallAddSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSection, error)
	PolicyInternetFirewallMoveSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallMoveSection, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
	PolicyInternetFirewallRemoveSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallRemoveSection, error)
	PolicyInternetFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwSectionsIndexPolicy, error)
	PolicyInternetFirewallUpdateSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdateSection, error)
}

func (r *internetFwSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyInternetFirewall(
	ctx context.Context,
	internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Policy, error) {
	return c.resource.client.Catov2.PolicyInternetFirewall(ctx, internetFirewallPolicyInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallAddSection(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallAddSection, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallAddSection(
		ctx,
		internetFirewallPolicyMutationInput,
		policyAddSectionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallMoveSection(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallMoveSection, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallMoveSection(
		ctx,
		internetFirewallPolicyMutationInput,
		policyMoveSectionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallPublishPolicyRevision(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		internetFirewallPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallRemoveSection(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallRemoveSection, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallRemoveSection(
		ctx,
		internetFirewallPolicyMutationInput,
		policyRemoveSectionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallSectionsIndex(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.IfwSectionsIndexPolicy, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallSectionsIndex(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallUpdateSection(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallUpdateSection, error) {
	return c.resource.client.Catov2.PolicyInternetFirewallUpdateSection(
		ctx,
		internetFirewallPolicyMutationInput,
		policyUpdateSectionInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *internetFwSectionResource }

var _ ResourceClient = (*sdkClient)(nil)
