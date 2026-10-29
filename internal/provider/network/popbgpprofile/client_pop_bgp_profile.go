package popbgpprofile

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped PoP BGP profile CRUD contract.
type ResourceClient interface {
	PopLocationMutationsAddBgpProfile(context.Context, string, m.PopLocationAddBgpProfileInput,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationMutationsAddBgpProfile, error)
	PopLocationMutationsUpdateBgpProfile(context.Context, string, m.PopLocationUpdateBgpProfileInput,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationMutationsUpdateBgpProfile, error)
	PopLocationMutationsRemoveBgpProfile(context.Context, string, string,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationMutationsRemoveBgpProfile, error)
	PopLocationsBgpProfile(context.Context, string, string,
		...clientv2.RequestInterceptor,
	) (*cato.PopLocationsBgpProfile, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) PopLocationMutationsAddBgpProfile(ctx context.Context, account string, input m.PopLocationAddBgpProfileInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationMutationsAddBgpProfile, error) {
	return c.client.V2().PopLocationMutationsAddBgpProfile(ctx, account, input, interceptors...)
}
func (c *sdkClient) PopLocationMutationsUpdateBgpProfile(ctx context.Context, account string, input m.PopLocationUpdateBgpProfileInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationMutationsUpdateBgpProfile, error) {
	return c.client.V2().PopLocationMutationsUpdateBgpProfile(ctx, account, input, interceptors...)
}
func (c *sdkClient) PopLocationMutationsRemoveBgpProfile(ctx context.Context, account string, input string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationMutationsRemoveBgpProfile, error) {
	return c.client.V2().PopLocationMutationsRemoveBgpProfile(ctx, account, input, interceptors...)
}
func (c *sdkClient) PopLocationsBgpProfile(ctx context.Context, account string, input string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PopLocationsBgpProfile, error) {
	return c.client.V2().PopLocationsBgpProfile(ctx, account, input, interceptors...)
}

type profileRecord = cato.PopLocationsBgpProfile_PopLocations_BgpProfile

func (r *bgpProfileResource) find(ctx context.Context, id string) (*profileRecord, error) {
	result, err := r.getClient().PopLocationsBgpProfile(ctx, r.client.AccountId, id)
	if err != nil {
		if missingProfile(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil || result.GetPopLocations() == nil {
		return nil, fmt.Errorf("API did not return a PoP location response")
	}
	item := result.GetPopLocations().GetBgpProfile()
	if item != nil && item.GetID() != id {
		return nil, fmt.Errorf("API returned a different BGP profile identity")
	}
	return item, nil
}
func missingProfile(err error) bool {
	return apperror.MatchesGraphQLMessages(err, "BGP profile not found", "GQL-User-Error: BGP profile not found")
}
