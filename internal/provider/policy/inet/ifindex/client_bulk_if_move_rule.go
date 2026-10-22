package ifindex

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// InternetFirewallBulkPolicyClient is the minimal Cato client surface used by
// cato_bulk_if_move_rule (mocked in unit tests via mockery).
type InternetFirewallBulkPolicyClient interface {
	PolicyInternetFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwSectionsIndexPolicy, error)
	PolicyInternetFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwRulesIndexPolicy, error)
	PolicyInternetFirewallMoveSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallMoveSection, error)
	PolicyInternetFirewallAddSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSection, error)
	PolicyInternetFirewallReorderPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyReorderInput cato_models.PolicyReorderInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallReorderPolicy, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
}

func (r *ifwRulesIndexResource) ifwBulkPolicy() InternetFirewallBulkPolicyClient {
	if r.ifwBulk != nil {
		return r.ifwBulk
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyInternetFirewallAddSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyAddSectionInput cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallAddSection, error)
	PolicyInternetFirewallMoveSection(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallMoveSection, error)
	PolicyInternetFirewallPublishPolicyRevision(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error)
	PolicyInternetFirewallReorderPolicy(
		ctx context.Context,
		internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
		policyReorderInput cato_models.PolicyReorderInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyInternetFirewallReorderPolicy, error)
	PolicyInternetFirewallRulesIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwRulesIndexPolicy, error)
	PolicyInternetFirewallSectionsIndex(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.IfwSectionsIndexPolicy, error)
}

func (r *ifwRulesIndexResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyInternetFirewallAddSection(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyAddSectionInput cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallAddSection, error) {
	return c.resource.ifwBulkPolicy().PolicyInternetFirewallAddSection(
		ctx,
		internetFirewallPolicyMutationInput,
		policyAddSectionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallMoveSection(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyMoveSectionInput cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallMoveSection, error) {
	return c.resource.ifwBulkPolicy().PolicyInternetFirewallMoveSection(
		ctx,
		internetFirewallPolicyMutationInput,
		policyMoveSectionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallPublishPolicyRevision(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision, error) {
	return c.resource.ifwBulkPolicy().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		internetFirewallPolicyMutationInput,
		policyPublishRevisionInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallReorderPolicy(
	ctx context.Context,
	internetFirewallPolicyMutationInput *cato_models.InternetFirewallPolicyMutationInput,
	policyReorderInput cato_models.PolicyReorderInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyInternetFirewallReorderPolicy, error) {
	return c.resource.ifwBulkPolicy().PolicyInternetFirewallReorderPolicy(
		ctx,
		internetFirewallPolicyMutationInput,
		policyReorderInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) PolicyInternetFirewallRulesIndex(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.IfwRulesIndexPolicy, error) {
	return c.resource.ifwBulkPolicy().PolicyInternetFirewallRulesIndex(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyInternetFirewallSectionsIndex(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.IfwSectionsIndexPolicy, error) {
	return c.resource.ifwBulkPolicy().PolicyInternetFirewallSectionsIndex(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *ifwRulesIndexResource }

var _ ResourceClient = (*sdkClient)(nil)
