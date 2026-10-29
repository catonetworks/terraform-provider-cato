package sbrule

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
	PolicySocketBypassCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassCreatePolicyRevision, error)
	PolicySocketBypassAddRule(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		socketBypassAddRuleInput cato_models.SocketBypassAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassAddRule, error)
	PolicySocketBypassPolicy(
		ctx context.Context,
		accountID string,
		socketBypassPolicyInput *cato_models.SocketBypassPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassPolicy, error)
	PolicySocketBypassUpdateRule(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		socketBypassUpdateRuleInput cato_models.SocketBypassUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassUpdateRule, error)
	PolicySocketBypassMoveRule(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassMoveRule, error)
	PolicySocketBypassRemoveRule(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		socketBypassRemoveRuleInput cato_models.SocketBypassRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassRemoveRule, error)
	PolicySocketBypassPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassPublishPolicyRevision, error)
	PolicySocketBypassDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySocketBypassCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassCreatePolicyRevision, error) {
	return c.client.Catov2.PolicySocketBypassCreatePolicyRevision(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassAddRule(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	socketBypassAddRuleInput cato_models.SocketBypassAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassAddRule, error) {
	return c.client.Catov2.PolicySocketBypassAddRule(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		socketBypassAddRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassPolicy(
	ctx context.Context,
	accountID string,
	socketBypassPolicyInput *cato_models.SocketBypassPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassPolicy, error) {
	return c.client.Catov2.PolicySocketBypassPolicy(
		ctx,
		accountID,
		socketBypassPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassUpdateRule(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	socketBypassUpdateRuleInput cato_models.SocketBypassUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassUpdateRule, error) {
	return c.client.Catov2.PolicySocketBypassUpdateRule(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		socketBypassUpdateRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassMoveRule(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassMoveRule, error) {
	return c.client.Catov2.PolicySocketBypassMoveRule(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyMoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassRemoveRule(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	socketBypassRemoveRuleInput cato_models.SocketBypassRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassRemoveRule, error) {
	return c.client.Catov2.PolicySocketBypassRemoveRule(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		socketBypassRemoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassPublishPolicyRevision, error) {
	return c.client.Catov2.PolicySocketBypassPublishPolicyRevision(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyPublishRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassDiscardPolicyRevision, error) {
	return c.client.Catov2.PolicySocketBypassDiscardPolicyRevision(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy

type objectResponse = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.SocketBypassPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicySocketBypassPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetSocketBypass().GetPolicy()
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
