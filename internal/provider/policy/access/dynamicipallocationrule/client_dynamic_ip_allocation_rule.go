package dynamicipallocationrule

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
	PolicyDynamicIPAllocationCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationCreatePolicyRevision, error)
	PolicyDynamicIPAllocationAddRule(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		dynamicIPAllocationAddRuleInput cato_models.DynamicIPAllocationAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationAddRule, error)
	PolicyDynamicIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyInput *cato_models.DynamicIPAllocationPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationPolicy, error)
	PolicyDynamicIPAllocationUpdateRule(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		dynamicIPAllocationUpdateRuleInput cato_models.DynamicIPAllocationUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationUpdateRule, error)
	PolicyDynamicIPAllocationMoveRule(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationMoveRule, error)
	PolicyDynamicIPAllocationRemoveRule(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		dynamicIPAllocationRemoveRuleInput cato_models.DynamicIPAllocationRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationRemoveRule, error)
	PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationPublishPolicyRevision, error)
	PolicyDynamicIPAllocationDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyDynamicIPAllocationCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationCreatePolicyRevision, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationCreatePolicyRevision(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationAddRule(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	dynamicIPAllocationAddRuleInput cato_models.DynamicIPAllocationAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationAddRule, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationAddRule(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		dynamicIPAllocationAddRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyInput *cato_models.DynamicIPAllocationPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationPolicy, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationPolicy(
		ctx,
		accountID,
		dynamicIPAllocationPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationUpdateRule(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	dynamicIPAllocationUpdateRuleInput cato_models.DynamicIPAllocationUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationUpdateRule, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationUpdateRule(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		dynamicIPAllocationUpdateRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationMoveRule(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationMoveRule, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationMoveRule(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyMoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationRemoveRule(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	dynamicIPAllocationRemoveRuleInput cato_models.DynamicIPAllocationRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationRemoveRule, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationRemoveRule(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		dynamicIPAllocationRemoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationPublishPolicyRevision, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyPublishRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationDiscardPolicyRevision, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationDiscardPolicyRevision(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy

type objectResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.DynamicIPAllocationPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicyDynamicIPAllocationPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetDynamicIPAllocation().GetPolicy()
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
