package swprule

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
	PolicySiteWebProxyAddRule(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		siteWebProxyAddRuleInput cato_models.SiteWebProxyAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyAddRule, error)
	PolicySiteWebProxyPolicy(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyInput *cato_models.SiteWebProxyPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyPolicy, error)
	PolicySiteWebProxyUpdateRule(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		siteWebProxyUpdateRuleInput cato_models.SiteWebProxyUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyUpdateRule, error)
	PolicySiteWebProxyMoveRule(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		policyMoveRuleInput cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyMoveRule, error)
	PolicySiteWebProxyRemoveRule(
		ctx context.Context,
		accountID string,
		siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
		siteWebProxyRemoveRuleInput cato_models.SiteWebProxyRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.PolicySiteWebProxyRemoveRule, error)
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

func (c *sdkClient) PolicySiteWebProxyAddRule(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	siteWebProxyAddRuleInput cato_models.SiteWebProxyAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyAddRule, error) {
	result := new(cato.PolicySiteWebProxyAddRule)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyAddRule",
		policySiteWebProxyAddRuleDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"siteWebProxyAddRuleInput":        siteWebProxyAddRuleInput},
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

func (c *sdkClient) PolicySiteWebProxyUpdateRule(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	siteWebProxyUpdateRuleInput cato_models.SiteWebProxyUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyUpdateRule, error) {
	result := new(cato.PolicySiteWebProxyUpdateRule)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyUpdateRule",
		policySiteWebProxyUpdateRuleDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"siteWebProxyUpdateRuleInput":     updateRuleInput(siteWebProxyUpdateRuleInput)},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyMoveRule(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	policyMoveRuleInput cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyMoveRule, error) {
	result := new(cato.PolicySiteWebProxyMoveRule)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyMoveRule",
		policySiteWebProxyMoveRuleDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"policyMoveRuleInput":             policyMoveRuleInput},
		interceptors...)
	return result, err
}

func (c *sdkClient) PolicySiteWebProxyRemoveRule(
	ctx context.Context,
	accountID string,
	siteWebProxyPolicyMutationInput *cato_models.SiteWebProxyPolicyMutationInput,
	siteWebProxyRemoveRuleInput cato_models.SiteWebProxyRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySiteWebProxyRemoveRule, error) {
	result := new(cato.PolicySiteWebProxyRemoveRule)
	err := c.client.V2().Client.Post(ctx,
		"policySiteWebProxyRemoveRule",
		policySiteWebProxyRemoveRuleDocument,
		result,
		map[string]interface{}{"accountId": accountID,
			"siteWebProxyPolicyMutationInput": siteWebProxyPolicyMutationInput,
			"siteWebProxyRemoveRuleInput":     siteWebProxyRemoveRuleInput},
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

type objectResponse = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule

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
							name
							description
							index
							section {
								id
								name
							}
							enabled
							fqdn
							port
							allowedDestinationProtocol
							associatedSite {
								id
								name
							}
							shouldAssociateWithAllSites
							authenticationConfig {
								method
								kerberosConfig {
									isEnabled
									encryptedKeytab
								}
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
							name
							description
							index
							section {
								id
								name
							}
							enabled
							fqdn
							port
							allowedDestinationProtocol
							associatedSite {
								id
								name
							}
							shouldAssociateWithAllSites
							authenticationConfig {
								method
								kerberosConfig {
									isEnabled
									encryptedKeytab
								}
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

const policySiteWebProxyRemoveRuleDocument = `mutation policySiteWebProxyRemoveRule (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $siteWebProxyRemoveRuleInput: SiteWebProxyRemoveRuleInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			removeRule(input: $siteWebProxyRemoveRuleInput) {
				rule {
					rule {
						id
						name
						description
						index
						section {
							id
							name
						}
						enabled
						fqdn
						port
						allowedDestinationProtocol
						associatedSite {
							id
							name
						}
						shouldAssociateWithAllSites
						authenticationConfig {
							method
							kerberosConfig {
								isEnabled
								encryptedKeytab
							}
						}
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

const policySiteWebProxyMoveRuleDocument = `mutation policySiteWebProxyMoveRule (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $policyMoveRuleInput: PolicyMoveRuleInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			moveRule(input: $policyMoveRuleInput) {
				rule {
					rule {
						id
						name
						description
						index
						section {
							id
							name
						}
						enabled
						fqdn
						port
						allowedDestinationProtocol
						associatedSite {
							id
							name
						}
						shouldAssociateWithAllSites
						authenticationConfig {
							method
							kerberosConfig {
								isEnabled
								encryptedKeytab
							}
						}
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

const policySiteWebProxyUpdateRuleDocument = `mutation policySiteWebProxyUpdateRule (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $siteWebProxyUpdateRuleInput: SiteWebProxyUpdateRuleInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			updateRule(input: $siteWebProxyUpdateRuleInput) {
				rule {
					rule {
						id
						name
						description
						index
						section {
							id
							name
						}
						enabled
						fqdn
						port
						allowedDestinationProtocol
						associatedSite {
							id
							name
						}
						shouldAssociateWithAllSites
						authenticationConfig {
							method
							kerberosConfig {
								isEnabled
								encryptedKeytab
							}
						}
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
						name
						description
						index
						section {
							id
							name
						}
						enabled
						fqdn
						port
						allowedDestinationProtocol
						associatedSite {
							id
							name
						}
						shouldAssociateWithAllSites
						authenticationConfig {
							method
							kerberosConfig {
								isEnabled
								encryptedKeytab
							}
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

const policySiteWebProxyAddRuleDocument = `mutation policySiteWebProxyAddRule (
$accountId: ID!,
 $siteWebProxyPolicyMutationInput: SiteWebProxyPolicyMutationInput,
 $siteWebProxyAddRuleInput: SiteWebProxyAddRuleInput!) {
	policy(accountId: $accountId) {
		siteWebProxy(input: $siteWebProxyPolicyMutationInput) {
			addRule(input: $siteWebProxyAddRuleInput) {
				rule {
					rule {
						id
						name
						description
						index
						section {
							id
							name
						}
						enabled
						fqdn
						port
						allowedDestinationProtocol
						associatedSite {
							id
							name
						}
						shouldAssociateWithAllSites
						authenticationConfig {
							method
							kerberosConfig {
								isEnabled
								encryptedKeytab
							}
						}
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
							name
							description
							index
							section {
								id
								name
							}
							enabled
							fqdn
							port
							allowedDestinationProtocol
							associatedSite {
								id
								name
							}
							shouldAssociateWithAllSites
							authenticationConfig {
								method
								kerberosConfig {
									isEnabled
									encryptedKeytab
								}
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

// Keep explicit [] clearing for associated_site despite the released SDK's omitempty tag.
func updateRuleInput(in cato_models.SiteWebProxyUpdateRuleInput) interface{} {
	if in.Rule == nil {
		return in
	}
	r := in.Rule
	sites := r.AssociatedSite
	if sites == nil {
		sites = []*cato_models.SiteRefInput{}
	}
	return map[string]interface{}{"id": in.ID, "rule": map[string]interface{}{
		"allowedDestinationProtocol": r.AllowedDestinationProtocol, "associatedSite": sites, "authenticationConfig": r.AuthenticationConfig,
		"description":                 r.Description,
		"enabled":                     r.Enabled,
		"fqdn":                        r.Fqdn,
		"name":                        r.Name,
		"port":                        r.Port,
		"shouldAssociateWithAllSites": r.ShouldAssociateWithAllSites,
	}}
}
