package wfsubpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

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
