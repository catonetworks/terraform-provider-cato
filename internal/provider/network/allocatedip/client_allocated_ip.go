package allocatedip

// SDK access layer for cato_allocated_ip — all Catov2 calls live here, behind
// the injectable ResourceClient interface (see resource_layers_test.go).

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient exposes the SDK operations the allocated-ip resource needs.
type ResourceClient interface {
	PopLocationList(
		ctx context.Context,
		accountID string,
		filter *cato_models.PopLocationFilterInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PopLocationsPopLocationList, error)
	AllocatedIPList(
		ctx context.Context,
		accountID string,
		input *cato_models.PopLocationAllocatedIPInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PopLocationsAllocatedIPList, error)
	AddAllocatedIP(
		ctx context.Context,
		accountID string,
		input cato_models.PopLocationAddAllocatedIPInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PopLocationMutationsAddAllocatedIP, error)
	UpdateAllocatedIPDescription(
		ctx context.Context,
		accountID string,
		input cato_models.PopLocationUpdateAllocatedIPDescriptionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PopLocationMutationsUpdateAllocatedIPDescription, error)
	RemoveAllocatedIP(
		ctx context.Context,
		accountID string,
		input cato_models.PopLocationRemoveAllocatedIPInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PopLocationMutationsRemoveAllocatedIP, error)
}

func (r *allocatedIPResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

type sdkClient struct {
	resource *allocatedIPResource
}

func (c *sdkClient) PopLocationList(
	ctx context.Context,
	accountID string,
	filter *cato_models.PopLocationFilterInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PopLocationsPopLocationList, error) {
	return c.resource.client.Catov2.PopLocationsPopLocationList(ctx, accountID, filter, interceptors...)
}

func (c *sdkClient) AllocatedIPList(
	ctx context.Context,
	accountID string,
	input *cato_models.PopLocationAllocatedIPInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PopLocationsAllocatedIPList, error) {
	return c.resource.client.Catov2.PopLocationsAllocatedIPList(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) AddAllocatedIP(
	ctx context.Context,
	accountID string,
	input cato_models.PopLocationAddAllocatedIPInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PopLocationMutationsAddAllocatedIP, error) {
	return c.resource.client.Catov2.PopLocationMutationsAddAllocatedIP(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) UpdateAllocatedIPDescription(
	ctx context.Context,
	accountID string,
	input cato_models.PopLocationUpdateAllocatedIPDescriptionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PopLocationMutationsUpdateAllocatedIPDescription, error) {
	return c.resource.client.Catov2.PopLocationMutationsUpdateAllocatedIPDescription(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) RemoveAllocatedIP(
	ctx context.Context,
	accountID string,
	input cato_models.PopLocationRemoveAllocatedIPInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PopLocationMutationsRemoveAllocatedIP, error) {
	return c.resource.client.Catov2.PopLocationMutationsRemoveAllocatedIP(ctx, accountID, input, interceptors...)
}
