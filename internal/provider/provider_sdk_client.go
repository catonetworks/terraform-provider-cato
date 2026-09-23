package provider

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

const (
	generatedGroupMembersLimit = 1000

	providerInternetFirewallAddSubPolicyDocument = `mutation policyInternetFirewallAddSubPolicy (
		$internetFirewallPolicyMutationInput: InternetFirewallPolicyMutationInput,
		$internetFirewallAddSubPolicyInput: InternetFirewallAddSubPolicyInput!,
		$accountId: ID!
	) {
		policy(accountId: $accountId) {
			internetFirewall(input: $internetFirewallPolicyMutationInput) {
				addSubPolicy(input: $internetFirewallAddSubPolicyInput) {
					status
					errors {
						errorMessage
						errorCode
					}
				}
			}
		}
	}`

	providerWanFirewallAddSubPolicyDocument = `mutation policyWanFirewallAddSubPolicy (
		$wanFirewallAddSubPolicyInput: WanFirewallAddSubPolicyInput!,
		$accountId: ID!,
		$wanFirewallPolicyMutationInput: WanFirewallPolicyMutationInput
	) {
		policy(accountId: $accountId) {
			wanFirewall(input: $wanFirewallPolicyMutationInput) {
				addSubPolicy(input: $wanFirewallAddSubPolicyInput) {
					status
					errors {
						errorMessage
						errorCode
					}
				}
			}
		}
	}`
)

// providerSDKClient localizes optional generated arguments added by SDK
// regeneration. Existing provider call sites keep their reviewed request shape.
type providerSDKClient struct {
	*cato.Client
}

func newProviderSDKClient(client *cato.Client) *providerSDKClient {
	return &providerSDKClient{Client: client}
}

func generatedGroupMembersInput() cato_models.GroupMembersListInput {
	return cato_models.GroupMembersListInput{
		Paging: &cato_models.PagingInput{
			From:  0,
			Limit: generatedGroupMembersLimit,
		},
		Sort: &cato_models.GroupMembersListSortInput{},
	}
}

func (c *providerSDKClient) GroupsCreateGroup(
	ctx context.Context,
	input cato_models.CreateGroupInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.GroupsCreateGroup, error) {
	return c.Client.GroupsCreateGroup(ctx, input, accountID, generatedGroupMembersInput(), interceptors...)
}

func (c *providerSDKClient) GroupsUpdateGroup(
	ctx context.Context,
	input cato_models.UpdateGroupInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.GroupsUpdateGroup, error) {
	return c.Client.GroupsUpdateGroup(ctx, input, accountID, generatedGroupMembersInput(), interceptors...)
}

func (c *providerSDKClient) GroupsDeleteGroup(
	ctx context.Context,
	input cato_models.GroupRefInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.GroupsDeleteGroup, error) {
	return c.Client.GroupsDeleteGroup(ctx, input, accountID, generatedGroupMembersInput(), interceptors...)
}

func (c *providerSDKClient) PolicyInternetFirewallAddSubPolicy(
	ctx context.Context,
	policyInput *cato_models.InternetFirewallPolicyMutationInput,
	input cato_models.InternetFirewallAddSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyInternetFirewallAddSubPolicy, error) {
	var result cato.PolicyInternetFirewallAddSubPolicy
	err := c.Client.Client.Post(
		ctx,
		"policyInternetFirewallAddSubPolicy",
		providerInternetFirewallAddSubPolicyDocument,
		&result,
		map[string]any{
			"internetFirewallPolicyMutationInput": policyInput,
			"internetFirewallAddSubPolicyInput":   input,
			"accountId":                           accountID,
		},
		interceptors...,
	)
	if err != nil {
		if c.Client.Client.ParseDataWhenErrors {
			return &result, err
		}
		return nil, err
	}
	return &result, nil
}

