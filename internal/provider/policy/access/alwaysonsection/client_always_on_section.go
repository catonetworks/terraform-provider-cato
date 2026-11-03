package alwaysonsection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyZtnaAlwaysOnAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnAddSection, error)
	PolicyZtnaAlwaysOnMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnMoveSection, error)
	PolicyZtnaAlwaysOnPolicy(
		ctx context.Context,
		accountID string,
		input *cato_models.ZtnaAlwaysOnPolicyInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnRemoveSection, error)
	PolicyZtnaAlwaysOnUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyZtnaAlwaysOnUpdateSection, error)
}

func (r *alwaysOnSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

// sdkClient forwards operations to the configured SDK client. All mutations
// target the draft policy (nil mutation input) - this resource never
// publishes; cato_bulk_ao_move_rule is the Always-On stack's publisher.
type sdkClient struct{ resource *alwaysOnSectionResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyZtnaAlwaysOnAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnAddSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnAddSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnMoveSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnMoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnPolicy(
	ctx context.Context,
	accountID string,
	input *cato_models.ZtnaAlwaysOnPolicyInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnPolicy, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnRemoveSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnRemoveSection(ctx, accountID, nil, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyZtnaAlwaysOnUpdateSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnUpdateSection(ctx, accountID, nil, input, interceptors...)
}
