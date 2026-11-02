package alwaysonrule

import (
	"context"
	"errors"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains the SDK operations used by the Always On rule resource.
type ResourceClient interface {
	PolicyZtnaAlwaysOnAddRule(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnAddRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnAddRule, error)
	PolicyZtnaAlwaysOnPolicy(ctx context.Context, accountID string, input *cato_models.ZtnaAlwaysOnPolicyInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnUpdateRule(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnUpdateRule, error)
	PolicyZtnaAlwaysOnMoveRule(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnMoveRule, error)
	PolicyZtnaAlwaysOnRemoveRule(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnRemoveRule, error)
}

func findAlwaysOnRule(result *cato.PolicyZtnaAlwaysOnPolicy, id string,
) (*cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy_Rules_Rule, error) {
	policy := result.GetPolicy().GetZtnaAlwaysOn()
	if policy == nil || policy.GetPolicy().GetRules() == nil {
		return nil, errors.New("API did not return an Always On policy with a rule list")
	}
	var found *cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy_Rules_Rule
	for _, entry := range policy.GetPolicy().GetRules() {
		rule := entry.GetRule()
		if rule.GetID() == "" {
			return nil, errors.New("API returned an Always On rule without an ID")
		}
		if rule.GetID() == id {
			found = rule
		}
	}
	return found, nil
}
