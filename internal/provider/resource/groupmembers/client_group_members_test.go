package groupmembers

import (
	"context"
	"encoding/json"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedmembers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
)

func TestGroupMembersReadRetainsOnlyConfiguredMembers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := mocks.NewGroupMembersResourceClient(t)
	var response cato.GroupsMembers
	require.NoError(t, json.Unmarshal([]byte(`{"groups":{"group":{"members":{"items":[
 {"id":"device-1","name":"managed-device","typeGroupMemberRefTyped":"DEVICE"},
 {"id":"device-2","name":"unmanaged-device","typeGroupMemberRefTyped":"DEVICE"}
 ]}}}}`), &response))
	api.EXPECT().GroupsMembers(mock.Anything, cato_models.GroupRefInput{By: cato_models.ObjectRefByID, Input: "group-123"},
		mock.MatchedBy(func(input cato_models.GroupMembersListInput) bool {
			return input.Paging != nil && input.Paging.Limit == 1000 && input.Paging.From == 0
		}), "account-123").Return(&response, nil).Once()
	members, diags := types.SetValueFrom(ctx, sharedmembers.MemberObjectType, []sharedmembers.Member{{
		ID: types.StringValue("device-1"), Name: types.StringNull(), Type: types.StringValue("DEVICE"),
	}})
	require.False(t, diags.HasError(), "%v", diags)
	r := &groupMembersResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, state.Set(ctx, Members{ID: types.StringValue("group-123"), GroupName: types.StringValue("group"), Members: members}).HasError())
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var actual Members
	require.False(t, resp.State.Get(ctx, &actual).HasError())
	var actualMembers []sharedmembers.Member
	require.False(t, actual.Members.ElementsAs(ctx, &actualMembers, false).HasError())
	require.Len(t, actualMembers, 1)
	require.Equal(t, "device-1", actualMembers[0].ID.ValueString())
	require.Equal(t, "managed-device", actualMembers[0].Name.ValueString())
}
