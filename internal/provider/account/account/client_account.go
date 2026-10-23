package account

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AccountManagement(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AccountManagement, error)
	AccountManagementAddAccount(
		ctx context.Context,
		addAccountInput cato_models.AddAccountInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AccountManagementAddAccount, error)
	AccountManagementRemoveAccount(
		ctx context.Context,
		accountIDToRemove string,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AccountManagementRemoveAccount, error)
	AccountManagementUpdateAccount(
		ctx context.Context,
		updateAccountInput cato_models.UpdateAccountInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AccountManagementUpdateAccount, error)
}

func (r *accountResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) AccountManagement(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AccountManagement, error) {
	return c.resource.client.Catov2.AccountManagement(ctx, accountID, interceptors...)
}

func (c *sdkClient) AccountManagementAddAccount(
	ctx context.Context,
	addAccountInput cato_models.AddAccountInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AccountManagementAddAccount, error) {
	return c.resource.client.Catov2.AccountManagementAddAccount(ctx, addAccountInput, accountID, interceptors...)
}

func (c *sdkClient) AccountManagementRemoveAccount(
	ctx context.Context,
	accountIDToRemove string,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AccountManagementRemoveAccount, error) {
	return c.resource.client.Catov2.AccountManagementRemoveAccount(ctx, accountIDToRemove, accountID, interceptors...)
}

func (c *sdkClient) AccountManagementUpdateAccount(
	ctx context.Context,
	updateAccountInput cato_models.UpdateAccountInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AccountManagementUpdateAccount, error) {
	return c.resource.client.Catov2.AccountManagementUpdateAccount(ctx, updateAccountInput, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *accountResource }

var _ ResourceClient = (*sdkClient)(nil)