func (c *providerSDKClient) PolicyInternetFirewallAddRule(
	ctx context.Context,
	input cato_models.InternetFirewallAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyInternetFirewallAddRule, error) {
	return c.Client.PolicyInternetFirewallAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicySocketLanAddSubPolicy(
	ctx context.Context,
	input cato_models.SocketLanAddSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketLanAddSubPolicy, error) {
	return c.Client.PolicySocketLanAddSubPolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicySocketLanMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketLanMoveRule, error) {
	return c.Client.PolicySocketLanMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicySocketLanMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketLanMoveSection, error) {
	return c.Client.PolicySocketLanMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicySocketLanAddRule(
	ctx context.Context,
	input cato_models.SocketLanAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketLanAddRule, error) {
	return c.Client.PolicySocketLanAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicySocketLanAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicySocketLanAddSection, error) {
	return c.Client.PolicySocketLanAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyPrivateAccessUpdatePolicy(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessPolicyUpdateInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessUpdatePolicy, error) {
	return c.Client.PolicyPrivateAccessUpdatePolicy(ctx, accountID, input, nil, interceptors...)
}

func (c *providerSDKClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyReadPrivateAccessPolicy, error) {
	return c.Client.PolicyReadPrivateAccessPolicy(ctx, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyPrivateAccessAddRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessAddRule, error) {
	return c.Client.PolicyPrivateAccessAddRule(ctx, accountID, input, nil, interceptors...)
}

func (c *providerSDKClient) PolicyPrivateAccessUpdateRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessUpdateRule, error) {
	return c.Client.PolicyPrivateAccessUpdateRule(ctx, accountID, input, nil, interceptors...)
}

func (c *providerSDKClient) PolicyPrivateAccessMoveRule(
	ctx context.Context,
	accountID string,
	input cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyPrivateAccessMoveRule, error) {
	return c.Client.PolicyPrivateAccessMoveRule(ctx, accountID, input, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallAddRule(
	ctx context.Context,
	input cato_models.WanFirewallAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallAddRule, error) {
	return c.Client.PolicyWanFirewallAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallAddSubPolicy(
	ctx context.Context,
	input cato_models.WanFirewallAddSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallAddSubPolicy, error) {
	var result cato.PolicyWanFirewallAddSubPolicy
	err := c.Client.Client.Post(
		ctx,
		"policyWanFirewallAddSubPolicy",
		providerWanFirewallAddSubPolicyDocument,
		&result,
		map[string]any{
			"wanFirewallAddSubPolicyInput":   input,
			"accountId":                      accountID,
			"wanFirewallPolicyMutationInput": nil,
		},
		interceptors...,
	)
	if err != nil {
		if c.Client.Client.ParseDataWhenErrors {
			return &result, err
		}
		return nil, err
	}
	return &result, nil
}

func (c *providerSDKClient) PolicyWanFirewallRemoveSubPolicy(
	ctx context.Context,
	input cato_models.WanFirewallRemoveSubPolicyInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallRemoveSubPolicy, error) {
	return c.Client.PolicyWanFirewallRemoveSubPolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallRemoveRule(
	ctx context.Context,
	input cato_models.WanFirewallRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallRemoveRule, error) {
	return c.Client.PolicyWanFirewallRemoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallMoveRule, error) {
	return c.Client.PolicyWanFirewallMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallUpdateRule(
	ctx context.Context,
	input cato_models.WanFirewallUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallUpdateRule, error) {
	return c.Client.PolicyWanFirewallUpdateRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallAddSection, error) {
	return c.Client.PolicyWanFirewallAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallRemoveSection, error) {
	return c.Client.PolicyWanFirewallRemoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallMoveSection, error) {
	return c.Client.PolicyWanFirewallMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallUpdateSection, error) {
	return c.Client.PolicyWanFirewallUpdateSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanFirewallPublishPolicyRevision(
	ctx context.Context,
	input *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanFirewallPublishPolicyRevision, error) {
	return c.Client.PolicyWanFirewallPublishPolicyRevision(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) WanNetworkPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.WanNetworkPolicy, error) {
	return c.Client.WanNetworkPolicy(ctx, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkAddRule(
	ctx context.Context,
	input cato_models.WanNetworkAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkAddRule, error) {
	return c.Client.PolicyWanNetworkAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkRemoveRule(
	ctx context.Context,
	input cato_models.WanNetworkRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkRemoveRule, error) {
	return c.Client.PolicyWanNetworkRemoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkMoveRule, error) {
	return c.Client.PolicyWanNetworkMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkUpdateRule(
	ctx context.Context,
	input cato_models.WanNetworkUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkUpdateRule, error) {
	return c.Client.PolicyWanNetworkUpdateRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkAddSection, error) {
	return c.Client.PolicyWanNetworkAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkRemoveSection, error) {
	return c.Client.PolicyWanNetworkRemoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkMoveSection, error) {
	return c.Client.PolicyWanNetworkMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkUpdateSection, error) {
	return c.Client.PolicyWanNetworkUpdateSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyWanNetworkPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyWanNetworkPublishPolicyRevision, error) {
	return c.Client.PolicyWanNetworkPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}

func (c *providerSDKClient) AppTenantRestrictionPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.AppTenantRestrictionPolicy, error) {
	return c.Client.AppTenantRestrictionPolicy(ctx, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionAddRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionAddRule, error) {
	return c.Client.PolicyAppTenantRestrictionAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionRemoveRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionRemoveRule, error) {
	return c.Client.PolicyAppTenantRestrictionRemoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionMoveRule, error) {
	return c.Client.PolicyAppTenantRestrictionMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionUpdateRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionUpdateRule, error) {
	return c.Client.PolicyAppTenantRestrictionUpdateRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionAddSection, error) {
	return c.Client.PolicyAppTenantRestrictionAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionRemoveSection, error) {
	return c.Client.PolicyAppTenantRestrictionRemoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionMoveSection, error) {
	return c.Client.PolicyAppTenantRestrictionMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionUpdateSection, error) {
	return c.Client.PolicyAppTenantRestrictionUpdateSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionUpdatePolicy(
	ctx context.Context,
	input cato_models.AppTenantRestrictionPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionUpdatePolicy, error) {
	return c.Client.PolicyAppTenantRestrictionUpdatePolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyAppTenantRestrictionPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyAppTenantRestrictionPublishPolicyRevision, error) {
	return c.Client.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}

func (c *providerSDKClient) ApplicationControlPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.ApplicationControlPolicy, error) {
	return c.Client.ApplicationControlPolicy(ctx, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlAddRule(
	ctx context.Context,
	input cato_models.ApplicationControlAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlAddRule, error) {
	return c.Client.PolicyApplicationControlAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlRemoveRule(
	ctx context.Context,
	input cato_models.ApplicationControlRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlRemoveRule, error) {
	return c.Client.PolicyApplicationControlRemoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlMoveRule, error) {
	return c.Client.PolicyApplicationControlMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlUpdateRule(
	ctx context.Context,
	input cato_models.ApplicationControlUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlUpdateRule, error) {
	return c.Client.PolicyApplicationControlUpdateRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlAddSection, error) {
	return c.Client.PolicyApplicationControlAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlRemoveSection, error) {
	return c.Client.PolicyApplicationControlRemoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlMoveSection, error) {
	return c.Client.PolicyApplicationControlMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlUpdateSection, error) {
	return c.Client.PolicyApplicationControlUpdateSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlUpdatePolicy(
	ctx context.Context,
	input cato_models.ApplicationControlPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlUpdatePolicy, error) {
	return c.Client.PolicyApplicationControlUpdatePolicy(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyApplicationControlPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyApplicationControlPublishPolicyRevision, error) {
	return c.Client.PolicyApplicationControlPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}

func (c *providerSDKClient) Tlsinspectpolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.Tlsinspectpolicy, error) {
	return c.Client.Tlsinspectpolicy(ctx, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectAddRule(
	ctx context.Context,
	input cato_models.TLSInspectAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectAddRule, error) {
	return c.Client.PolicyTLSInspectAddRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectRemoveRule(
	ctx context.Context,
	input cato_models.TLSInspectRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectRemoveRule, error) {
	return c.Client.PolicyTLSInspectRemoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectMoveRule, error) {
	return c.Client.PolicyTLSInspectMoveRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectUpdateRule(
	ctx context.Context,
	input cato_models.TLSInspectUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectUpdateRule, error) {
	return c.Client.PolicyTLSInspectUpdateRule(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectAddSection, error) {
	return c.Client.PolicyTLSInspectAddSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectRemoveSection, error) {
	return c.Client.PolicyTLSInspectRemoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectMoveSection, error) {
	return c.Client.PolicyTLSInspectMoveSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectUpdateSection, error) {
	return c.Client.PolicyTLSInspectUpdateSection(ctx, input, accountID, nil, interceptors...)
}

func (c *providerSDKClient) PolicyTLSInspectPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyTLSInspectPublishPolicyRevision, error) {
	return c.Client.PolicyTLSInspectPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
