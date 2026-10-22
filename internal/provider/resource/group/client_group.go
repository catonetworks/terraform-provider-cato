package group

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	GroupsCreateGroup(
		ctx context.Context,
		input cato_models.CreateGroupInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.GroupsCreateGroup, error)
	GroupsDeleteGroup(
		ctx context.Context,
		input cato_models.GroupRefInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.GroupsDeleteGroup, error)
	GroupsList(
		ctx context.Context,
		groupListInput *cato_models.GroupListInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.GroupsList, error)
	GroupsMembers(
		ctx context.Context,
		groupRefInput cato_models.GroupRefInput,
		groupMembersListInput cato_models.GroupMembersListInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.GroupsMembers, error)
	GroupsUpdateGroup(
		ctx context.Context,
		input cato_models.UpdateGroupInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.GroupsUpdateGroup, error)
}

func (r *groupResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) GroupsCreateGroup(
	ctx context.Context,
	input cato_models.CreateGroupInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.GroupsCreateGroup, error) {
	return c.resource.client.Catov2.GroupsCreateGroup(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) GroupsDeleteGroup(
	ctx context.Context,
	input cato_models.GroupRefInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.GroupsDeleteGroup, error) {
	return c.resource.client.Catov2.GroupsDeleteGroup(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) GroupsList(
	ctx context.Context,
	groupListInput *cato_models.GroupListInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.GroupsList, error) {
	return c.resource.client.Catov2.GroupsList(ctx, groupListInput, accountID, interceptors...)
}

func (c *sdkClient) GroupsMembers(
	ctx context.Context,
	groupRefInput cato_models.GroupRefInput,
	groupMembersListInput cato_models.GroupMembersListInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.GroupsMembers, error) {
	return c.resource.client.Catov2.GroupsMembers(ctx, groupRefInput, groupMembersListInput, accountID, interceptors...)
}

func (c *sdkClient) GroupsUpdateGroup(
	ctx context.Context,
	input cato_models.UpdateGroupInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.GroupsUpdateGroup, error) {
	return c.resource.client.Catov2.GroupsUpdateGroup(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *groupResource }

var _ ResourceClient = (*sdkClient)(nil)
