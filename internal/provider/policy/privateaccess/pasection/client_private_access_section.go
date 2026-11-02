package pasection

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
	PolicyPrivateAccessCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessCreatePolicyRevision, error)
	PolicyPrivateAccessAddSection(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessAddSection, error)
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		privateAccessPolicyInput *cato_models.PrivateAccessPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyReadPrivateAccessPolicy, error)
	PolicyPrivateAccessUpdateSection(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessUpdateSection, error)
	PolicyPrivateAccessMoveSection(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessMoveSection, error)
	PolicyPrivateAccessRemoveSection(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessRemoveSection, error)
	PolicyPrivateAccessPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessPublishPolicyRevision, error)
	PolicyPrivateAccessDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyPrivateAccessDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyPrivateAccessCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessCreatePolicyRevision, error) {
	return c.client.V2().PolicyPrivateAccessCreatePolicyRevision(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyCreateRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyPrivateAccessAddSection(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessAddSection, error) {
	return c.client.V2().PolicyPrivateAccessAddSection(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyAddSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	privateAccessPolicyInput *cato_models.PrivateAccessPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyReadPrivateAccessPolicy, error) {
	return c.client.V2().PolicyReadPrivateAccessPolicy(
		ctx,
		accountID,
		privateAccessPolicyInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyPrivateAccessUpdateSection(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessUpdateSection, error) {
	return c.client.V2().PolicyPrivateAccessUpdateSection(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyUpdateSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyPrivateAccessMoveSection(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessMoveSection, error) {
	return c.client.V2().PolicyPrivateAccessMoveSection(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyMoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyPrivateAccessRemoveSection(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessRemoveSection, error) {
	return c.client.V2().PolicyPrivateAccessRemoveSection(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyRemoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyPrivateAccessPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessPublishPolicyRevision, error) {
	return c.client.V2().PolicyPrivateAccessPublishPolicyRevision(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyPublishRevisionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyPrivateAccessDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	privateAccessPolicyMutationInput *cato_models.PrivateAccessPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessDiscardPolicyRevision, error) {
	return c.client.V2().PolicyPrivateAccessDiscardPolicyRevision(
		ctx,
		accountID,
		privateAccessPolicyMutationInput,
		policyDiscardRevisionInput,
		interceptors...,
	)
}

type policyResponse = cato.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy

type objectResponse = cato.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Sections_Section

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.PrivateAccessPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if policyshared.RevisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
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
