package user

import (
	"context"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient contains the SDK operations needed to manage a local user.
type ResourceClient interface {
	UserCreateUser(
		ctx context.Context,
		accountID string,
		createUserInput cato_models.CreateUserInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.UserCreateUser, error)
	UserUpdateUser(
		ctx context.Context,
		accountID string,
		updateUserInput cato_models.UpdateUserInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.UserUpdateUser, error)
	UserDeleteUser(
		ctx context.Context,
		accountID string,
		deleteUserInput cato_models.DeleteUserInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.UserDeleteUser, error)
	UserUserList(
		ctx context.Context,
		accountID string,
		userListInput cato_models.UserListInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.UserUserList, error)
	UserEnableUser(
		ctx context.Context,
		accountID string,
		enableUserInput cato_models.EnableUserInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.UserEnableUser, error)
	UserDisableUser(
		ctx context.Context,
		accountID string,
		disableUserInput cato_models.DisableUserInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.UserDisableUser, error)
}

type sdkClient struct{ client *catoclient.Service }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) UserCreateUser(
	ctx context.Context,
	accountID string,
	createUserInput cato_models.CreateUserInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.UserCreateUser, error) {
	return c.client.V2().UserCreateUser(
		ctx,
		accountID,
		createUserInput,
		interceptors...,
	)
}
func (c *sdkClient) UserUpdateUser(
	ctx context.Context,
	accountID string,
	updateUserInput cato_models.UpdateUserInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.UserUpdateUser, error) {
	// SDK pointer fields use omitempty. This resource owns all three optional
	// attributes, so nil means explicit clearing on update rather than omission.
	var result cato.UserUpdateUser
	err := c.client.V2().Client.Post(ctx, "userUpdateUser", cato.UserUpdateUserDocument, &result, map[string]any{
		"accountId": accountID,
		"updateUserInput": map[string]any{
			"id": updateUserInput.ID, "email": updateUserInput.Email, "firstName": updateUserInput.FirstName, "lastName": updateUserInput.LastName,
			"department": updateUserInput.Department, "jobTitle": updateUserInput.JobTitle, "phoneNumber": updateUserInput.PhoneNumber,
		},
	}, interceptors...)
	if err != nil && !c.client.V2().Client.ParseDataWhenErrors {
		return nil, err
	}
	return &result, err
}
func (c *sdkClient) UserDeleteUser(
	ctx context.Context,
	accountID string,
	deleteUserInput cato_models.DeleteUserInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.UserDeleteUser, error) {
	return c.client.V2().UserDeleteUser(
		ctx,
		accountID,
		deleteUserInput,
		interceptors...,
	)
}
func (c *sdkClient) UserUserList(
	ctx context.Context,
	accountID string,
	userListInput cato_models.UserListInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.UserUserList, error) {
	return c.client.V2().UserUserList(
		ctx,
		accountID,
		userListInput,
		interceptors...,
	)
}
func (c *sdkClient) UserEnableUser(
	ctx context.Context,
	accountID string,
	enableUserInput cato_models.EnableUserInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.UserEnableUser, error) {
	return c.client.V2().UserEnableUser(
		ctx,
		accountID,
		enableUserInput,
		interceptors...,
	)
}
func (c *sdkClient) UserDisableUser(
	ctx context.Context,
	accountID string,
	disableUserInput cato_models.DisableUserInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.UserDisableUser, error) {
	return c.client.V2().UserDisableUser(
		ctx,
		accountID,
		disableUserInput,
		interceptors...,
	)
}

const userPageSize = 100

func (r *userResource) findUser(ctx context.Context, id string) (remoteUser, error) {
	seen := make(map[string]bool)
	for offset := int64(0); ; {
		result, err := r.getClient().UserUserList(ctx, r.client.AccountId, cato_models.UserListInput{
			Filter: &cato_models.UserFilterInput{ID: []*cato_models.IDFilterInput{{Eq: &id}}},
			Paging: &cato_models.PagingInput{From: offset, Limit: userPageSize},
		})
		if err != nil {
			return nil, err
		}
		page := result.GetUser().GetUserList()
		if page == nil || page.GetItems() == nil {
			return nil, fmt.Errorf("API did not return a complete user list")
		}
		for _, item := range page.GetItems() {
			if item.GetID() == "" {
				return nil, fmt.Errorf("API returned a user without an ID")
			}
			if item.GetID() == id {
				return item, nil
			}
			if seen[item.GetID()] {
				return nil, fmt.Errorf("API user pagination did not advance")
			}
			seen[item.GetID()] = true
		}
		offset += int64(len(page.GetItems()))
		if offset >= page.GetPageInfo().GetTotal() {
			return nil, nil
		}
		if len(page.GetItems()) == 0 {
			return nil, fmt.Errorf("API returned an incomplete user page")
		}
	}
}
