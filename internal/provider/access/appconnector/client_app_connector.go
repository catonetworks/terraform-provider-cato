package appconnector

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AppConnectorCreateConnector(
		ctx context.Context,
		accountID string,
		newConnector cato_models.AddZtnaAppConnectorInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppConnectorCreateConnector, error)
	AppConnectorDeleteConnector(
		ctx context.Context,
		accountID string,
		connectorRef cato_models.RemoveZtnaAppConnectorInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppConnectorDeleteConnector, error)
	AppConnectorReadConnector(
		ctx context.Context,
		accountID string,
		ref cato_models.ZtnaAppConnectorRefInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppConnectorReadConnector, error)
	AppConnectorUpdateConnector(
		ctx context.Context,
		accountID string,
		updateConnector cato_models.UpdateZtnaAppConnectorInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppConnectorUpdateConnector, error)
	ZtnaAppConnectorAddZtnaAppConnectorBwLicense(
		ctx context.Context,
		accountID string,
		addZtnaAppConnectorBwLicenseInput cato_models.AddZtnaAppConnectorBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ZtnaAppConnectorAddZtnaAppConnectorBwLicense, error)
	ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense(
		ctx context.Context,
		accountID string,
		removeZtnaAppConnectorBwLicenseInput cato_models.RemoveZtnaAppConnectorBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense, error)
	ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense(
		ctx context.Context,
		accountID string,
		updateZtnaAppConnectorBwLicenseInput cato_models.UpdateZtnaAppConnectorBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense, error)
}

func (r *appConnectorResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) AppConnectorCreateConnector(
	ctx context.Context,
	accountID string,
	newConnector cato_models.AddZtnaAppConnectorInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppConnectorCreateConnector, error) {
	return c.resource.client.Catov2.AppConnectorCreateConnector(ctx, accountID, newConnector, interceptors...)
}

func (c *sdkClient) AppConnectorDeleteConnector(
	ctx context.Context,
	accountID string,
	connectorRef cato_models.RemoveZtnaAppConnectorInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppConnectorDeleteConnector, error) {
	return c.resource.client.Catov2.AppConnectorDeleteConnector(ctx, accountID, connectorRef, interceptors...)
}

func (c *sdkClient) AppConnectorReadConnector(
	ctx context.Context,
	accountID string,
	ref cato_models.ZtnaAppConnectorRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppConnectorReadConnector, error) {
	return c.resource.client.Catov2.AppConnectorReadConnector(ctx, accountID, ref, interceptors...)
}

func (c *sdkClient) AppConnectorUpdateConnector(
	ctx context.Context,
	accountID string,
	updateConnector cato_models.UpdateZtnaAppConnectorInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppConnectorUpdateConnector, error) {
	return c.resource.client.Catov2.AppConnectorUpdateConnector(ctx, accountID, updateConnector, interceptors...)
}

func (c *sdkClient) ZtnaAppConnectorAddZtnaAppConnectorBwLicense(
	ctx context.Context,
	accountID string,
	addZtnaAppConnectorBwLicenseInput cato_models.AddZtnaAppConnectorBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ZtnaAppConnectorAddZtnaAppConnectorBwLicense, error) {
	return c.resource.client.Catov2.ZtnaAppConnectorAddZtnaAppConnectorBwLicense(
		ctx,
		accountID,
		addZtnaAppConnectorBwLicenseInput,
		interceptors...,
	)
}

func (c *sdkClient) ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense(
	ctx context.Context,
	accountID string,
	removeZtnaAppConnectorBwLicenseInput cato_models.RemoveZtnaAppConnectorBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense, error) {
	return c.resource.client.Catov2.ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense(
		ctx,
		accountID,
		removeZtnaAppConnectorBwLicenseInput,
		interceptors...,
	)
}

func (c *sdkClient) ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense(
	ctx context.Context,
	accountID string,
	updateZtnaAppConnectorBwLicenseInput cato_models.UpdateZtnaAppConnectorBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense, error) {
	return c.resource.client.Catov2.ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense(
		ctx,
		accountID,
		updateZtnaAppConnectorBwLicenseInput,
		interceptors...,
	)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *appConnectorResource }

var _ ResourceClient = (*sdkClient)(nil)
