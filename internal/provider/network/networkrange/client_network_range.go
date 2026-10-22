package networkrange

import (
	"context"
	"errors"
	"fmt"
	"strings"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	cast "github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
)

type NetworkRangeClient interface {
	SiteAddNetworkRange(ctx context.Context, lanSocketInterfaceID string, addNetworkRangeInput cato_models.AddNetworkRangeInput,
		accountID string, interceptors ...clientv2.RequestInterceptor) (*cato.SiteAddNetworkRange, error)
	SiteUpdateNetworkRange(ctx context.Context, networkRangeID string, updateNetworkRangeInput cato_models.UpdateNetworkRangeInput,
		accountID string, interceptors ...clientv2.RequestInterceptor) (*cato.SiteUpdateNetworkRange, error)
	SiteRemoveNetworkRange(ctx context.Context, networkRangeID string, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.SiteRemoveNetworkRange, error)
	NetworkRange(ctx context.Context, accountID string, networkRangeID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkRange, error)
	EntityLookup(ctx context.Context, accountID string, typeArg cato_models.EntityType, limit *int64, from *int64,
		parent *cato_models.EntityInput, search *string, entityIDs []string, sort []*cato_models.SortInput,
		filters []*cato_models.LookupFilterInput, helperFields []string, interceptors ...clientv2.RequestInterceptor) (*cato.EntityLookup, error)
}

func (r *networkRangeResource) getNetworkRangeClient() NetworkRangeClient {
	if r.networkRangeClient != nil {
		return r.networkRangeClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

// fetchNetworkRange retrieves the network range details from the API, and returns nil if the network range is not found,
// on error the diags gets updated.
func (r *networkRangeResource) fetchNetworkRange(ctx context.Context, networkRangeID string, diags *diag.Diagnostics,
) (responseRange *cato.NetworkRange_Site_NetworkRange) {
	const notFoundMsg = "Invalid network range id: "

	queryRangeResult, err := r.resourceAPI().NetworkRange(ctx, r.client.AccountId, networkRangeID)
	if err != nil {
		// Check if error is not found error, if so return (nil, false, nil) to indicate resource should be removed from state without error
		if gqlError, ok := errors.AsType[*clientv2.ErrorResponse](err); ok {
			if (gqlError.GqlErrors != nil) && (len(*gqlError.GqlErrors) > 0) && strings.Contains((*gqlError.GqlErrors)[0].Message, notFoundMsg) {
				return nil
			}
		}
		diags.AddError(
			"Catov2 NetworkRange API error",
			fmt.Sprintf("error fetching network range details for range ID '%s': %v", networkRangeID, err),
		)
		return nil
	}
	if queryRangeResult == nil || queryRangeResult.Site.NetworkRange == nil {
		return nil
	}
	return queryRangeResult.Site.NetworkRange
}

// getSiteIdFromNetworkRange retrieves the site_id and interface info for a network range using entityLookup
func (r *networkRangeResource) getSiteIDFromNetworkRange(ctx context.Context, networkRangeID string,
) (siteID string, interfaceName string, err error) {
	result, err := r.resourceAPI().EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"),
		nil, nil, nil, nil, []string{networkRangeID}, nil, nil, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to lookup network range site: %w", err)
	}

	if result == nil || len(result.EntityLookup.GetItems()) == 0 {
		return "", "", fmt.Errorf("network range %s not found in entityLookup", networkRangeID)
	}

	item := result.EntityLookup.GetItems()[0]
	helperFields := item.GetHelperFields()
	if helperFields == nil {
		return "", "", fmt.Errorf("no helperFields returned for network range %s", networkRangeID)
	}

	// Extract siteID from helperFields
	siteID = cast.ToString(helperFields["siteId"])
	if siteID == "" {
		return "", "", fmt.Errorf("siteId not found in helperFields for network range %s", networkRangeID)
	}

	// Extract interfaceName from helperFields
	interfaceName = cast.ToString(helperFields["interfaceName"])

	return siteID, interfaceName, nil
}

// getInterfaceIndex looks up the network interface Index for a given site ID and interface ID
func (r *networkRangeResource) getInterfaceIndex(ctx context.Context, siteID, interfaceID string, diags *diag.Diagnostics,
) (interfaceIndex string) {
	site := &cato_models.EntityInput{Type: cato_models.EntityTypeSite, ID: siteID}
	networkInterfaceResponse, err := r.resourceAPI().EntityLookup(ctx, r.client.AccountId, cato_models.EntityTypeNetworkInterface,
		new(int64(0)), nil, site, nil, []string{interfaceID}, nil, nil, nil)
	if err != nil {
		diags.AddError("Error retrieving network interface "+interfaceID, err.Error())
		return ""
	}
	for _, item := range networkInterfaceResponse.GetEntityLookup().GetItems() {
		if item.GetEntity().GetID() == interfaceID {
			helperFields := item.GetHelperFields()
			curIfaceIndex := cast.ToString(helperFields["interfaceId"])
			if numberRE.MatchString(curIfaceIndex) {
				curIfaceIndex = "INT_" + curIfaceIndex
			}
			return curIfaceIndex
		}
	}
	diags.AddError("Error retrieving network interface",
		"network interface with id '"+interfaceID+"' not found in site '"+siteID+"'")
	return ""
}

// getInterfaceID looks up the network interface ID for a given site ID and interface index
func (r *networkRangeResource) getInterfaceID(ctx context.Context, siteID, interfaceIndex string, diags *diag.Diagnostics,
) (interfaceID string) {
	site := &cato_models.EntityInput{Type: cato_models.EntityTypeSite, ID: siteID}
	networkInterfaceResponse, err := r.resourceAPI().EntityLookup(ctx, r.client.AccountId,
		cato_models.EntityTypeNetworkInterface, new(int64(0)), nil, site, nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Error retrieving network interface "+interfaceIndex, err.Error())
		return ""
	}
	for _, item := range networkInterfaceResponse.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		curIfaceIndex := cast.ToString(helperFields["interfaceId"])
		if numberRE.MatchString(curIfaceIndex) {
			curIfaceIndex = "INT_" + curIfaceIndex
		}
		if curIfaceIndex == interfaceIndex {
			return item.GetEntity().GetID()
		}
	}
	diags.AddError("Error retrieving network interface",
		"network interface with index '"+interfaceIndex+"' not found in site '"+siteID+"'")
	return ""
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
	NetworkRange(
		ctx context.Context,
		accountID string,
		networkRangeID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.NetworkRange, error)
	SiteAddNetworkRange(
		ctx context.Context,
		lanSocketInterfaceID string,
		addNetworkRangeInput cato_models.AddNetworkRangeInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteAddNetworkRange, error)
	SiteRemoveNetworkRange(
		ctx context.Context,
		networkRangeID string,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteRemoveNetworkRange, error)
	SiteUpdateNetworkRange(
		ctx context.Context,
		networkRangeID string,
		updateNetworkRangeInput cato_models.UpdateNetworkRangeInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateNetworkRange, error)
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
	) (*cato.EntityLookup, error)
}

