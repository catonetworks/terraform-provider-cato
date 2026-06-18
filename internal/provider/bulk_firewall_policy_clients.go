package provider

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// InternetFirewallBulkPolicyClient is the minimal Cato client surface used by
// cato_bulk_if_move_rule (mocked in unit tests via mockery).
type InternetFirewallBulkPolicyClient interface {
	PolicyInternetFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwSectionsIndexPolicy, error)
	PolicyInternetFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwRulesIndexPolicy, error)
	PolicyInternetFirewallMoveSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallMoveSection, error)
	PolicyInternetFirewallAddSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSection, error)
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

// WanFirewallBulkPolicyClient is the minimal Cato client surface used by
// cato_bulk_wf_move_rule (mocked in unit tests via mockery).
type WanFirewallBulkPolicyClient interface {
	PolicyWanFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanSectionsIndexPolicy, error)
	PolicyWanFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanRulesIndexPolicy, error)
	PolicyWanFirewallMoveSection(
		ctx context.Context,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallMoveSection, error)
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
}
