package alwaysonrule

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// Client contains the SDK operations used by the Always On rule resource.
type Client interface {
	PolicyZtnaAlwaysOnAddRule(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnAddRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnAddRule, error)
}

var _ Client = (*cato.Client)(nil)
