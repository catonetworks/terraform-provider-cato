package lfsubpolicy

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
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
