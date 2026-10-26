package socketsite

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
)

type Client interface {
	SiteAddSocketSite(ctx context.Context, addSocketSiteInput cato_models.AddSocketSiteInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteAddSocketSite, error)
	SiteSocketConfiguration(ctx context.Context, input cato_models.SiteSocketConfigurationInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteSocketConfiguration, error)
	SiteRemoveSite(ctx context.Context, siteID string, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteRemoveSite, error)
	SiteUpdateHa(ctx context.Context, accountID string, siteID string, updateHaInput cato_models.UpdateHaInput,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteUpdateHa, error)
}

func (r *socketSiteResource) getSocketSiteClient() Client {
	if r.socketSiteClient != nil {
		return r.socketSiteClient
	}

	if r.client == nil {
		return nil
	}

	return r.client.Catov2
}

func removeSocketSiteWithRetry(ctx context.Context, clientX Client, siteID, accountID string) error {
	const (
		maxAttempts = 12
		retryWait   = 5 * time.Second
	)

	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		_, err = clientX.SiteRemoveSite(ctx, siteID, accountID)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "currently in use") || attempt == maxAttempts-1 {
			return err
		}

		timer := time.NewTimer(retryWait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return err
}

func (r *socketSiteResource) fetchSocketConfiguration(ctx context.Context, siteID string,
	diags *diag.Diagnostics,
) *cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration {
	socketConfiguration, err := r.resourceAPI().SiteSocketConfiguration(ctx,
		cato_models.SiteSocketConfigurationInput{
			Site: &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID},
		},
		r.client.AccountId,
	)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to fetch SiteSocketConfiguration for site '%s'", siteID), err.Error())
		return nil
	}
	if socketConfiguration == nil || socketConfiguration.GetSite().GetSiteSocketConfiguration() == nil {
		diags.AddError(fmt.Sprintf("socket configuration for site '%s' was not returned", siteID), "")
		return nil
	}
	return socketConfiguration.GetSite().GetSiteSocketConfiguration()
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
	) (*cato_go_sdk.EntityLookup, error)
	NetworkRangeList(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkRangeListInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.NetworkRangeList, error)
	SiteAddSocketSite(
		ctx context.Context,
		addSocketSiteInput cato_models.AddSocketSiteInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteAddSocketSite, error)
	SiteExchangeSocketPorts(
		ctx context.Context,
		accountID string,
		exchangeSocketPortsInput cato_models.ExchangeSocketPortsInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteExchangeSocketPorts, error)
	SiteGeneralDetails(
		ctx context.Context,
		siteRefInput cato_models.SiteRefInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteGeneralDetails, error)
	SiteRemoveSite(
		ctx context.Context,
		siteID string,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteRemoveSite, error)
	SiteSocketConfiguration(
		ctx context.Context,
		input cato_models.SiteSocketConfigurationInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteSocketConfiguration, error)
	SiteUpdateHa(
		ctx context.Context,
		accountID string,
		siteID string,
		updateHaInput cato_models.UpdateHaInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateHa, error)
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
	SiteUpdateSocketInterface(
		ctx context.Context,
		siteID string,
		socketInterfaceID cato_models.SocketInterfaceIDEnum,
		updateSocketInterfaceInput cato_models.UpdateSocketInterfaceInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.SiteUpdateSocketInterface, error)
	EntityLookupMinimal(
		ctx context.Context,
		accountID string,
		typeArg cato_models.EntityType,
		limit *int64,
		from *int64,
		parent *cato_models.EntityInput,
		sort []*cato_models.SortInput,
		filters []*cato_models.LookupFilterInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.EntityLookup, error)
}

func (r *socketSiteResource) resourceAPI() ResourceClient {
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

func (c *sdkClient) NetworkRangeList(
	ctx context.Context,
	accountID string,
	input cato_models.NetworkRangeListInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.NetworkRangeList, error) {
	return c.resource.client.Catov2.NetworkRangeList(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) SiteAddSocketSite(
	ctx context.Context,
	addSocketSiteInput cato_models.AddSocketSiteInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteAddSocketSite, error) {
	return c.resource.getSocketSiteClient().SiteAddSocketSite(ctx, addSocketSiteInput, accountID, interceptors...)
}

func (c *sdkClient) SiteExchangeSocketPorts(
	ctx context.Context,
	accountID string,
	exchangeSocketPortsInput cato_models.ExchangeSocketPortsInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteExchangeSocketPorts, error) {
	return c.resource.client.Catov2.SiteExchangeSocketPorts(ctx, accountID, exchangeSocketPortsInput, interceptors...)
}

func (c *sdkClient) SiteGeneralDetails(
	ctx context.Context,
	siteRefInput cato_models.SiteRefInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteGeneralDetails, error) {
	return c.resource.client.Catov2.SiteGeneralDetails(ctx, siteRefInput, accountID, interceptors...)
}

func (c *sdkClient) SiteRemoveSite(
	ctx context.Context,
	siteID string,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteRemoveSite, error) {
	return c.resource.getSocketSiteClient().SiteRemoveSite(ctx, siteID, accountID, interceptors...)
}

func (c *sdkClient) SiteSocketConfiguration(
	ctx context.Context,
	input cato_models.SiteSocketConfigurationInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteSocketConfiguration, error) {
	return c.resource.getSocketSiteClient().SiteSocketConfiguration(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) SiteUpdateHa(
	ctx context.Context,
	accountID string,
	siteID string,
	updateHaInput cato_models.UpdateHaInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateHa, error) {
	return c.resource.getSocketSiteClient().SiteUpdateHa(ctx, accountID, siteID, updateHaInput, interceptors...)
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
type sdkClient struct{ resource *socketSiteResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) EntityLookupMinimal(
	ctx context.Context,
	accountID string,
	typeArg cato_models.EntityType,
	limit *int64,
	from *int64,
	parent *cato_models.EntityInput,
	sort []*cato_models.SortInput,
	filters []*cato_models.LookupFilterInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.EntityLookup, error) {
	return c.resource.client.Catov2.EntityLookupMinimal(ctx, accountID, typeArg, limit, from, parent, sort, filters, interceptors...)
}

// dhcpLookupClient keeps relay-group resolution on the resource API boundary.
type dhcpLookupClient struct {
	ResourceClient
	accountID string
}

func (c dhcpLookupClient) AccountID() string { return c.accountID }
func (r *socketSiteResource) dhcpClient() dhcp.SdkClienter {
	return dhcpLookupClient{ResourceClient: r.resourceAPI(), accountID: r.client.AccountId}
}
