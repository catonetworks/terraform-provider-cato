package statichost

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type Client interface {
	SiteStaticHost(ctx context.Context, accountID string, siteRefInput cato_models.SiteRefInput, hostID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.SiteStaticHost, error)
}

func (r *staticHostResource) getStaticHostClient() Client {
	if r.staticHostClient != nil {
		return r.staticHostClient
	}
	if r.client == nil {
		return nil
	}

	return r.client.Catov2
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
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
	) (*cato.EntityLookup, error)
	SiteAddStaticHost(
		ctx context.Context,
		siteID string,
		addStaticHostInput cato_models.AddStaticHostInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteAddStaticHost, error)
	SiteRemoveStaticHost(
		ctx context.Context,
		hostID string,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteRemoveStaticHost, error)
	SiteStaticHost(
		ctx context.Context,
		accountID string,
		siteRefInput cato_models.SiteRefInput,
		hostID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteStaticHost, error)
	SiteUpdateStaticHost(
		ctx context.Context,
		hostID string,
		updateStaticHostInput cato_models.UpdateStaticHostInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateStaticHost, error)
}

func (r *staticHostResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
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
) (*cato.EntityLookup, error) {
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

func (c *sdkClient) SiteAddStaticHost(
	ctx context.Context,
	siteID string,
	addStaticHostInput cato_models.AddStaticHostInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteAddStaticHost, error) {
	return c.resource.client.Catov2.SiteAddStaticHost(ctx, siteID, addStaticHostInput, accountID, interceptors...)
}

func (c *sdkClient) SiteRemoveStaticHost(
	ctx context.Context,
	hostID string,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveStaticHost, error) {
	return c.resource.client.Catov2.SiteRemoveStaticHost(ctx, hostID, accountID, interceptors...)
}

func (c *sdkClient) SiteStaticHost(
	ctx context.Context,
	accountID string,
	siteRefInput cato_models.SiteRefInput,
	hostID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteStaticHost, error) {
	return c.resource.getStaticHostClient().SiteStaticHost(ctx, accountID, siteRefInput, hostID, interceptors...)
}

func (c *sdkClient) SiteUpdateStaticHost(
	ctx context.Context,
	hostID string,
	updateStaticHostInput cato_models.UpdateStaticHostInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateStaticHost, error) {
	return c.resource.client.Catov2.SiteUpdateStaticHost(ctx, hostID, updateStaticHostInput, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *staticHostResource }

var _ ResourceClient = (*sdkClient)(nil)
