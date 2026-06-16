package provider

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type WanRulesIndexClient interface {
	PolicyWanFirewall(
		ctx context.Context,
		wanFirewallPolicyInput *cato_models.WanFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyWanFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanSectionsIndexPolicy, error)
	PolicyWanFirewallCreatePolicyRevision(
		ctx context.Context,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallCreatePolicyRevision, error)
	PolicyWanFirewallMoveSection(
		ctx context.Context,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallMoveSection, error)
	PolicyWanFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanRulesIndexPolicy, error)
	PolicyWanFirewallReorderPolicy(
		ctx context.Context,
		wanFirewallPolicyMutationInput *cato_models.WanFirewallPolicyMutationInput,
		policyReorderInput cato_models.PolicyReorderInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallReorderPolicy, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
	PolicyWanFirewallDiscardPolicyRevision(
		ctx context.Context,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallDiscardPolicyRevision, error)
}
