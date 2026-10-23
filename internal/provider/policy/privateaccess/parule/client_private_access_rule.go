package parule

import (
	"context"
	"fmt"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyPrivateAccessAddRule(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessAddRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessAddRule, error)
	PolicyPrivateAccessDeleteRule(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessRemoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessDeleteRule, error)
	PolicyPrivateAccessPublishRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessPublishRevision, error)
	PolicyPrivateAccessUpdateRule(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessUpdateRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessUpdateRule, error)
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error)
}

func (r *privAccessRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyPrivateAccessAddRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessAddRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessAddRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessAddRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessDeleteRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessRemoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessDeleteRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessDeleteRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessPublishRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessPublishRevision, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessPublishRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessUpdateRule(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessUpdateRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessUpdateRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessUpdateRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error) {
	return c.resource.client.Catov2.PolicyReadPrivateAccessPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *privAccessRuleResource }

var _ ResourceClient = (*sdkClient)(nil)

// lookupPrivateAccessRule retries only a missing rule immediately after a successful create.
// Refresh reads remain immediate so externally deleted rules are removed from state.
func lookupPrivateAccessRule(
	ctx context.Context,
	client ResourceClient,
	accountID, ruleID string,
	waitForCreation bool,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		result, err := client.PolicyReadPrivateAccessPolicy(ctx, accountID)
		if err != nil {
			return nil, err
		}
		if result == nil || result.Policy.PrivateAccess == nil {
			return nil, fmt.Errorf("private access policy not returned")
		}
		for _, rule := range result.Policy.PrivateAccess.Policy.Rules {
			if rule == nil {
				return nil, fmt.Errorf("nil private access rule returned")
			}
			if rule.Rule.ID == ruleID {
				return rule, nil
			}
		}
		if !waitForCreation {
			return nil, ErrPrivateAcccessRuleNotFound
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("waiting for created private access rule %s: %w", ruleID, ctx.Err())
		case <-timer.C:
		}
	}
}
