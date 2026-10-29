package allocatedip

import (
	"context"
	"fmt"
	"net/netip"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped public allocated-IP lifecycle contract.
type ResourceClient interface {
	PopLocationMutationsAddAllocatedIP(context.Context, string, m.PopLocationAddAllocatedIPInput,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationMutationsAddAllocatedIP, error)
	PopLocationMutationsUpdateAllocatedIPDescription(context.Context, string, m.PopLocationUpdateAllocatedIPDescriptionInput,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationMutationsUpdateAllocatedIPDescription, error)
	PopLocationMutationsRemoveAllocatedIP(context.Context, string, m.PopLocationRemoveAllocatedIPInput,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationMutationsRemoveAllocatedIP, error)
	PopLocationsAllocatedIPList(context.Context, string, *m.PopLocationAllocatedIPInput,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationsAllocatedIPList, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) PopLocationMutationsAddAllocatedIP(ctx context.Context, account string, input m.PopLocationAddAllocatedIPInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationMutationsAddAllocatedIP, error) {
	return c.client.V2().PopLocationMutationsAddAllocatedIP(ctx, account, input, interceptors...)
}
func (c *sdkClient) PopLocationMutationsUpdateAllocatedIPDescription(
	ctx context.Context, account string, input m.PopLocationUpdateAllocatedIPDescriptionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationMutationsUpdateAllocatedIPDescription, error) {
	return c.client.V2().PopLocationMutationsUpdateAllocatedIPDescription(ctx, account, input, interceptors...)
}
func (c *sdkClient) PopLocationMutationsRemoveAllocatedIP(ctx context.Context, account string, input m.PopLocationRemoveAllocatedIPInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationMutationsRemoveAllocatedIP, error) {
	return c.client.V2().PopLocationMutationsRemoveAllocatedIP(ctx, account, input, interceptors...)
}
func (c *sdkClient) PopLocationsAllocatedIPList(ctx context.Context, account string, input *m.PopLocationAllocatedIPInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationsAllocatedIPList, error) {
	return c.client.V2().PopLocationsAllocatedIPList(ctx, account, input, interceptors...)
}

type allocationRecord = cato.PopLocationsAllocatedIpList_PopLocations_AllocatedIPList_AllocatedIP

func (r *allocatedIPResource) find(ctx context.Context, id string) (*allocationRecord, error) {
	result, err := r.getClient().PopLocationsAllocatedIPList(ctx, r.client.AccountId, &m.PopLocationAllocatedIPInput{
		Filter: &m.PopLocationAllocatedIPFilterInput{ID: &m.StringFilterInput{Eq: &id}},
	})
	if err != nil {
		return nil, err
	}
	page := result.GetPopLocations().GetAllocatedIPList()
	if page == nil || page.GetAllocatedIP() == nil {
		return nil, fmt.Errorf("API did not return a complete allocated-IP list")
	}
	items := page.GetAllocatedIP()
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) != 1 || items[0].GetID() != id {
		return nil, fmt.Errorf("API did not return only the requested allocation identity")
	}
	item := items[0]
	if item.GetAccount().GetID() != r.client.AccountId {
		return nil, fmt.Errorf("API returned an allocation outside the configured account")
	}
	if kind := item.GetAllocationType(); kind == nil || *kind != m.PopLocationAllocatedIPTypeSystem {
		return nil, fmt.Errorf("only SYSTEM allocations support this resource's public lifecycle")
	}
	if ip, err := netip.ParseAddr(item.GetIP()); err != nil || !ip.Is4() {
		return nil, fmt.Errorf("API returned an invalid allocated IPv4 address")
	}
	return item, nil
}
