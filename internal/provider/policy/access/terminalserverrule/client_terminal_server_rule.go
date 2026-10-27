package terminalserverrule

import (
	"context"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTerminalServerCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerCreatePolicyRevision, error)
	PolicyTerminalServerAddRule(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		terminalServerAddRuleInput cato_models.TerminalServerAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerAddRule, error)
	PolicyTerminalServerPolicy(
		ctx context.Context,
		accountID string,
		terminalServerPolicyInput *cato_models.TerminalServerPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerPolicy, error)
	PolicyTerminalServerUpdateRule(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		terminalServerUpdateRuleInput cato_models.TerminalServerUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerUpdateRule, error)
	PolicyTerminalServerMoveRule(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerMoveRule, error)
	PolicyTerminalServerRemoveRule(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		terminalServerRemoveRuleInput cato_models.TerminalServerRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerRemoveRule, error)
	PolicyTerminalServerPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerPublishPolicyRevision, error)
	PolicyTerminalServerDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyTerminalServerCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerCreatePolicyRevision, error) {
	return c.client.Catov2.PolicyTerminalServerCreatePolicyRevision(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerAddRule(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	terminalServerAddRuleInput cato_models.TerminalServerAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerAddRule, error) {
	return c.client.Catov2.PolicyTerminalServerAddRule(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		terminalServerAddRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerPolicy(
	ctx context.Context,
	accountID string,
	terminalServerPolicyInput *cato_models.TerminalServerPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerPolicy, error) {
	return c.client.Catov2.PolicyTerminalServerPolicy(
		ctx,
		accountID,
		terminalServerPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerUpdateRule(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	terminalServerUpdateRuleInput cato_models.TerminalServerUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerUpdateRule, error) {
	return c.client.Catov2.PolicyTerminalServerUpdateRule(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		terminalServerUpdateRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerMoveRule(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerMoveRule, error) {
	return c.client.Catov2.PolicyTerminalServerMoveRule(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyMoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerRemoveRule(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	terminalServerRemoveRuleInput cato_models.TerminalServerRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerRemoveRule, error) {
	return c.client.Catov2.PolicyTerminalServerRemoveRule(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		terminalServerRemoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerPublishPolicyRevision, error) {
	return c.client.Catov2.PolicyTerminalServerPublishPolicyRevision(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyPublishRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerDiscardPolicyRevision, error) {
	return c.client.Catov2.PolicyTerminalServerDiscardPolicyRevision(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy

type objectResponse = cato.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	var input *cato_models.TerminalServerPolicyInput
	if revision != "" {
		input = &cato_models.TerminalServerPolicyInput{Revision: &cato_models.PolicyRevisionInput{ID: &revision}}
	}
	result, err := r.getClient().PolicyTerminalServerPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		return nil, err
	}
	policy := result.GetPolicy().GetTerminalServer().GetPolicy()
	if policy == nil || policy.GetRules() == nil {
		return nil, fmt.Errorf("API did not return a policy with a complete rules list")
	}
	for _, entry := range policy.GetRules() {
		if entry.GetRule().GetID() == "" {
			return nil, fmt.Errorf("API returned a rule without an ID")
		}
	}
	return policy, nil
}

func findObject(policy *policyResponse, id string) *objectResponse {
	for _, entry := range policy.GetRules() {
		object := entry.GetRule()
		if object.GetID() == id {
			return object
		}
	}
	return nil
}
