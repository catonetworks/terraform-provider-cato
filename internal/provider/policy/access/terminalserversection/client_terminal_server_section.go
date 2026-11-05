package terminalserversection

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
	PolicyTerminalServerCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerCreatePolicyRevision, error)
	PolicyTerminalServerAddSection(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerAddSection, error)
	PolicyTerminalServerPolicy(
		ctx context.Context,
		accountID string,
		terminalServerPolicyInput *cato_models.TerminalServerPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerPolicy, error)
	PolicyTerminalServerUpdateSection(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerUpdateSection, error)
	PolicyTerminalServerMoveSection(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerMoveSection, error)
	PolicyTerminalServerRemoveSection(
		ctx context.Context,
		accountID string,
		terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyTerminalServerRemoveSection, error)
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

func (c *sdkClient) PolicyTerminalServerAddSection(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerAddSection, error) {
	return c.client.Catov2.PolicyTerminalServerAddSection(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyAddSectionInput,
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

func (c *sdkClient) PolicyTerminalServerUpdateSection(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerUpdateSection, error) {
	return c.client.Catov2.PolicyTerminalServerUpdateSection(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyUpdateSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerMoveSection(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerMoveSection, error) {
	return c.client.Catov2.PolicyTerminalServerMoveSection(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyMoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyTerminalServerRemoveSection(
	ctx context.Context,
	accountID string,
	terminalServerPolicyMutationInput *cato_models.TerminalServerPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTerminalServerRemoveSection, error) {
	return c.client.Catov2.PolicyTerminalServerRemoveSection(
		ctx,
		accountID,
		terminalServerPolicyMutationInput,
		policyRemoveSectionInput,
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

type objectResponse = cato.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Sections_Section

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.TerminalServerPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicyTerminalServerPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if policyshared.RevisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetTerminalServer().GetPolicy()
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
