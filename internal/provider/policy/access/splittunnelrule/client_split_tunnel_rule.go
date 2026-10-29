package splittunnelrule

import (
	"context"
	"errors"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicySplitTunnelCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelCreatePolicyRevision, error)
	PolicySplitTunnelAddRule(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		splitTunnelAddRuleInput cato_models.SplitTunnelAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelAddRule, error)
	PolicySplitTunnelPolicy(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyInput *cato_models.SplitTunnelPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelPolicy, error)
	PolicySplitTunnelUpdateRule(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		splitTunnelUpdateRuleInput cato_models.SplitTunnelUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelUpdateRule, error)
	PolicySplitTunnelMoveRule(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelMoveRule, error)
	PolicySplitTunnelRemoveRule(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		splitTunnelRemoveRuleInput cato_models.SplitTunnelRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelRemoveRule, error)
	PolicySplitTunnelPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelPublishPolicyRevision, error)
	PolicySplitTunnelDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySplitTunnelCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelCreatePolicyRevision, error) {
	return c.client.Catov2.PolicySplitTunnelCreatePolicyRevision(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelAddRule(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	splitTunnelAddRuleInput cato_models.SplitTunnelAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelAddRule, error) {
	return c.client.Catov2.PolicySplitTunnelAddRule(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		splitTunnelAddRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelPolicy(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyInput *cato_models.SplitTunnelPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelPolicy, error) {
	return c.client.Catov2.PolicySplitTunnelPolicy(
		ctx,
		accountID,
		splitTunnelPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelUpdateRule(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	splitTunnelUpdateRuleInput cato_models.SplitTunnelUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelUpdateRule, error) {
	return c.client.Catov2.PolicySplitTunnelUpdateRule(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		splitTunnelUpdateRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelMoveRule(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelMoveRule, error) {
	return c.client.Catov2.PolicySplitTunnelMoveRule(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyMoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelRemoveRule(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	splitTunnelRemoveRuleInput cato_models.SplitTunnelRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelRemoveRule, error) {
	return c.client.Catov2.PolicySplitTunnelRemoveRule(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		splitTunnelRemoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelPublishPolicyRevision, error) {
	return c.client.Catov2.PolicySplitTunnelPublishPolicyRevision(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyPublishRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelDiscardPolicyRevision, error) {
	return c.client.Catov2.PolicySplitTunnelDiscardPolicyRevision(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy

type objectResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.SplitTunnelPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicySplitTunnelPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetSplitTunnel().GetPolicy()
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

var errMissingRevision = errors.New("policy revision no longer exists")

// Match the policy API's structured error, never unrelated transport failures.
func revisionMissing(err error) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.GqlErrors == nil {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item.Extensions["errorCode"] == "PolicyRevisionNotFound" || item.Message == "Revision was not found" {
			return true
		}
	}
	return false
}
