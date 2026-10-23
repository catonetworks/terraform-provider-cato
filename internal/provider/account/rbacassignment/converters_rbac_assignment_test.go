package rbacassignment

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAssignmentInput(t *testing.T) {
	plan := Model{UserID: types.StringValue("u"), GroupID: types.StringNull(), RoleIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("r")}), AllowedEntities: types.SetValueMust(entityObjectType, []attr.Value{}), Description: types.StringValue(""), AllowViewingUnobfuscatedData: types.BoolNull()}
	input, d := createInput(context.Background(), plan)
	require.False(t, d.HasError())
	require.Len(t, input.Users, 1)
	require.Empty(t, input.Groups)
	require.NotNil(t, input.Groups)
	require.Nil(t, input.AllowViewingUnobfuscatedData)
	plan.GroupID = types.StringValue("g")
	_, d = createInput(context.Background(), plan)
	require.True(t, d.HasError())
	plan.GroupID = types.StringNull()
	plan.RoleIDs = types.SetUnknown(types.StringType)
	_, d = createInput(context.Background(), plan)
	require.True(t, d.HasError())
}
func TestObfuscationAndAccountValidation(t *testing.T) {
	record := &assignmentRecord{ID: "a", Account: "123", User: "u", Kind: "ADMIN_USER", Roles: []string{"r"}, Entities: []entityRecord{}, AllSites: "NONE", AllUsers: "NONE", Obfuscation: "DISABLED"}
	previous := Model{UserID: types.StringValue("u"), GroupID: types.StringNull(), AllowViewingUnobfuscatedData: types.BoolValue(true)}
	state, d := assignmentState(context.Background(), "123", record, &previous)
	require.False(t, d.HasError())
	require.True(t, state.AllowViewingUnobfuscatedData.ValueBool())
	_, d = assignmentState(context.Background(), "456", record, &previous)
	require.True(t, d.HasError())
	record.Kind = "ADMIN"
	_, d = assignmentState(context.Background(), "123", record, &previous)
	require.True(t, d.HasError())
}
