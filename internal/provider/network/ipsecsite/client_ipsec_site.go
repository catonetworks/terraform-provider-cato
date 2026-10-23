package ipsecsite

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
	SiteAddIpsecIkeV2Site(
		ctx context.Context,
		addIpsecIkeV2SiteInput cato_models.AddIpsecIkeV2SiteInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteAddIpsecIkeV2Site, error)
	SiteAddIpsecIkeV2SiteTunnels(
		ctx context.Context,
		siteID string,
		addIpsecIkeV2SiteTunnelsInput cato_models.AddIpsecIkeV2SiteTunnelsInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteAddIpsecIkeV2SiteTunnels, error)
	SiteRemoveSite(
		ctx context.Context,
		siteID string,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteRemoveSite, error)
	SiteUpdateIpsecIkeV2SiteGeneralDetails(
		ctx context.Context,
		siteID string,
		updateIpsecIkeV2SiteGeneralDetailsInput cato_models.UpdateIpsecIkeV2SiteGeneralDetailsInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateIpsecIkeV2SiteGeneralDetails, error)
	SiteUpdateIpsecIkeV2SiteTunnels(
		ctx context.Context,
		siteID string,
		updateIpsecIkeV2SiteTunnelsInput cato_models.UpdateIpsecIkeV2SiteTunnelsInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateIpsecIkeV2SiteTunnels, error)
	SiteUpdateNetworkRange(
		ctx context.Context,
		networkRangeID string,
		updateNetworkRangeInput cato_models.UpdateNetworkRangeInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateNetworkRange, error)
	SiteUpdateSiteGeneralDetails(
		ctx context.Context,
		siteID string,
		updateSiteGeneralDetailsInput cato_models.UpdateSiteGeneralDetailsInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateSiteGeneralDetails, error)
}

func (r *siteIpsecResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) SiteAddIpsecIkeV2Site(
	ctx context.Context,
	addIpsecIkeV2SiteInput cato_models.AddIpsecIkeV2SiteInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteAddIpsecIkeV2Site, error) {
	return c.resource.client.Catov2.SiteAddIpsecIkeV2Site(ctx, addIpsecIkeV2SiteInput, accountID, interceptors...)
}

func (c *sdkClient) SiteAddIpsecIkeV2SiteTunnels(
	ctx context.Context,
	siteID string,
	addIpsecIkeV2SiteTunnelsInput cato_models.AddIpsecIkeV2SiteTunnelsInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteAddIpsecIkeV2SiteTunnels, error) {
	return c.resource.client.Catov2.SiteAddIpsecIkeV2SiteTunnels(ctx, siteID, addIpsecIkeV2SiteTunnelsInput, accountID, interceptors...)
}

func (c *sdkClient) SiteRemoveSite(
	ctx context.Context,
	siteID string,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteRemoveSite, error) {
	return c.resource.client.Catov2.SiteRemoveSite(ctx, siteID, accountID, interceptors...)
}

func (c *sdkClient) SiteUpdateIpsecIkeV2SiteGeneralDetails(
	ctx context.Context,
	siteID string,
	updateIpsecIkeV2SiteGeneralDetailsInput cato_models.UpdateIpsecIkeV2SiteGeneralDetailsInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateIpsecIkeV2SiteGeneralDetails, error) {
	return c.resource.client.Catov2.SiteUpdateIpsecIkeV2SiteGeneralDetails(
		ctx,
		siteID,
		updateIpsecIkeV2SiteGeneralDetailsInput,
		accountID,
		interceptors...,
	)
}

func (c *sdkClient) SiteUpdateIpsecIkeV2SiteTunnels(
	ctx context.Context,
	siteID string,
	updateIpsecIkeV2SiteTunnelsInput cato_models.UpdateIpsecIkeV2SiteTunnelsInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateIpsecIkeV2SiteTunnels, error) {
	return c.resource.client.Catov2.SiteUpdateIpsecIkeV2SiteTunnels(ctx, siteID, updateIpsecIkeV2SiteTunnelsInput, accountID, interceptors...)
}

func (c *sdkClient) SiteUpdateNetworkRange(
	ctx context.Context,
	networkRangeID string,
	updateNetworkRangeInput cato_models.UpdateNetworkRangeInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateNetworkRange, error) {
	return c.resource.client.Catov2.SiteUpdateNetworkRange(ctx, networkRangeID, updateNetworkRangeInput, accountID, interceptors...)
}

func (c *sdkClient) SiteUpdateSiteGeneralDetails(
	ctx context.Context,
	siteID string,
	updateSiteGeneralDetailsInput cato_models.UpdateSiteGeneralDetailsInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateSiteGeneralDetails, error) {
	return c.resource.client.Catov2.SiteUpdateSiteGeneralDetails(ctx, siteID, updateSiteGeneralDetailsInput, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *siteIpsecResource }

var _ ResourceClient = (*sdkClient)(nil)
