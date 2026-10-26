package dnssitesettings

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	SiteGeneralDetails(context.Context, cato_models.SiteRefInput, string, ...clientv2.RequestInterceptor) (*cato.SiteGeneralDetails, error)
	NetworkConfigDNSRemoveSiteSettings(
		ctx context.Context,
		accountID string,
		input cato_models.NetworkConfigDNSRemoveSiteSettingsInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSRemoveSiteSettings, error)
	NetworkConfigDNSSiteSettings(
		ctx context.Context,
		accountID string,
		siteRefInput cato_models.SiteRefInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSSiteSettings, error)
	UpdateDNSSiteSettings(
		ctx context.Context,
		accountID string,
		site cato_models.SiteRefInput,
		input nc.DNSSettingsInput,
	) (*cato.NetworkConfigDNSUpdateSiteSettings, error)
}
type sdkClient struct{ *nc.DNSClient }

var _ ResourceClient = (*sdkClient)(nil)

func (r *siteSettingsResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{nc.NewDNSClient(r.client.Catov2)}
}

func lookupSite(ctx context.Context,
	client ResourceClient,
	account string,
	ref cato_models.SiteRefInput) (*cato.SiteGeneralDetails_Site_SiteGeneralDetails_Site,
	error) {
	response, err := client.SiteGeneralDetails(ctx, ref, account)
	if err != nil {
		return nil, err
	}
	details := response.GetSite().GetSiteGeneralDetails()
	if details == nil || details.Site.ID == "" || details.Site.Name == "" {
		return nil, fmt.Errorf("site lookup returned an absent or malformed reference")
	}
	if ref.By == cato_models.ObjectRefByID && details.Site.ID != ref.Input {
		return nil, fmt.Errorf("site lookup returned a different site")
	}
	return &details.Site, nil
}
