package alwaysonpolicy

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ZtnaPolicyResourceClient contains only the SDK operations used by this resource.
type ZtnaPolicyResourceClient interface {
	PolicyZtnaAlwaysOnPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnUpdatePolicy(
		ctx context.Context,
		input cato_models.ZtnaAlwaysOnPolicyUpdateInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnUpdatePolicy, error)
	PolicyZtnaAlwaysOnPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnPublishPolicyRevision, error)
}

func (r *ztnaAlwaysOnPolicyResource) resourceAPI() ZtnaPolicyResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client (draft-targeted:
// nil mutation inputs).
type sdkClient struct {
	resource *ztnaAlwaysOnPolicyResource
}

var _ ZtnaPolicyResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyZtnaAlwaysOnPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnPolicy, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPolicy(ctx, accountID, nil, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnUpdatePolicy(
	ctx context.Context,
	input cato_models.ZtnaAlwaysOnPolicyUpdateInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnUpdatePolicy(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, accountID, nil, nil, interceptors...)
}
