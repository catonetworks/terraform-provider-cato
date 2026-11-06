package tssection

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyTerminalServerPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPolicy, error)
	PolicyTerminalServerAddSection(
		ctx context.Context,
		input cato_models.PolicyAddSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerAddSection, error)
	PolicyTerminalServerMoveSection(
		ctx context.Context,
		input cato_models.PolicyMoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerMoveSection, error)
	PolicyTerminalServerRemoveSection(
		ctx context.Context,
		input cato_models.PolicyRemoveSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerRemoveSection, error)
	PolicyTerminalServerUpdateSection(
		ctx context.Context,
		input cato_models.PolicyUpdateSectionInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerUpdateSection, error)
	PolicyTerminalServerPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error)
}

func (r *terminalServerSectionResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyTerminalServerPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPolicy, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerAddSection(
	ctx context.Context,
	input cato_models.PolicyAddSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerAddSection, error) {
	return c.resource.client.Catov2.PolicyTerminalServerAddSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerMoveSection(
	ctx context.Context,
	input cato_models.PolicyMoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerMoveSection, error) {
	return c.resource.client.Catov2.PolicyTerminalServerMoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerRemoveSection(
	ctx context.Context,
	input cato_models.PolicyRemoveSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerRemoveSection, error) {
	return c.resource.client.Catov2.PolicyTerminalServerRemoveSection(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyTerminalServerUpdateSection(
	ctx context.Context,
	input cato_models.PolicyUpdateSectionInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerUpdateSection, error) {
	return c.resource.client.Catov2.PolicyTerminalServerUpdateSection(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *terminalServerSectionResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyTerminalServerPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyTerminalServerPublishPolicyRevision(ctx, accountID, interceptors...)
}
