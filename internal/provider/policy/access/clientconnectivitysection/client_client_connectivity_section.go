package clientconnectivitysection

import (
	"context"
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
	PolicyClientConnectivityAddSection(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityAddSection, error)
	PolicyClientConnectivityPolicy(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyInput *cato_models.ClientConnectivityPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityPolicy, error)
	PolicyClientConnectivityUpdateSection(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityUpdateSection, error)
	PolicyClientConnectivityMoveSection(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityMoveSection, error)
	PolicyClientConnectivityRemoveSection(
		ctx context.Context,
		accountID string,
		clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyClientConnectivityRemoveSection, error)
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

func (c *sdkClient) PolicyClientConnectivityAddSection(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityAddSection, error) {
	return c.client.Catov2.PolicyClientConnectivityAddSection(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyAddSectionInput,
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

func (c *sdkClient) PolicyClientConnectivityUpdateSection(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityUpdateSection, error) {
	return c.client.Catov2.PolicyClientConnectivityUpdateSection(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyUpdateSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityMoveSection(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityMoveSection, error) {
	return c.client.Catov2.PolicyClientConnectivityMoveSection(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyMoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyClientConnectivityRemoveSection(
	ctx context.Context,
	accountID string,
	clientConnectivityPolicyMutationInput *cato_models.ClientConnectivityPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyClientConnectivityRemoveSection, error) {
	return c.client.Catov2.PolicyClientConnectivityRemoveSection(
		ctx,
		accountID,
		clientConnectivityPolicyMutationInput,
		policyRemoveSectionInput,
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

type objectResponse = cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Sections_Section

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	var input *cato_models.ClientConnectivityPolicyInput
	if revision != "" {
		input = &cato_models.ClientConnectivityPolicyInput{Revision: &cato_models.PolicyRevisionInput{ID: &revision}}
	}
	result, err := r.getClient().PolicyClientConnectivityPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		return nil, err
	}
	policy := result.GetPolicy().GetClientConnectivity().GetPolicy()
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
