package lanindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
)

type LanFwRuleClient interface {
	PolicySocketLanPolicy(ctx context.Context, accountID string, socketLanPolicyInput *cato_models.SocketLanPolicyInput,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPolicy, error)
	PolicySocketLanMoveSection(ctx context.Context, policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveSection, error)
	PolicySocketLanMoveRule(ctx context.Context, policyMoveRuleInput cato_models.PolicyMoveRuleInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveRule, error)
	PolicySocketLanFirewallMoveRule(ctx context.Context, accountID string,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyMoveSubRuleInput cato_models.PolicyMoveSubRuleInput, interceptors ...clientv2.RequestInterceptor) (
		*cato_go_sdk.PolicySocketLanFirewallMoveRule, error)
	PolicySocketLanPublishPolicyRevision(ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error)
}

func (r *lanRulesIndexResource) getClient() LanFwRuleClient { return r.catov2Client }

// publish calls the API to publish the draft policy revision
func (r *lanRulesIndexResource) publish(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "failed to publish LAN firewall policy"
	const notFound = "PolicyRevisionNotFound"
	result, err := r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, nil, r.client.AccountId)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	publish := result.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
	errors := publish.GetErrors()
	if len(errors) > 0 {
		hasAPIError := false
		for _, e := range errors {
			if code := e.GetErrorCode(); code != nil && *code == notFound {
				continue
			}
			hasAPIError = true
			if msg := e.GetErrorMessage(); msg != nil {
				diags.AddError(summary, *msg)
			} else {
				diags.AddError(summary, "unknown error")
			}
		}
		if hasAPIError {
			return
		}
	}
	if err := lanPolicyMutationStatusError(publish.GetStatus()); err != nil {
		diags.AddError(summary, err.Error())
	}
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketLanFirewallMoveRule(
		ctx context.Context,
		accountID string,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyMoveSubRuleInput cato_models.PolicyMoveSubRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanFirewallMoveRule, error)
	PolicySocketLanMoveRule(
		ctx context.Context,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanMoveRule, error)
	PolicySocketLanMoveSection(
		ctx context.Context,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanMoveSection, error)
	PolicySocketLanPolicy(
		ctx context.Context,
		accountID string,
		socketLanPolicyInput *cato_models.SocketLanPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPolicy, error)
	PolicySocketLanPublishPolicyRevision(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error)
}

func (r *lanRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketLanFirewallMoveRule(
	ctx context.Context,
	accountID string,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	policyMoveSubRuleInput cato_models.PolicyMoveSubRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanFirewallMoveRule, error) {
	return c.resource.getClient().PolicySocketLanFirewallMoveRule(
		ctx,
		accountID,
		socketLanPolicyMutationInput,
		policyMoveSubRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketLanMoveRule(
	ctx context.Context,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanMoveRule, error) {
	return c.resource.getClient().PolicySocketLanMoveRule(ctx, policyMoveRuleInput, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketLanMoveSection(
	ctx context.Context,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanMoveSection, error) {
	return c.resource.getClient().PolicySocketLanMoveSection(ctx, policyMoveSectionInput, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketLanPolicy(
	ctx context.Context,
	accountID string,
	socketLanPolicyInput *cato_models.SocketLanPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPolicy, error) {
	return c.resource.getClient().PolicySocketLanPolicy(ctx, accountID, socketLanPolicyInput, interceptors...)
}

func (c *sdkClient) PolicySocketLanPublishPolicyRevision(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error) {
	return c.resource.getClient().PolicySocketLanPublishPolicyRevision(
		ctx,
		socketLanPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *lanRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)
