package lfsubpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// SocketLanSubPolicyClient is the narrow SDK surface used by the
// cato_lf_sub_policy resource.
type SocketLanSubPolicyClient interface {
	PolicySocketLanPolicy(
		ctx context.Context,
		accountID string,
		socketLanPolicyInput *cato_models.SocketLanPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPolicy, error)
	PolicySocketLanAddSubPolicy(
		ctx context.Context,
		input cato_models.SocketLanAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanAddSubPolicy, error)
	PolicySocketLanUpdateRule(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		socketLanUpdateRuleInput cato_models.SocketLanUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanUpdateRule, error)
	PolicySocketLanRemoveSubPolicy(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		socketLanRemoveSubPolicyInput cato_models.SocketLanRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanRemoveSubPolicy, error)
	PolicySocketLanMoveRule(
		ctx context.Context,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanMoveRule, error)
	PolicySocketLanPublishPolicyRevision(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error)
}

func (r *lfSubPolicyResource) getClient() SocketLanSubPolicyClient {
	if r.subPolicyClient != nil {
		return r.subPolicyClient
	}
	return r.client.Catov2
}

type fromtoer interface {
	GetFrom() string
	GetTo() string
}

// publish calls the API to publish the draft policy revision
func (r *lfSubPolicyResource) publish(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "failed to publish LAN firewall policy"
	const notFound = "PolicyRevisionNotFound"
	result, err := r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, nil, r.client.AccountId)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	publish := result.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
	apiErrors := publish.GetErrors()
	if len(apiErrors) > 0 {
		allNotFound := true
		for _, apiError := range apiErrors {
			if code := apiError.GetErrorCode(); code == nil || *code != notFound {
				allNotFound = false
				break
			}
		}
		if allNotFound {
			return
		}
	}
	if utils.CheckAPIErrors(nil, apiErrors, summary, diags) {
		return
	}
	checkPolicyMutationStatus(publish.GetStatus(), summary, diags)
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySocketLanAddSubPolicy(
		ctx context.Context,
		input cato_models.SocketLanAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanAddSubPolicy, error)
	PolicySocketLanMoveRule(
		ctx context.Context,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanMoveRule, error)
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
	PolicySocketLanRemoveSubPolicy(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		socketLanRemoveSubPolicyInput cato_models.SocketLanRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanRemoveSubPolicy, error)
	PolicySocketLanUpdateRule(
		ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		socketLanUpdateRuleInput cato_models.SocketLanUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySocketLanUpdateRule, error)
}

func (r *lfSubPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicySocketLanAddSubPolicy(
	ctx context.Context,
	input cato_models.SocketLanAddSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanAddSubPolicy, error) {
	return c.resource.getClient().PolicySocketLanAddSubPolicy(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicySocketLanMoveRule(
	ctx context.Context,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanMoveRule, error) {
	return c.resource.getClient().PolicySocketLanMoveRule(ctx, policyMoveRuleInput, accountID, interceptors...)
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

func (c *sdkClient) PolicySocketLanRemoveSubPolicy(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	socketLanRemoveSubPolicyInput cato_models.SocketLanRemoveSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanRemoveSubPolicy, error) {
	return c.resource.getClient().PolicySocketLanRemoveSubPolicy(
		ctx,
		socketLanPolicyMutationInput,
		socketLanRemoveSubPolicyInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketLanUpdateRule(
	ctx context.Context,
	socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
	socketLanUpdateRuleInput cato_models.SocketLanUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanUpdateRule, error) {
	return c.resource.getClient().PolicySocketLanUpdateRule(
		ctx,
		socketLanPolicyMutationInput,
		socketLanUpdateRuleInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *lfSubPolicyResource }

var _ ResourceClient = (*sdkClient)(nil)
