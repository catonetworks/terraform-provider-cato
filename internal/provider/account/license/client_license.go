package license

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	Licensing(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.Licensing, error)
	RemoveSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.RemoveSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.RemoveSiteBwLicense, error)
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
	UpdateSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.UpdateSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.UpdateSiteBwLicense, error)
	ReplaceSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.ReplaceSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ReplaceSiteBwLicense, error)
	AssignSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.AssignSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AssignSiteBwLicense, error)
}

func (r *licenseResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) Licensing(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Licensing, error) {
	return c.resource.client.Catov2.Licensing(ctx, accountID, interceptors...)
}

func (c *sdkClient) RemoveSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.RemoveSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.RemoveSiteBwLicense, error) {
	return c.resource.client.Catov2.RemoveSiteBwLicense(ctx, accountID, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *licenseResource }

var _ ResourceClient = (*sdkClient)(nil)

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
func (c *sdkClient) UpdateSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.UpdateSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.UpdateSiteBwLicense, error) {
	return c.resource.client.Catov2.UpdateSiteBwLicense(ctx, accountID, input, interceptors...)
}
func (c *sdkClient) ReplaceSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.ReplaceSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ReplaceSiteBwLicense, error) {
	return c.resource.client.Catov2.ReplaceSiteBwLicense(ctx, accountID, input, interceptors...)
}
func (c *sdkClient) AssignSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.AssignSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AssignSiteBwLicense, error) {
	return c.resource.client.Catov2.AssignSiteBwLicense(ctx, accountID, input, interceptors...)
}

// licenseAPI selects an injected client or the configured SDK adapter.
func licenseAPI(service *catoclient.Service, clients ...ResourceClient) ResourceClient {
	if len(clients) > 0 {
		return clients[0]
	}
	return service.Catov2
}
