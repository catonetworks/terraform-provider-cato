package alwaysonpolicy

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains the read, update, and publish operations used by the policy resource.
type ResourceClient interface {
	PolicyZtnaAlwaysOnPolicy(ctx context.Context, accountID string, input *cato_models.ZtnaAlwaysOnPolicyInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnUpdatePolicy(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnPolicyUpdateInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnUpdatePolicy, error)
	PolicyZtnaAlwaysOnPublishPolicyRevision(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnPublishPolicyRevision, error)
}
