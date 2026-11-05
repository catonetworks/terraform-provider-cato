package admin

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AccountRoles(ctx context.Context, accountID string, accountType *cato_models.AccountType,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.AccountRoles, error)
	Admin(
		ctx context.Context,
		accountID string,
		adminID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Admin, error)
	AdminAddAdmin(
		ctx context.Context,
		addAdminInput cato_models.AddAdminInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AdminAddAdmin, error)
	AdminRemoveAdmin(
		ctx context.Context,
		adminID string,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AdminRemoveAdmin, error)
	AdminUpdateAdmin(
		ctx context.Context,
		adminID string,
		updateAdminInput cato_models.UpdateAdminInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AdminUpdateAdmin, error)
	Admins(
		ctx context.Context,
		accountID string,
		limit *int64,
		from *int64,
		search *string,
		sort []*cato_models.SortInput,
		adminIDs []string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.Admins, error)
}

func (r *adminResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) Admin(
	ctx context.Context,
	accountID string,
	adminID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Admin, error) {
	return c.resource.client.Catov2.Admin(ctx, accountID, adminID, interceptors...)
}

func (c *sdkClient) AdminAddAdmin(
	ctx context.Context,
	addAdminInput cato_models.AddAdminInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AdminAddAdmin, error) {
	return c.resource.client.Catov2.AdminAddAdmin(ctx, addAdminInput, accountID, interceptors...)
}

func (c *sdkClient) AdminRemoveAdmin(
	ctx context.Context,
	adminID string,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AdminRemoveAdmin, error) {
	return c.resource.client.Catov2.AdminRemoveAdmin(ctx, adminID, accountID, interceptors...)
}

func (c *sdkClient) AdminUpdateAdmin(
	ctx context.Context,
	adminID string,
	updateAdminInput cato_models.UpdateAdminInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AdminUpdateAdmin, error) {
	return c.resource.client.Catov2.AdminUpdateAdmin(ctx, adminID, updateAdminInput, accountID, interceptors...)
}

func (c *sdkClient) Admins(
	ctx context.Context,
	accountID string,
	limit *int64,
	from *int64,
	search *string,
	sort []*cato_models.SortInput,
	adminIDs []string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Admins, error) {
	return c.resource.client.Catov2.Admins(ctx, accountID, limit, from, search, sort, adminIDs, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *adminResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) AccountRoles(ctx context.Context, accountID string, accountType *cato_models.AccountType,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AccountRoles, error) {
	return c.resource.client.Catov2.AccountRoles(ctx, accountID, accountType, interceptors...)
}

// resolveManagedRoleID requires an exact, unambiguous name within the admin's account.
func resolveManagedRoleID(ctx context.Context, client ResourceClient, accountID, name string) (string, error) {
	response, err := client.AccountRoles(ctx, accountID, nil)
	if err != nil {
		return "", err
	}
	roles := response.GetAccountRoles()
	if roles == nil || roles.GetItems() == nil {
		return "", fmt.Errorf("API returned no account roles")
	}
	var id string
	for _, role := range roles.GetItems() {
		if role == nil || role.GetName() != name || role.GetID() == "" {
			continue
		}
		if id != "" && id != role.GetID() {
			return "", fmt.Errorf("managed role name %q is ambiguous; specify an explicit ID", name)
		}
		id = role.GetID()
	}
	if id == "" {
		return "", fmt.Errorf("managed role %q not found in account %s", name, accountID)
	}
	return id, nil
}
