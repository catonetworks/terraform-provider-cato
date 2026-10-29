package cloudinterconnectsite

import (
	"context"
	"errors"
	"fmt"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped Cloud Interconnect site contract.
type ResourceClient interface {
	LookupSite(context.Context, string, string) (bool, error)
	SiteAddCloudInterconnectSite(ctx context.Context, input m.AddCloudInterconnectSiteInput, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteAddCloudInterconnectSite, error)
	SiteUpdateSiteGeneralDetails(ctx context.Context, id string, input m.UpdateSiteGeneralDetailsInput, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateSiteGeneralDetails, error)
	SiteRemoveSite(ctx context.Context, id, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteRemoveSite, error)
	SiteGeneralDetails(ctx context.Context, input m.SiteRefInput, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteGeneralDetails, error)
	SiteCloudInterconnectPhysicalConnectionID(ctx context.Context, account string, input m.CloudInterconnectPhysicalConnectionIDInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteCloudInterconnectPhysicalConnectionID, error)
	AccountSnapshot(ctx context.Context, siteIDs, userIDs []string, account *string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.AccountSnapshot, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) SiteAddCloudInterconnectSite(ctx context.Context, input m.AddCloudInterconnectSiteInput, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteAddCloudInterconnectSite, error) {
	return c.client.V2().SiteAddCloudInterconnectSite(ctx, input, account, interceptors...)
}
func (c *sdkClient) SiteUpdateSiteGeneralDetails(ctx context.Context, id string, input m.UpdateSiteGeneralDetailsInput, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateSiteGeneralDetails, error) {
	return c.client.V2().SiteUpdateSiteGeneralDetails(ctx, id, input, account, interceptors...)
}
func (c *sdkClient) SiteRemoveSite(ctx context.Context, id, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveSite, error) {
	return c.client.V2().SiteRemoveSite(ctx, id, account, interceptors...)
}
func (c *sdkClient) SiteGeneralDetails(ctx context.Context, input m.SiteRefInput, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteGeneralDetails, error) {
	return c.client.V2().SiteGeneralDetails(ctx, input, account, interceptors...)
}
func (c *sdkClient) SiteCloudInterconnectPhysicalConnectionID(
	ctx context.Context, account string, input m.CloudInterconnectPhysicalConnectionIDInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteCloudInterconnectPhysicalConnectionID, error) {
	return c.client.V2().SiteCloudInterconnectPhysicalConnectionID(ctx, account, input, interceptors...)
}
func (c *sdkClient) AccountSnapshot(ctx context.Context, siteIDs, userIDs []string, account *string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.AccountSnapshot, error) {
	return c.client.V2().AccountSnapshot(ctx, siteIDs, userIDs, account, interceptors...)
}

func (c *sdkClient) LookupSite(ctx context.Context, account, id string) (bool, error) {
	limit, from := int64(1), int64(0)
	result, err := c.client.V2().EntityLookup(ctx, account, m.EntityTypeSite, &limit, &from, nil, nil, []string{id}, nil, nil, nil)
	if err != nil {
		return false, err
	}
	page := result.GetEntityLookup()
	if page == nil || page.GetItems() == nil {
		return false, fmt.Errorf("API did not return a complete site lookup")
	}
	items := page.GetItems()
	if len(items) == 0 {
		return false, nil
	}
	if len(items) != 1 || items[0].GetEntity().GetID() != id {
		return false, fmt.Errorf("site lookup returned an unexpected identity")
	}
	return true, nil
}

var errSnapshotPending = errors.New("site exists in configuration but its snapshot is not yet available")

type siteRecord = cato.SiteGeneralDetails_Site_SiteGeneralDetails

func (r *cloudSiteResource) find(ctx context.Context, id string) (*siteRecord, error) {
	exists, lookupErr := r.getClient().LookupSite(ctx, r.client.AccountId, id)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if !exists {
		return nil, nil
	}
	snapshot, err := r.getClient().AccountSnapshot(ctx, []string{id}, []string{}, &r.client.AccountId)
	if err != nil {
		return nil, err
	}
	account := snapshot.GetAccountSnapshot()
	if account.GetID() == nil || *account.GetID() != r.client.AccountId {
		return nil, fmt.Errorf("snapshot account identity is missing or differs from configured account")
	}
	sites := account.GetSites()
	if len(sites) == 0 {
		return nil, errSnapshotPending
	}
	if len(sites) != 1 || sites[0].GetID() == nil || *sites[0].GetID() != id {
		return nil, fmt.Errorf("snapshot did not return only the requested site")
	}
	kind := sites[0].GetInfoSiteSnapshot().GetConnType()
	if kind == nil || *kind != m.ProtoTypeCrossConnect {
		return nil, fmt.Errorf("only CROSS_CONNECT sites can be managed by this resource")
	}
	result, err := r.getClient().SiteGeneralDetails(ctx, m.SiteRefInput{By: m.ObjectRefByID, Input: id}, r.client.AccountId)
	if err != nil {
		return nil, err
	}
	item := result.GetSite().GetSiteGeneralDetails()
	if item == nil || item.GetSite().GetID() != id {
		return nil, fmt.Errorf("API did not return the requested site details")
	}
	return item, nil
}

const consistencyTimeout = 30 * time.Second
const consistencyInterval = 500 * time.Millisecond

func (r *cloudSiteResource) waitForSite(ctx context.Context, id string, expected m.UpdateSiteGeneralDetailsInput) (*siteRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, consistencyTimeout)
	defer cancel()
	ticker := time.NewTicker(consistencyInterval)
	defer ticker.Stop()
	for {
		item, err := r.find(ctx, id)
		if err != nil && !errors.Is(err, errSnapshotPending) {
			return nil, err
		}
		if item != nil && matchesSite(item, expected) {
			return item, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for site read convergence: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
func matchesSite(item *siteRecord, input m.UpdateSiteGeneralDetailsInput) bool {
	if input.Name != nil && item.GetSite().GetName() != *input.Name {
		return false
	}
	if input.SiteType != nil && (item.GetSiteType() == nil || *item.GetSiteType() != *input.SiteType) {
		return false
	}
	if input.Description != nil && valueOrEmpty(item.GetDescription()) != *input.Description {
		return false
	}
	location := input.SiteLocation
	if location == nil {
		return true
	}
	actual := item.GetSiteLocation()
	for _, pair := range [][2]*string{
		{location.CountryCode, &actual.CountryCode}, {location.Timezone, &actual.Timezone},
		{location.CityName, actual.CityName}, {location.StateCode, actual.StateCode}, {location.Address, actual.Address},
	} {
		if pair[0] != nil && *pair[0] != valueOrEmpty(pair[1]) {
			return false
		}
	}
	return true
}
func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func (r *cloudSiteResource) hasPhysicalConnections(ctx context.Context, id string) (bool, error) {
	for _, role := range m.AllHaRole {
		result, err := r.getClient().SiteCloudInterconnectPhysicalConnectionID(ctx, r.client.AccountId,
			m.CloudInterconnectPhysicalConnectionIDInput{Site: &m.SiteRefInput{By: m.ObjectRefByID, Input: id}, HaRole: role})
		if err != nil {
			if missingConnection(err) {
				continue
			}
			return false, err
		}
		if result.GetSite().GetCloudInterconnectPhysicalConnectionID().GetID() == "" {
			return false, fmt.Errorf("connection lookup returned no identity or not-found error")
		}
		return true, nil
	}
	return false, nil
}
func missingConnection(err error) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item == nil || (item.Message != "connection not found" && item.Message != "GQL-User-Error: connection not found") {
			return false
		}
	}
	return true
}
