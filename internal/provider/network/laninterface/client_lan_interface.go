package laninterface

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AccountSnapshot(
		ctx context.Context,
		siteIDs []string,
		userIDs []string,
		accountID *string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AccountSnapshot, error)
	EntityLookup(
		ctx context.Context,
		accountID string,
		typeArg cato_models.EntityType,
		limit *int64,
		from *int64,
		parent *cato_models.EntityInput,
		search *string,
		entityIDs []string,
		sort []*cato_models.SortInput,
		filters []*cato_models.LookupFilterInput,
		helperFields []string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.EntityLookup, error)
	SiteUpdateSocketInterface(
		ctx context.Context,
		siteID string,
		socketInterfaceID cato_models.SocketInterfaceIDEnum,
		updateSocketInterfaceInput cato_models.UpdateSocketInterfaceInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateSocketInterface, error)
}

func (r *lanInterfaceResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) AccountSnapshot(
	ctx context.Context,
	siteIDs []string,
	userIDs []string,
	accountID *string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AccountSnapshot, error) {
	return c.resource.client.Catov2.AccountSnapshot(ctx, siteIDs, userIDs, accountID, interceptors...)
}

func (c *sdkClient) EntityLookup(
	ctx context.Context,
	accountID string,
	typeArg cato_models.EntityType,
	limit *int64,
	from *int64,
	parent *cato_models.EntityInput,
	search *string,
	entityIDs []string,
	sort []*cato_models.SortInput,
	filters []*cato_models.LookupFilterInput,
	helperFields []string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.EntityLookup, error) {
	return c.resource.client.Catov2.EntityLookup(
		ctx,
		accountID,
		typeArg,
		limit,
		from,
		parent,
		search,
		entityIDs,
		sort,
		filters,
		helperFields,
		interceptors...,
	)
}

func (c *sdkClient) SiteUpdateSocketInterface(
	ctx context.Context,
	siteID string,
	socketInterfaceID cato_models.SocketInterfaceIDEnum,
	updateSocketInterfaceInput cato_models.UpdateSocketInterfaceInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateSocketInterface, error) {
	return c.resource.client.Catov2.SiteUpdateSocketInterface(
		ctx,
		siteID,
		socketInterfaceID,
		updateSocketInterfaceInput,
		accountID,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *lanInterfaceResource }

var _ ResourceClient = (*sdkClient)(nil)
