package privateapp

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PrivateAppCreatePrivateApp(
		ctx context.Context,
		accountID string,
		input cato_models.CreatePrivateApplicationInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PrivateAppCreatePrivateApp, error)
	PrivateAppDeletePrivateApp(
		ctx context.Context,
		accountID string,
		input cato_models.DeletePrivateApplicationInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PrivateAppDeletePrivateApp, error)
	PrivateAppReadPrivateApp(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateApplicationRefInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PrivateAppReadPrivateApp, error)
	PrivateAppUpdatePrivateApp(
		ctx context.Context,
		accountID string,
		input cato_models.UpdatePrivateApplicationInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PrivateAppUpdatePrivateApp, error)
}

func (r *privateAppResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PrivateAppCreatePrivateApp(
	ctx context.Context,
	accountID string,
	input cato_models.CreatePrivateApplicationInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PrivateAppCreatePrivateApp, error) {
	return c.resource.client.Catov2.PrivateAppCreatePrivateApp(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PrivateAppDeletePrivateApp(
	ctx context.Context,
	accountID string,
	input cato_models.DeletePrivateApplicationInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PrivateAppDeletePrivateApp, error) {
	return c.resource.client.Catov2.PrivateAppDeletePrivateApp(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PrivateAppReadPrivateApp(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateApplicationRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PrivateAppReadPrivateApp, error) {
	return c.resource.client.Catov2.PrivateAppReadPrivateApp(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PrivateAppUpdatePrivateApp(
	ctx context.Context,
	accountID string,
	input cato_models.UpdatePrivateApplicationInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PrivateAppUpdatePrivateApp, error) {
	return c.resource.client.Catov2.PrivateAppUpdatePrivateApp(ctx, accountID, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *privateAppResource }

var _ ResourceClient = (*sdkClient)(nil)
