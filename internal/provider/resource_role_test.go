package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestRolePermissionInput(t *testing.T) {
	ctx := context.Background()
	for _, action := range []string{"VIEW", "EDIT", "NONE"} {
		t.Run(action, func(t *testing.T) {
			permissions := types.SetValueMust(rolePermissionObjectType, []attr.Value{types.ObjectValueMust(rolePermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue("Sites"), "action": types.StringValue(action)})})
			input, diags := rolePermissionsInput(ctx, permissions)
			if action == "NONE" {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError())
			require.Len(t, input, 1)
			require.Equal(t, "Sites", input[0].Resource)
			require.Equal(t, action, string(input[0].Action))
		})
	}
	for _, v := range []types.Set{types.SetNull(rolePermissionObjectType), types.SetUnknown(rolePermissionObjectType)} {
		_, diags := rolePermissionsInput(ctx, v)
		require.True(t, diags.HasError())
	}
}

func TestRoleImport(t *testing.T) {
	ctx := context.Background()
	r := &roleResource{}
	schemaResp := resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	for _, id := range []string{"123:42", "42", "123:", ":42", "123:42:7"} {
		t.Run(id, func(t *testing.T) {
			state := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, state.Set(ctx, emptyRole()).HasError())
			resp := resource.ImportStateResponse{State: state}
			r.ImportState(ctx, resource.ImportStateRequest{ID: id}, &resp)
			if id != "123:42" {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError())
			var got Role
			require.False(t, resp.State.Get(ctx, &got).HasError())
			require.Equal(t, "123", got.AccountID.ValueString())
			require.Equal(t, "42", got.RoleID.ValueString())
		})
	}
}

func emptyRole() Role {
	return Role{ID: types.StringNull(), AccountID: types.StringNull(), RoleID: types.StringNull(), Name: types.StringNull(), Description: types.StringNull(), Permissions: types.SetNull(rolePermissionObjectType), Predefined: types.BoolNull(), AccountType: types.StringNull(), IsUsedOnExternalAccess: types.BoolNull()}
}
