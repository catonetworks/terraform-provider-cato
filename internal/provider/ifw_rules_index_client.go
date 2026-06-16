package provider

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type IfwRulesIndexClient interface {
	PolicyInternetFirewall(
		ctx context.Context,
		internetFirewallPolicyInput *cato_models.InternetFirewallPolicyInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Policy, error)
	PolicyInternetFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwSectionsIndexPolicy, error)
	PolicyInternetFirewallCreatePolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision, error)
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
	PolicyInternetFirewallReorderPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyReorderInput cato_models.PolicyReorderInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallReorderPolicy, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
}
