package clientconnectivityrule

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
	PolicyClientConnectivityCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityCreatePolicyRevision, error)
	PolicyClientConnectivityAddRule(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		clientConnectivityAddRuleInput cato_models.ClientConnectivityAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityAddRule, error)
	PolicyClientConnectivityPolicy(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyInput *cato_models.ClientConnectivityPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityPolicy, error)
	PolicyClientConnectivityUpdateRule(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		clientConnectivityUpdateRuleInput cato_models.ClientConnectivityUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityUpdateRule, error)
	PolicyClientConnectivityMoveRule(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityMoveRule, error)
	PolicyClientConnectivityRemoveRule(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		clientConnectivityRemoveRuleInput cato_models.ClientConnectivityRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityRemoveRule, error)
	PolicyClientConnectivityPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityPublishPolicyRevision, error)
	PolicyClientConnectivityDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyClientConnectivityCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityCreatePolicyRevision, error) {
	return c.client.Catov2.PolicyClientConnectivityCreatePolicyRevision(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityAddRule(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	clientConnectivityAddRuleInput cato_models.ClientConnectivityAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityAddRule, error) {
	return c.client.Catov2.PolicyClientConnectivityAddRule(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		clientConnectivityAddRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityPolicy(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyInput *cato_models.ClientConnectivityPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityPolicy, error) {
	return c.client.Catov2.PolicyClientConnectivityPolicy(
		ctx,
		accountID,
		clientConnectivityPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityUpdateRule(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	clientConnectivityUpdateRuleInput cato_models.ClientConnectivityUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityUpdateRule, error) {
	return c.client.Catov2.PolicyClientConnectivityUpdateRule(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		clientConnectivityUpdateRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityMoveRule(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityMoveRule, error) {
	return c.client.Catov2.PolicyClientConnectivityMoveRule(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyMoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityRemoveRule(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	clientConnectivityRemoveRuleInput cato_models.ClientConnectivityRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityRemoveRule, error) {
	return c.client.Catov2.PolicyClientConnectivityRemoveRule(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		clientConnectivityRemoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityPublishPolicyRevision, error) {
	return c.client.Catov2.PolicyClientConnectivityPublishPolicyRevision(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyPublishRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityDiscardPolicyRevision, error) {
	return c.client.Catov2.PolicyClientConnectivityDiscardPolicyRevision(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy

type objectResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.ClientConnectivityPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicyClientConnectivityPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetClientConnectivity().GetPolicy()
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
