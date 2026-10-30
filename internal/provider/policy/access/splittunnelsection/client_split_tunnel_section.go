package splittunnelsection

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
	PolicySplitTunnelCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelCreatePolicyRevision, error)
	PolicySplitTunnelAddSection(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelAddSection, error)
	PolicySplitTunnelPolicy(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyInput *cato_models.SplitTunnelPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelPolicy, error)
	PolicySplitTunnelUpdateSection(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelUpdateSection, error)
	PolicySplitTunnelMoveSection(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelMoveSection, error)
	PolicySplitTunnelRemoveSection(
		ctx context.Context,
		accountID string,
		splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySplitTunnelRemoveSection, error)
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

func (c *sdkClient) PolicySplitTunnelAddSection(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelAddSection, error) {
	return c.client.Catov2.PolicySplitTunnelAddSection(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyAddSectionInput,
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

func (c *sdkClient) PolicySplitTunnelUpdateSection(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelUpdateSection, error) {
	return c.client.Catov2.PolicySplitTunnelUpdateSection(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyUpdateSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelMoveSection(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelMoveSection, error) {
	return c.client.Catov2.PolicySplitTunnelMoveSection(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyMoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySplitTunnelRemoveSection(
	ctx context.Context,
	accountID string,
	splitTunnelPolicyMutationInput *cato_models.SplitTunnelPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySplitTunnelRemoveSection, error) {
	return c.client.Catov2.PolicySplitTunnelRemoveSection(
		ctx,
		accountID,
		splitTunnelPolicyMutationInput,
		policyRemoveSectionInput,
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

type objectResponse = cato.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Sections_Section

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.SplitTunnelPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicySplitTunnelPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if policyshared.RevisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetSplitTunnel().GetPolicy()
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
