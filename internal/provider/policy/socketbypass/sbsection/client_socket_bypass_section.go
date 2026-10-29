package sbsection

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
	PolicySocketBypassAddSection(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassAddSection, error)
	PolicySocketBypassPolicy(
		ctx context.Context,
		accountID string,
		socketBypassPolicyInput *cato_models.SocketBypassPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassPolicy, error)
	PolicySocketBypassUpdateSection(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassUpdateSection, error)
	PolicySocketBypassMoveSection(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassMoveSection, error)
	PolicySocketBypassRemoveSection(
		ctx context.Context,
		accountID string,
		socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySocketBypassRemoveSection, error)
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

func (c *sdkClient) PolicySocketBypassAddSection(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassAddSection, error) {
	return c.client.Catov2.PolicySocketBypassAddSection(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyAddSectionInput,
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

func (c *sdkClient) PolicySocketBypassUpdateSection(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassUpdateSection, error) {
	return c.client.Catov2.PolicySocketBypassUpdateSection(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyUpdateSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassMoveSection(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassMoveSection, error) {
	return c.client.Catov2.PolicySocketBypassMoveSection(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyMoveSectionInput,
		interceptors...,
	)
}

func (c *sdkClient) PolicySocketBypassRemoveSection(
	ctx context.Context,
	accountID string,
	socketBypassPolicyMutationInput *cato_models.SocketBypassPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketBypassRemoveSection, error) {
	return c.client.Catov2.PolicySocketBypassRemoveSection(
		ctx,
		accountID,
		socketBypassPolicyMutationInput,
		policyRemoveSectionInput,
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

type objectResponse = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Sections_Section

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