func (r *networkRangeResource) resourceAPI() ResourceClient {
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
	return c.resource.getNetworkRangeClient().EntityLookup(
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

func (c *sdkClient) NetworkRange(
	ctx context.Context,
	accountID string,
	networkRangeID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkRange, error) {
	return c.resource.getNetworkRangeClient().NetworkRange(ctx, accountID, networkRangeID, interceptors...)
}

func (c *sdkClient) SiteAddNetworkRange(
	ctx context.Context,
	lanSocketInterfaceID string,
	addNetworkRangeInput cato_models.AddNetworkRangeInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteAddNetworkRange, error) {
	return c.resource.getNetworkRangeClient().SiteAddNetworkRange(ctx, lanSocketInterfaceID, addNetworkRangeInput, accountID, interceptors...)
}

func (c *sdkClient) SiteRemoveNetworkRange(
	ctx context.Context,
	networkRangeID string,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveNetworkRange, error) {
	return c.resource.getNetworkRangeClient().SiteRemoveNetworkRange(ctx, networkRangeID, accountID, interceptors...)
}

func (c *sdkClient) SiteUpdateNetworkRange(
	ctx context.Context,
	networkRangeID string,
	updateNetworkRangeInput cato_models.UpdateNetworkRangeInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateNetworkRange, error) {
	return c.resource.getNetworkRangeClient().SiteUpdateNetworkRange(ctx, networkRangeID, updateNetworkRangeInput, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *networkRangeResource }

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
) (*cato.EntityLookup, error) {
	return c.resource.client.Catov2.EntityLookupMinimal(ctx, accountID, typeArg, limit, from, parent, sort, filters, interceptors...)
}

// dhcpLookupClient keeps relay-group resolution on the resource API boundary.
type dhcpLookupClient struct {
	ResourceClient
	accountID string
}

func (c dhcpLookupClient) AccountID() string { return c.accountID }
func (r *networkRangeResource) dhcpClient() dhcp.SdkClienter {
	return dhcpLookupClient{ResourceClient: r.resourceAPI(), accountID: r.client.AccountId}
}
