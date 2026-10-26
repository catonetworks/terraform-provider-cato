package bgppeer

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	SiteAddBgpPeer(
		ctx context.Context,
		addBgpPeerInput cato_models.AddBgpPeerInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteAddBgpPeer, error)
	SiteBgpPeer(
		ctx context.Context,
		bgpPeerRefInput cato_models.BgpPeerRefInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Site, error)
	SiteRemoveBgpPeer(
		ctx context.Context,
		removeBgpPeerInput cato_models.RemoveBgpPeerInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteRemoveBgpPeer, error)
	SiteUpdateBgpPeer(
		ctx context.Context,
		updateBgpPeerInput cato_models.UpdateBgpPeerInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateBgpPeer, error)
}

func (r *bgpPeerResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) SiteAddBgpPeer(
	ctx context.Context,
	addBgpPeerInput cato_models.AddBgpPeerInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteAddBgpPeer, error) {
	return c.resource.client.Catov2.SiteAddBgpPeer(ctx, addBgpPeerInput, accountID, interceptors...)
}

func (c *sdkClient) SiteBgpPeer(
	ctx context.Context,
	bgpPeerRefInput cato_models.BgpPeerRefInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Site, error) {
	return c.resource.client.Catov2.SiteBgpPeer(ctx, bgpPeerRefInput, accountID, interceptors...)
}

func (c *sdkClient) SiteRemoveBgpPeer(
	ctx context.Context,
	removeBgpPeerInput cato_models.RemoveBgpPeerInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteRemoveBgpPeer, error) {
	return c.resource.client.Catov2.SiteRemoveBgpPeer(ctx, removeBgpPeerInput, accountID, interceptors...)
}

func (c *sdkClient) SiteUpdateBgpPeer(
	ctx context.Context,
	updateBgpPeerInput cato_models.UpdateBgpPeerInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateBgpPeer, error) {
	return c.resource.client.Catov2.SiteUpdateBgpPeer(ctx, updateBgpPeerInput, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *bgpPeerResource }

var _ ResourceClient = (*sdkClient)(nil)
