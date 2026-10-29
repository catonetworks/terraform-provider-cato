package swpsection

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
	PolicySiteWebProxyCreatePolicyRevision(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyCreatePolicyRevision, error)
	PolicySiteWebProxyAddSection(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyAddSection, error)
	PolicySiteWebProxyPolicy(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyInput *cato_models.SiteWebProxyPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyPolicy, error)
	PolicySiteWebProxyUpdateSection(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyUpdateSection, error)
	PolicySiteWebProxyMoveSection(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyMoveSection, error)
	PolicySiteWebProxyRemoveSection(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyRemoveSection, error)
	PolicySiteWebProxyPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyPublishPolicyRevision, error)
	PolicySiteWebProxyDiscardPolicyRevision(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyDiscardPolicyRevision, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicySiteWebProxyCreatePolicyRevision(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyCreateRevisionInput cato_models.PolicyCreateRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyCreatePolicyRevision, error) {
	result := new(cato.PolicySiteWebProxyCreatePolicyRevision)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyCreatePolicyRevision",
		policySiteWebProxyCreatePolicyRevisionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyCreateRevisionInput":       policyCreateRevisionInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyAddSection(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyAddSection, error) {
	result := new(cato.PolicySiteWebProxyAddSection)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyAddSection",
		policySiteWebProxyAddSectionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyAddSectionInput":           policyAddSectionInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyPolicy(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyInput *cato_models.SiteWebProxyPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyPolicy, error) {
	result := new(cato.PolicySiteWebProxyPolicy)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyPolicy",
		policySiteWebProxyPolicyDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyInput": siteWebProxyPolicyInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyUpdateSection(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyUpdateSectionInput cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyUpdateSection, error) {
	result := new(cato.PolicySiteWebProxyUpdateSection)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyUpdateSection",
		policySiteWebProxyUpdateSectionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyUpdateSectionInput":        policyUpdateSectionInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyMoveSection(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyMoveSection, error) {
	result := new(cato.PolicySiteWebProxyMoveSection)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyMoveSection",
		policySiteWebProxyMoveSectionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyMoveSectionInput":          policyMoveSectionInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyRemoveSection(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyRemoveSectionInput cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyRemoveSection, error) {
	result := new(cato.PolicySiteWebProxyRemoveSection)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyRemoveSection",
		policySiteWebProxyRemoveSectionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyRemoveSectionInput":        policyRemoveSectionInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyPublishPolicyRevision, error) {
	result := new(cato.PolicySiteWebProxyPublishPolicyRevision)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyPublishPolicyRevision",
		policySiteWebProxyPublishPolicyRevisionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyPublishRevisionInput":      policyPublishRevisionInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyDiscardPolicyRevision(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyDiscardRevisionInput *cato_models.PolicyDiscardRevisionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyDiscardPolicyRevision, error) {
	result := new(cato.PolicySiteWebProxyDiscardPolicyRevision)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyDiscardPolicyRevision",
		policySiteWebProxyDiscardPolicyRevisionDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyDiscardRevisionInput":      policyDiscardRevisionInput},
		interceptors...)
	return result, err
}

type policyResponse = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy

type objectResponse = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Sections_Section

func (r *policyResource) readPolicy(ctx context.Context, revision string) (*policyResponse, error) {
	revisionType := cato_models.PolicyRevisionTypePublic
	input := &cato_models.SiteWebProxyPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}}
	if revision != "" {
		revisionType = cato_models.PolicyRevisionTypePrivate
		input.Revision.ID = &revision
	}
	result, err := r.getClient().PolicySiteWebProxyPolicy(ctx, r.client.AccountId, input)
	if err != nil {
		if revisionMissing(err) {
			return nil, fmt.Errorf("%w: %v", errMissingRevision, err)
		}
		return nil, err
	}
	policy := result.GetPolicy().GetSiteWebProxy().GetPolicy()
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

