package groupmembers

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// getGroupByName fetches a group ID by name
func (r *groupMembersResource) getGroupByName(ctx context.Context, groupName string) (string, error) {
	var groupID string

	// Read group metadata using GroupsList
	groupListInput := &cato_models.GroupListInput{
		Paging: &cato_models.PagingInput{
			Limit: 1000,
			From:  0,
		},
		Sort: &cato_models.GroupListSortInput{},
	}

	groupListResult, err := r.resourceAPI().GroupsList(ctx, groupListInput, r.client.AccountId)
	tflog.Debug(ctx, "getGroupByName.GroupsList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(groupListResult),
	})
	if err != nil {
		return groupID, fmt.Errorf("GroupsList API error: %w", err)
	}

	// Check if the group was found
	if groupListResult.Groups.GroupList == nil || len(groupListResult.Groups.GroupList.Items) == 0 {
		return groupID, fmt.Errorf("no groups found in account")
	}

	groupPresent := false

	for _, groupItem := range groupListResult.Groups.GroupList.Items {
		if groupItem.Name == groupName {
			groupPresent = true
			groupID = groupItem.ID
			break
		}
	}

	if !groupPresent {
		return groupID, fmt.Errorf("group not found: %s", groupName)
	}

	return groupID, nil
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
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

func (r *groupMembersResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
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
type sdkClient struct{ resource *groupMembersResource }

var _ ResourceClient = (*sdkClient)(nil)
