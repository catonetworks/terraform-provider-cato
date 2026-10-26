package iam

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPermissionsInput(t *testing.T) {
	ctx := context.Background()
	permission := func(resource, action string) attr.Value {
		return types.ObjectValueMust(PermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue(resource), "action": types.StringValue(action)})
	}
	for _, test := range []struct {
		name    string
		value   types.Set
		invalid bool
	}{
		{"empty", types.SetValueMust(PermissionObjectType, []attr.Value{}), false},
		{"valid", types.SetValueMust(PermissionObjectType, []attr.Value{permission("Sites", "VIEW")}), false},
		{"duplicate", types.SetValueMust(PermissionObjectType, []attr.Value{permission("Sites", "VIEW"), permission("Sites", "EDIT")}), true},
		{"unknown", types.SetUnknown(PermissionObjectType), true},
		{"null", types.SetNull(PermissionObjectType), true},
		{"invalid action", types.SetValueMust(PermissionObjectType, []attr.Value{permission("Sites", "NONE")}), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			input, d := PermissionsInput(ctx, test.value)
			require.Equal(t, test.invalid, d.HasError())
			if !test.invalid {
				require.NotNil(t, input)
			}
		})
	}
}
func TestRoleStateRejectsMalformedResponse(t *testing.T) {
	_, d := RoleState(context.Background(), "123", nil)
	require.True(t, d.HasError())
	_, d = RoleState(context.Background(), "123", &RoleRecord{ID: "r", Name: "test"})
	require.True(t, d.HasError())
}
func TestAccountAndCompositeID(t *testing.T) {
	account, id, err := ParseID("123:role")
	require.NoError(t, err)
	require.Equal(t, "123", account)
	require.Equal(t, "role", id)
	for _, value := range []string{"", ":a", "a:", "a:b:c", " a:b"} {
		_, _, err = ParseID(value)
		require.Error(t, err)
	}
	_, err = Account(types.StringUnknown(), "123")
	require.Error(t, err)
	account, err = Account(types.StringNull(), "123")
	require.NoError(t, err)
	require.Equal(t, "123", account)
}
