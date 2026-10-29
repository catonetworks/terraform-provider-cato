package sdpstaticipallocationrule

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
	PolicySdpStaticIPAllocationCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationCreatePolicyRevision, error)
	PolicySdpStaticIPAllocationAddRule(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		sdpStaticIPAllocationAddRuleInput cato_models.SdpStaticIPAllocationAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationAddRule, error)
	PolicySdpStaticIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyInput *cato_models.SdpStaticIPAllocationPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationPolicy, error)
	PolicySdpStaticIPAllocationUpdateRule(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		sdpStaticIPAllocationUpdateRuleInput cato_models.SdpStaticIPAllocationUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationUpdateRule, error)
	PolicySdpStaticIPAllocationMoveRule(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationMoveRule, error)
	PolicySdpStaticIPAllocationRemoveRule(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		sdpStaticIPAllocationRemoveRuleInput cato_models.SdpStaticIPAllocationRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationRemoveRule, error)
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
	PolicySdpStaticIPAllocationDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySdpStaticIPAllocationDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySdpStaticIPAllocationCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationCreatePolicyRevision, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationCreatePolicyRevision(
		ctx,
		accountID,
		policyCreateRevisionInput,
		sdpStaticIPAllocationPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationAddRule(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	sdpStaticIPAllocationAddRuleInput cato_models.SdpStaticIPAllocationAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationAddRule, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationAddRule(
		ctx,
		accountID,
		sdpStaticIPAllocationAddRuleInput,
		sdpStaticIPAllocationPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationPolicy(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyInput *cato_models.SdpStaticIPAllocationPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationPolicy, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationPolicy(
		ctx,
		accountID,
		sdpStaticIPAllocationPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationUpdateRule(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	sdpStaticIPAllocationUpdateRuleInput cato_models.SdpStaticIPAllocationUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationUpdateRule, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationUpdateRule(
		ctx,
		accountID,
		sdpStaticIPAllocationPolicyMutationInput,
		sdpStaticIPAllocationUpdateRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationMoveRule(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationMoveRule, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationMoveRule(
		ctx,
		accountID,
		policyMoveRuleInput,
		sdpStaticIPAllocationPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationRemoveRule(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	sdpStaticIPAllocationRemoveRuleInput cato_models.SdpStaticIPAllocationRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationRemoveRule, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationRemoveRule(
		ctx,
		accountID,
		sdpStaticIPAllocationPolicyMutationInput,
		sdpStaticIPAllocationRemoveRuleInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationPublishPolicyRevision, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx,
		accountID,
		policyPublishRevisionInput,
		sdpStaticIPAllocationPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySdpStaticIPAllocationDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	sdpStaticIPAllocationPolicyMutationInput *cato_models.SdpStaticIPAllocationPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySdpStaticIPAllocationDiscardPolicyRevision, error) {
	return c.client.Catov2.PolicySdpStaticIPAllocationDiscardPolicyRevision(
		ctx,
		accountID,
		policyDiscardRevisionInput,
		sdpStaticIPAllocationPolicyMutationInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy

type objectResponse = cato.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Rules_Rule

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.SdpStaticIPAllocationPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicySdpStaticIPAllocationPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetSdpStaticIPAllocation().GetPolicy()
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
