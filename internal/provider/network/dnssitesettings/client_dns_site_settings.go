package dnssitesettings

import (
	"context"
	"errors"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	SiteExists(context.Context, string, string) (bool, error)
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

func (c *sdkClient) SiteExists(ctx context.Context, accountID, siteID string) (bool, error) {
	response, err := c.EntityLookup(ctx, accountID, cato_models.EntityTypeSite, nil, nil, nil, nil, []string{siteID}, nil, nil, nil)
	if err != nil {
		return false, err
	}
	if response == nil || response.EntityLookup.Items == nil || response.EntityLookup.Total == nil {
		return false, fmt.Errorf("site lookup returned an absent or malformed response")
	}
	items := response.EntityLookup.Items
	total := *response.EntityLookup.Total
	if len(items) == 0 && total == 0 {
		return false, nil
	}
	if len(items) == 1 && total == 1 && items[0] != nil && items[0].Entity.ID == siteID {
		return true, nil
	}
	return false, fmt.Errorf("site lookup returned unexpected items or total")
}

// Backend wrappers do not distinguish a deleted site from other failures.
func confirmParentAbsent(ctx context.Context, client ResourceClient, accountID, siteID, leaf string, operationErr error) (bool, error) {
	if !isSiteBackendError(operationErr, leaf) {
		return false, operationErr
	}
	exists, err := client.SiteExists(ctx, accountID, siteID)
	if err != nil {
		return false, fmt.Errorf("%w; parent-site lookup failed: %v", operationErr, err)
	}
	if exists {
		return false, operationErr
	}
	return true, nil
}

func isSiteBackendError(err error, leaf string) bool {
	var response *clientv2.ErrorResponse
	var list *clientv2.GqlErrorList
	switch {
	case errors.As(err, &response):
		if response.NetworkError != nil || response.GqlErrors == nil {
			return false
		}
		list = &clientv2.GqlErrorList{Errors: *response.GqlErrors}
	case errors.As(err, &list):
	default:
		return false
	}
	if len(list.Errors) == 0 {
		return false
	}
	for _, item := range list.Errors {
		code := item.Extensions["code"]
		if (code != "CC2_ERROR" && code != "DOWNSTREAM_SERVICE_ERROR") || len(item.Path) == 0 || fmt.Sprint(item.Path[len(item.Path)-1]) != leaf {
			return false
		}
	}
	return true
}