// Keep SDK variable contracts and transport, selecting only owned resource data.
const policySiteWebProxyDiscardPolicyRevisionDocument = `mutation policySiteWebProxyDiscardPolicyRevision (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyDiscardRevisionInput: PolicyDiscardRevisionInput) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			discardPolicyRevision(input: $policyDiscardRevisionInput) {
				policy {
					enabled
					rules {
						rule {
							id
							section {
								id
								name
							}
						}
						properties
					}
					sections {
						section {
							id
							name
						}
						properties
					}
					revision {
						id
						name
						description
						changes
						createdTime
						updatedTime
					}
				}
				status
				errors {
					errorMessage
					errorCode
				}
			}
		}
	}
}
`

const policySiteWebProxyPublishPolicyRevisionDocument = `mutation policySiteWebProxyPublishPolicyRevision (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyPublishRevisionInput: PolicyPublishRevisionInput) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			publishPolicyRevision(input: $policyPublishRevisionInput) {
				policy {
					enabled
					rules {
						rule {
							id
							section {
								id
								name
							}
						}
						properties
					}
					sections {
						section {
							id
							name
						}
						properties
					}
					revision {
						id
						name
						description
						changes
						createdTime
						updatedTime
					}
				}
				status
				errors {
					errorMessage
					errorCode
				}
			}
		}
	}
}
`

const policySiteWebProxyRemoveSectionDocument = `mutation policySiteWebProxyRemoveSection (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyRemoveSectionInput: PolicyRemoveSectionInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			removeSection(input: $policyRemoveSectionInput) {
				section {
					section {
						id
						name
					}
					properties
				}
				status
				errors {
					errorMessage
					errorCode
				}
				revision {
					id
					name
					description
					changes
					createdTime
					updatedTime
				}
			}
		}
	}
}
`

const policySiteWebProxyMoveSectionDocument = `mutation policySiteWebProxyMoveSection (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyMoveSectionInput: PolicyMoveSectionInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			moveSection(input: $policyMoveSectionInput) {
				section {
					section {
						id
						name
					}
					properties
				}
				status
				errors {
					errorMessage
					errorCode
				}
				revision {
					id
					name
					description
					changes
					createdTime
					updatedTime
				}
			}
		}
	}
}
`

const policySiteWebProxyUpdateSectionDocument = `mutation policySiteWebProxyUpdateSection (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyUpdateSectionInput: PolicyUpdateSectionInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			updateSection(input: $policyUpdateSectionInput) {
				section {
					section {
						id
						name
					}
					properties
				}
				status
				errors {
					errorMessage
					errorCode
				}
				revision {
					id
					name
					description
					changes
					createdTime
					updatedTime
				}
			}
		}
	}
}
`

const policySiteWebProxyPolicyDocument = `query policySiteWebProxyPolicy (
$accountId: ID!,
 $siteWebProxyPolicyInput: SiteWebProxyPolicyInput) {
	policy(accountId: $accountId) {
		siteWebProxy {
			policy(input: $siteWebProxyPolicyInput) {
				enabled
				rules {
					rule {
						id
						section {
							id
							name
						}
					}
					properties
				}
				sections {
					section {
						id
						name
					}
					properties
				}
				revision {
					id
					name
					description
					changes
					createdTime
					updatedTime
				}
			}
		}
	}
}
`

const policySiteWebProxyAddSectionDocument = `mutation policySiteWebProxyAddSection (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyAddSectionInput: PolicyAddSectionInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			addSection(input: $policyAddSectionInput) {
				section {
					section {
						id
						name
					}
					properties
				}
				status
				errors {
					errorMessage
					errorCode
				}
				revision {
					id
					name
					description
					changes
					createdTime
					updatedTime
				}
			}
		}
	}
}
`

const policySiteWebProxyCreatePolicyRevisionDocument = `mutation policySiteWebProxyCreatePolicyRevision (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyCreateRevisionInput: PolicyCreateRevisionInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			createPolicyRevision(input: $policyCreateRevisionInput) {
				policy {
					enabled
					rules {
						rule {
							id
							section {
								id
								name
							}
						}
						properties
					}
					sections {
						section {
							id
							name
						}
						properties
					}
					revision {
						id
						name
						description
						changes
						createdTime
						updatedTime
					}
				}
				status
				errors {
					errorMessage
					errorCode
				}
			}
		}
	}
}
`
