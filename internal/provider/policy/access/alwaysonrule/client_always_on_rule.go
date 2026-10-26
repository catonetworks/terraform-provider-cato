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

func (r *alwaysOnRuleResource) configuredClient() ResourceClient {
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{resource: r}
}

type sdkClient struct{ resource *alwaysOnRuleResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyZtnaAlwaysOnAddRule(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnAddRule, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnAddRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnPolicy(ctx context.Context, accountID string,
	input *cato_models.ZtnaAlwaysOnPolicyInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnPolicy, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnUpdateRule(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnUpdateRule, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnUpdateRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnMoveRule(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnMoveRule, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnMoveRule(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnRemoveRule(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.ZtnaAlwaysOnRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnRemoveRule, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnRemoveRule(ctx, accountID, policyInput, input, interceptors...)
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
