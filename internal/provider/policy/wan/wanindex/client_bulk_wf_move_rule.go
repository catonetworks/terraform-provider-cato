package wanindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// WanFirewallBulkPolicyClient is the minimal Cato client surface used by
// cato_bulk_wf_move_rule (mocked in unit tests via mockery).
type WanFirewallBulkPolicyClient interface {
	PolicyWanFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanSectionsIndexPolicy, error)
	PolicyWanFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanRulesIndexPolicy, error)
	PolicyWanFirewallMoveSection(
		ctx context.Context,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallMoveSection, error)
	PolicyWanFirewallReorderPolicy(
		ctx context.Context,
		wanFirewallPolicyMutationInput *cato_models.WanFirewallPolicyMutationInput,
		policyReorderInput cato_models.PolicyReorderInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallReorderPolicy, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
}

func (r *wanRulesIndexResource) wanBulkPolicy() WanFirewallBulkPolicyClient {
	if r.wanBulk != nil {
		return r.wanBulk
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	DiscardPolicyRevisions(ctx context.Context) error
	PolicyWanFirewallMoveSection(
		ctx context.Context,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallMoveSection, error)
	PolicyWanFirewallPublishPolicyRevision(
		ctx context.Context,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error)
	PolicyWanFirewallReorderPolicy(
		ctx context.Context,
		wanFirewallPolicyMutationInput *cato_models.WanFirewallPolicyMutationInput,
		policyReorderInput cato_models.PolicyReorderInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyWanFirewallReorderPolicy, error)
	PolicyWanFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanRulesIndexPolicy, error)
	PolicyWanFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.WanSectionsIndexPolicy, error)
}

func (r *wanRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyWanFirewallMoveSection(
	ctx context.Context,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallMoveSection, error) {
	return c.resource.wanBulkPolicy().PolicyWanFirewallMoveSection(ctx, policyMoveSectionInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallPublishPolicyRevision(
	ctx context.Context,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallPublishPolicyRevision, error) {
	return c.resource.wanBulkPolicy().PolicyWanFirewallPublishPolicyRevision(ctx, policyPublishRevisionInput, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallReorderPolicy(
	ctx context.Context,
	wanFirewallPolicyMutationInput *cato_models.WanFirewallPolicyMutationInput,
	policyReorderInput cato_models.PolicyReorderInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyWanFirewallReorderPolicy, error) {
	return c.resource.wanBulkPolicy().PolicyWanFirewallReorderPolicy(
		ctx,
		wanFirewallPolicyMutationInput,
		policyReorderInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyWanFirewallRulesIndex(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanRulesIndexPolicy, error) {
	return c.resource.wanBulkPolicy().PolicyWanFirewallRulesIndex(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyWanFirewallSectionsIndex(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.WanSectionsIndexPolicy, error) {
	return c.resource.wanBulkPolicy().PolicyWanFirewallSectionsIndex(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *wanRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) DiscardPolicyRevisions(ctx context.Context) error {
	return policy.DiscardFirewallAndWANPolicyRevisions(ctx, c.resource.client.Catov2.Client, c.resource.client.AccountId)
}
