package dynamicipallocationsection

import (
	"context"
	"errors"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
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
	PolicyDynamicIPAllocationAddSection(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationAddSection, error)
	PolicyDynamicIPAllocationPolicy(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyInput *cato_models.DynamicIPAllocationPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationPolicy, error)
	PolicyDynamicIPAllocationUpdateSection(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationUpdateSection, error)
	PolicyDynamicIPAllocationMoveSection(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationMoveSection, error)
	PolicyDynamicIPAllocationRemoveSection(
		ctx context.Context,
		accountID string,
		dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyDynamicIPAllocationRemoveSection, error)
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

func (c *sdkClient) PolicyDynamicIPAllocationAddSection(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationAddSection, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationAddSection(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyAddSectionInput,
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

func (c *sdkClient) PolicyDynamicIPAllocationUpdateSection(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationUpdateSection, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationUpdateSection(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyUpdateSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationMoveSection(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationMoveSection, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationMoveSection(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyMoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyDynamicIPAllocationRemoveSection(
	ctx context.Context,
	accountID string,
	dynamicIPAllocationPolicyMutationInput *cato_models.DynamicIPAllocationPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyDynamicIPAllocationRemoveSection, error) {
	return c.client.Catov2.PolicyDynamicIPAllocationRemoveSection(
		ctx,
		accountID,
		dynamicIPAllocationPolicyMutationInput,
		policyRemoveSectionInput,
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

type objectResponse = cato.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Sections_Section

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.DynamicIPAllocationPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicyDynamicIPAllocationPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if policyshared.RevisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetDynamicIPAllocation().GetPolicy()
	if policy == nil || policy.GetSections() == nil || policy.GetRules() == nil {
		return nil, fmt.Errorf("API did not return a policy with a complete sections list")
	}
	for _, entry := range policy.GetSections() {
		if entry.GetSection().GetID() == "" {
			return nil, fmt.Errorf("API returned a section without an ID")
		}
	}
	return policy, nil
}

func findObject(policy *policyResponse, id string) *objectResponse {
	for _, entry := range policy.GetSections() {
		object := entry.GetSection()
		if object.GetID() == id {
			return object
		}
	}
	return nil
}

var errMissingRevision = errors.New("policy revision no longer exists")
