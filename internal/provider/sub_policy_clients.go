package provider

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// InternetFirewallSubPolicyClient is the narrow SDK surface used by the
// cato_if_sub_policy resource. It exists so the resource can be unit tested with
// generated mocks.
type InternetFirewallSubPolicyClient interface {
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallAddSubPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallAddSubPolicyInput cato_models.InternetFirewallAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSubPolicy, error)
	PolicyInternetFirewallRemoveSubPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallRemoveSubPolicyInput cato_models.InternetFirewallRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy, error)
	PolicyInternetFirewallUpdateRule(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		internetFirewallUpdateRuleInput cato_models.InternetFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallUpdateRule, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
}

// WanFirewallSubPolicyClient is the narrow SDK surface used by the
// cato_wf_sub_policy resource.
type WanFirewallSubPolicyClient interface {
	PolicyWanFirewall(
		ctx context.Context,
		wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyWanFirewallAddSubPolicy(
		ctx context.Context,
		wanFirewallAddSubPolicyInput cato_models.WanFirewallAddSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallAddSubPolicy, error)
	PolicyWanFirewallRemoveSubPolicy(
		ctx context.Context,
		wanFirewallRemoveSubPolicyInput cato_models.WanFirewallRemoveSubPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallRemoveSubPolicy, error)
	PolicyWanFirewallUpdateRule(
		ctx context.Context,
		wanFirewallUpdateRuleInput cato_models.WanFirewallUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallUpdateRule, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
}

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
