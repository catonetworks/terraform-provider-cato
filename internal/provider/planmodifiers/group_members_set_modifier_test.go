package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func groupMember(kind string, name, id types.String) types.Object {
	return objectWithValues(map[string]attr.Value{"type": types.StringValue(kind), "name": name, "id": id})
}
func TestGroupMembersSetModifier(t *testing.T) {
	t.Parallel()
	a := groupMember("USER", types.StringValue("Alice"), types.StringValue("id-a"))
	b := groupMember("GROUP", types.StringValue("team"), types.StringValue("id-b"))
	typ := a.Type(context.Background())
	set := func(v ...attr.Value) types.Set { return types.SetValueMust(typ, v) }
	renamed := groupMember("USER", types.StringValue("Alice renamed"), types.StringValue("id-a"))
	byName := groupMember("GROUP", types.StringValue("team"), types.StringUnknown())
	newMember := groupMember("USER", types.StringValue("team"), types.StringUnknown())
	null := types.SetNull(typ)
	unknown := types.SetUnknown(typ)
	for _, tc := range []struct {
		name                string
		config, state, plan types.Set
	}{
		{"reordered and renamed IDs", set(renamed, b), set(a, b), set(b, renamed)},
		{"name-only member", set(byName), set(a, b), set(byName)},
		{"same name different type", set(newMember), set(b), set(newMember)},
		{"empty", set(), set(a), set()}, {"null config", null, set(a), null}, {"unknown config", unknown, set(a), unknown},
		{"null state", set(a), null, set(a)}, {"unknown state", set(a), unknown, set(a)},
		{"null plan", set(a), set(a), null}, {"unknown plan", set(a), set(a), unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.SetResponse{PlanValue: tc.plan}
			GroupMembersSetModifier().PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}, resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, tc.plan.Equal(resp.PlanValue), "configured members must not be overwritten by stale state")
		})
	}
}
