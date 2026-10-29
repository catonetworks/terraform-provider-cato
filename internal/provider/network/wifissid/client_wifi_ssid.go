package wifissid

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped Wi-Fi SSID API contract.
type ResourceClient interface {
	SiteGeneralDetails(context.Context, m.SiteRefInput, string, ...clientv2.RequestInterceptor) (*cato.SiteGeneralDetails, error)
	EntityLookup(context.Context, string, m.EntityType, *int64, *int64, *m.EntityInput, *string, []string,
		[]*m.SortInput, []*m.LookupFilterInput, []string, ...clientv2.RequestInterceptor) (*cato.EntityLookup, error)

	SiteCreateWifiSsid(context.Context, string, m.CreateWifiSsidInput, ...clientv2.RequestInterceptor) (*cato.SiteCreateWifiSsid, error)
	SiteUpdateWifiSsid(context.Context, string, m.UpdateWifiSsidInput, ...clientv2.RequestInterceptor) (*cato.SiteUpdateWifiSsid, error)
	SiteRemoveWifiSsid(context.Context, string, m.RemoveWifiSsidInput, ...clientv2.RequestInterceptor) (*cato.SiteRemoveWifiSsid, error)
	SiteWifiSsid(context.Context, string, m.WifiSsidRefInput, ...clientv2.RequestInterceptor) (*cato.SiteWifiSsid, error)
	SiteWifiSsidList(context.Context, string, m.WifiSsidListInput, ...clientv2.RequestInterceptor) (*cato.SiteWifiSsidList, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) SiteCreateWifiSsid(
	ctx context.Context, account string, in m.CreateWifiSsidInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteCreateWifiSsid, error) {
	return c.client.V2().SiteCreateWifiSsid(ctx, account, in, i...)
}
func (c *sdkClient) SiteUpdateWifiSsid(
	ctx context.Context, account string, in m.UpdateWifiSsidInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateWifiSsid, error) {
	return c.client.V2().SiteUpdateWifiSsid(ctx, account, in, i...)
}
func (c *sdkClient) SiteRemoveWifiSsid(
	ctx context.Context, account string, in m.RemoveWifiSsidInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveWifiSsid, error) {
	// The backend clears transient IP fields before returning the deleted object.
	// Selecting them produces non-null GraphQL errors after successful deletion.
	result := new(cato.SiteRemoveWifiSsid)
	err := c.client.V2().Client.Post(ctx, "siteRemoveWifiSsid", removeSSIDDocument, result, map[string]interface{}{
		"accountId": account, "removeWifiSsidInput": in,
	}, i...)
	return result, err
}
func (c *sdkClient) SiteWifiSsid(
	ctx context.Context, account string, in m.WifiSsidRefInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteWifiSsid, error) {
	return c.client.V2().SiteWifiSsid(ctx, account, in, i...)
}
func (c *sdkClient) SiteWifiSsidList(
	ctx context.Context, account string, in m.WifiSsidListInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteWifiSsidList, error) {
	return c.client.V2().SiteWifiSsidList(ctx, account, in, i...)
}

func (c *sdkClient) SiteGeneralDetails(ctx context.Context, ref m.SiteRefInput, account string,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteGeneralDetails, error) {
	return c.client.V2().SiteGeneralDetails(ctx, ref, account, i...)
}
func (c *sdkClient) EntityLookup(ctx context.Context, account string, kind m.EntityType, limit, from *int64,
	parent *m.EntityInput, search *string, ids []string, sort []*m.SortInput, filters []*m.LookupFilterInput,
	fields []string, i ...clientv2.RequestInterceptor,
) (*cato.EntityLookup, error) {
	return c.client.V2().EntityLookup(ctx, account, kind, limit, from, parent, search, ids, sort, filters, fields, i...)
}

var errStaleSSID = errors.New("the SSID interface configuration no longer exists; its local row can be removed or reclaimed")

func staleSSID(err error, id string) bool {
	var api *clientv2.ErrorResponse
	if !errors.As(err, &api) || api.NetworkError != nil || api.GqlErrors == nil || len(*api.GqlErrors) == 0 {
		return false
	}
	pattern := regexp.MustCompile(`^Wifi SSID '[^'\r\n]*' \(ID ` + regexp.QuoteMeta(id) + `\) not found$`)
	for _, e := range *api.GqlErrors {
		if e == nil || !pattern.MatchString(strings.TrimPrefix(e.Message, "GQL-User-Error: ")) {
			return false
		}
	}
	return true
}

type ssidRecord = cato.SiteWifiSsid_Site_WifiSsid

func missingSSID(err error, id string) bool {
	var api *clientv2.ErrorResponse
	if !errors.As(err, &api) || api.NetworkError != nil || api.GqlErrors == nil || len(*api.GqlErrors) == 0 {
		return false
	}
	expected := "Wifi SSID " + id + " not found"
	for _, e := range *api.GqlErrors {
		if e == nil || (e.Message != expected && e.Message != "GQL-User-Error: "+expected) {
			return false
		}
	}
	return true
}

// Never include SDK request bodies or transport errors in diagnostics: writes contain PSKs.
func safeAPIError(err error, secret string) string {
	var api *clientv2.ErrorResponse
	if errors.As(err, &api) && api.GqlErrors != nil && len(*api.GqlErrors) > 0 {
		messages := []string{}
		for _, e := range *api.GqlErrors {
			if e != nil {
				message := e.Message
				if secret != "" {
					message = strings.ReplaceAll(message, secret, "[REDACTED]")
				}
				messages = append(messages, message)
			}
		}
		return strings.Join(messages, "; ")
	}
	return "API request failed; SDK request details are omitted because they may contain a passkey."
}
func (r *ssidResource) find(ctx context.Context, id string) (*ssidRecord, error) {
	result, err := r.getClient().SiteWifiSsid(ctx, r.client.AccountId, m.WifiSsidRefInput{By: m.ObjectRefByID, Input: id})
	if err != nil {
		if staleSSID(err, id) {
			return nil, errStaleSSID
		}
		if missingSSID(err, id) {
			return nil, nil
		}
		return nil, err
	}
	item := result.GetSite().GetWifiSsid()
	if item == nil || item.GetID() != id || item.GetSite().GetID() == "" || item.GetNetworkInterfaceID() == "" {
		return nil, fmt.Errorf("API did not confirm the requested SSID identity")
	}
	return item, nil
}
func (r *ssidResource) slotOccupied(ctx context.Context, in *m.CreateWifiSsidInput) (bool, error) {
	result, err := r.getClient().SiteWifiSsidList(ctx, r.client.AccountId, m.WifiSsidListInput{Site: in.Site})
	if err != nil {
		return false, err
	}
	payload := result.GetSite().GetWifiSsidList()
	if payload == nil || payload.GetPageInfo() == nil || payload.GetPageInfo().GetTotal() != int64(len(payload.GetWifiSsid())) {
		return false, fmt.Errorf("API returned an incomplete SSID list")
	}
	for _, item := range payload.GetWifiSsid() {
		if item == nil || item.GetID() == "" || !matchesReference(in.Site, item.GetSite().GetID(), item.GetSite().GetName()) {
			return false, fmt.Errorf("API returned inconsistent SSID ownership")
		}
		if item.GetNetworkInterfaceID() == in.NetworkInterfaceID {
			return true, nil
		}
	}
	return r.rawInterfaceOccupied(ctx, in)
}

// The local SSID list cannot detect legacy CC2-only Wi-Fi configuration. The
// networkInterface lookup reads CC2 NetworkInterface/SocketInterface tables.
func (r *ssidResource) rawInterfaceOccupied(ctx context.Context, in *m.CreateWifiSsidInput) (bool, error) {
	details, err := r.getClient().SiteGeneralDetails(ctx, *in.Site, r.client.AccountId)
	if err != nil {
		return false, err
	}
	site := details.GetSite().GetSiteGeneralDetails().GetSite()
	if site.GetID() == "" || !matchesReference(in.Site, site.GetID(), site.GetName()) {
		return false, fmt.Errorf("site lookup did not confirm the requested identity")
	}
	// Pin resolved ID so a name change cannot redirect the subsequent create.
	in.Site = &m.SiteRefInput{By: m.ObjectRefByID, Input: site.GetID()}
	parent := &m.EntityInput{Type: m.EntityTypeSite, ID: site.GetID()}
	const pageSize int64 = 200
	from := int64(0)
	for {
		limit := pageSize
		result, err := r.getClient().EntityLookup(ctx, r.client.AccountId, m.EntityTypeNetworkInterface, &limit, &from, parent,
			nil, nil, nil, nil, []string{"siteId", "interfaceId"})
		if err != nil {
			return false, err
		}
		page := result.GetEntityLookup()
		if page == nil || page.GetTotal() == nil || page.GetItems() == nil || *page.GetTotal() < from+int64(len(page.GetItems())) {
			return false, fmt.Errorf("incomplete account/site-scoped interface lookup")
		}
		occupied, err := interfaceRowsOccupied(page.GetItems(), site.GetID(), in.NetworkInterfaceID)
		if err != nil || occupied {
			return occupied, err
		}
		from += int64(len(page.GetItems()))
		if from == *page.GetTotal() {
			return false, nil
		}
		if len(page.GetItems()) == 0 {
			return false, fmt.Errorf("interface lookup ended before the declared total")
		}
	}
}
func helperIdentity(value any) (string, error) {
	switch v := value.(type) {
	case string:
		if v != "" {
			return v, nil
		}
	case float64:
		if v >= 0 && v <= 9007199254740991 && math.Trunc(v) == v {
			return strconv.FormatInt(int64(v), 10), nil
		}
	}
	return "", fmt.Errorf("missing or noncanonical identifier")
}

const removeSSIDDocument = `mutation siteRemoveWifiSsid($accountId: ID!, $removeWifiSsidInput: RemoveWifiSsidInput!) {
 site(accountId: $accountId) {
  removeWifiSsid(input: $removeWifiSsidInput) { wifiSsid { id networkInterfaceId site { id name } } }
 }
}`

func interfaceRowsOccupied(items []*cato.EntityLookup_EntityLookup_Items, siteID, portID string) (bool, error) {
	for _, item := range items {
		if item == nil || item.GetEntity().GetID() == "" {
			return false, fmt.Errorf("interface lookup returned an invalid identity")
		}
		owner, err := helperIdentity(item.GetHelperFields()["siteId"])
		if err != nil || owner != siteID {
			return false, fmt.Errorf("interface lookup returned inconsistent site ownership")
		}
		port, err := helperIdentity(item.GetHelperFields()["interfaceId"])
		if err != nil {
			return false, fmt.Errorf("interface lookup did not identify a physical port")
		}
		if port == portID {
			return true, nil
		}
	}
	return false, nil
}

// Mutation errors may echo the PSK in a different encoding; do not forward their text.
func safeMutationError(err error, passkey string) string {
	if passkey != "" {
		return "API mutation failed; detailed SDK errors are omitted because the request contains a passkey."
	}
	return safeAPIError(err, "")
}
