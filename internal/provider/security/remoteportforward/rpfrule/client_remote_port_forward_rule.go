package rpfrule

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
	PolicyRemotePortFwdCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdCreatePolicyRevision, error)
	PolicyRemotePortFwdAddRule(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		remotePortFwdAddRuleInput cato_models.RemotePortFwdAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdAddRule, error)
	RemotePortFwdPolicy(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyInput *cato_models.RemotePortFwdPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.RemotePortFwdPolicy, error)
	PolicyRemotePortFwdUpdateRule(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		remotePortFwdUpdateRuleInput cato_models.RemotePortFwdUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdUpdateRule, error)
	PolicyRemotePortFwdMoveRule(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdMoveRule, error)
	PolicyRemotePortFwdRemoveRule(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		remotePortFwdRemoveRuleInput cato_models.RemotePortFwdRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdRemoveRule, error)
	PolicyRemotePortFwdPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdPublishPolicyRevision, error)
	PolicyRemotePortFwdDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyRemotePortFwdCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdCreatePolicyRevision, error) {
	return c.client.V2().PolicyRemotePortFwdCreatePolicyRevision(
		ctx,
		accountID,
		remotePortFwdPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdAddRule(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	remotePortFwdAddRuleInput cato_models.RemotePortFwdAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdAddRule, error) {
	return c.client.V2().PolicyRemotePortFwdAddRule(
		ctx,
		remotePortFwdAddRuleInput,
		accountID,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) RemotePortFwdPolicy(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyInput *cato_models.RemotePortFwdPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.RemotePortFwdPolicy, error) {
	return c.client.V2().RemotePortFwdPolicy(
		ctx,
		accountID,
		remotePortFwdPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdUpdateRule(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	remotePortFwdUpdateRuleInput cato_models.RemotePortFwdUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdUpdateRule, error) {
	return c.client.V2().PolicyRemotePortFwdUpdateRule(
		ctx,
		remotePortFwdUpdateRuleInput,
		accountID,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdMoveRule(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdMoveRule, error) {
	return c.client.V2().PolicyRemotePortFwdMoveRule(
		ctx,
		policyMoveRuleInput,
		accountID,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdRemoveRule(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	remotePortFwdRemoveRuleInput cato_models.RemotePortFwdRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdRemoveRule, error) {
	return c.client.V2().PolicyRemotePortFwdRemoveRule(
		ctx,
		remotePortFwdRemoveRuleInput,
		accountID,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdPublishPolicyRevision, error) {
	return c.client.V2().PolicyRemotePortFwdPublishPolicyRevision(
		ctx,
		accountID,
		policyPublishRevisionInput,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdDiscardPolicyRevision, error) {
	return c.client.V2().PolicyRemotePortFwdDiscardPolicyRevision(
		ctx,
		accountID,
		remotePortFwdPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy

type objectResponse = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.RemotePortFwdPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().RemotePortFwdPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetRemotePortFwd().GetPolicy()
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
