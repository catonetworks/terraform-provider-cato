package admin

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestAdminUpdateRefreshesManagedRoleNames(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, response, wantError string
		readError                 error
	}{
		{name: "updated role name is hydrated", response: `{"admin":{"id":"admin-123","managedRoles":[{"role":{"id":"role-2","name":"Viewer"}}]}}`},
		{name: "read error", readError: errors.New("unavailable"), wantError: "Failed to read admin after update: unavailable"},
		{name: "missing admin", response: `{}`, wantError: "no admin data returned"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			client := mocks.NewAdminResourceClient(t)
			r := &adminResource{client: &catoclient.Service{AccountId: "provider-account"}, apiClient: client}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			role := func(id string, name types.String) types.Set {
				return types.SetValueMust(idname.RefObjectType, []attr.Value{
					types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue(id), "name": name}),
				})
			}
			model := Admin{
				ID: types.StringValue("managed-account:admin-123"), AdminID: types.StringValue("admin-123"),
				AccountID: types.StringValue("managed-account"), Email: types.StringValue("admin@example.test"),
				FirstName: types.StringValue("Example"), LastName: types.StringValue("Admin"),
				PasswordNeverExpires: types.BoolValue(false), MFAEnabled: types.BoolValue(true),
				ManagedRoles: role("role-1", types.StringValue("Editor")),
				ResellerRoles: types.SetNull(types.ObjectType{AttrTypes: map[string]attr.Type{
					"id": types.StringType, "name": types.StringType, "allowed_accounts": types.SetType{ElemType: types.StringType},
				}}),
			}
			state := tfsdk.State{Schema: schemaResponse.Schema}
			require.False(t, state.Set(ctx, model).HasError())
			model.ManagedRoles = role("role-2", types.StringUnknown())
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, model).HasError())
			var updateResponse cato.AdminUpdateAdmin
			require.NoError(t, json.Unmarshal([]byte(`{"admin":{"updateAdmin":{"adminID":"admin-123"}}}`), &updateResponse))
			client.EXPECT().AdminUpdateAdmin(mock.Anything, "admin-123", mock.MatchedBy(func(input models.UpdateAdminInput) bool {
				return len(input.ManagedRoles) == 1 && input.ManagedRoles[0].Role.ID == "role-2"
			}), "managed-account").Return(&updateResponse, nil).Once()
			var readResponse cato.Admin
			if tc.response != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.response), &readResponse))
			}
			client.EXPECT().Admin(mock.Anything, "managed-account", "admin-123").Return(&readResponse, tc.readError).Once()
			resp := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &resp)
			if tc.wantError != "" {
				require.True(t, resp.Diagnostics.HasError())
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.wantError)
				require.Equal(t, state.Raw, resp.State.Raw)
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.False(t, resp.State.Get(ctx, &model).HasError())
			require.True(t, role("role-2", types.StringValue("Viewer")).Equal(model.ManagedRoles), "got %s", model.ManagedRoles)
			require.Equal(t, types.StringValue("managed-account:admin-123"), model.ID)
		})
	}
}
