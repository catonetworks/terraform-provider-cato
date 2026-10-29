package rpfsection

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
	PolicyRemotePortFwdAddSection(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdAddSection, error)
	RemotePortFwdPolicy(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyInput *cato_models.RemotePortFwdPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.RemotePortFwdPolicy, error)
	PolicyRemotePortFwdUpdateSection(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdUpdateSection, error)
	PolicyRemotePortFwdMoveSection(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdMoveSection, error)
	PolicyRemotePortFwdRemoveSection(
		ctx context.Context,
		accountID string,
		remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicyRemotePortFwdRemoveSection, error)
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

func (c *sdkClient) PolicyRemotePortFwdAddSection(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdAddSection, error) {
	return c.client.V2().PolicyRemotePortFwdAddSection(
		ctx,
		policyAddSectionInput,
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

func (c *sdkClient) PolicyRemotePortFwdUpdateSection(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdUpdateSection, error) {
	return c.client.V2().PolicyRemotePortFwdUpdateSection(
		ctx,
		policyUpdateSectionInput,
		accountID,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdMoveSection(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdMoveSection, error) {
	return c.client.V2().PolicyRemotePortFwdMoveSection(
		ctx,
		policyMoveSectionInput,
		accountID,
		remotePortFwdPolicyMutationInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicyRemotePortFwdRemoveSection(
	ctx context.Context,
	accountID string,
	remotePortFwdPolicyMutationInput *cato_models.RemotePortFwdPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyRemotePortFwdRemoveSection, error) {
	return c.client.V2().PolicyRemotePortFwdRemoveSection(
		ctx,
		policyRemoveSectionInput,
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

type objectResponse = cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Sections_Section

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
