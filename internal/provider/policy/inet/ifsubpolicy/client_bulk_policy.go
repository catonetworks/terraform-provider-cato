package ifsubpolicy

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
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
