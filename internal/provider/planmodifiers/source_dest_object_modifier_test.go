package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestSourceDestObjectModifier(t *testing.T) {
	t.Parallel()
	a := reference(types.StringValue("host-a"), types.StringValue("id-a"))
	b := reference(types.StringValue("host-b"), types.StringValue("id-b"))
	unresolved := func(name string) types.Object { return reference(types.StringValue(name), types.StringUnknown()) }
	object := func(set types.Set) types.Object {
		return objectWithValues(map[string]attr.Value{"host": set, "ip": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.1")})})
	}
	state := object(references(a, b))
	plan := object(references(unresolved("host-b"), unresolved("host-a")))
	null := types.ObjectNull(state.AttributeTypes(context.Background()))
	unknown := types.ObjectUnknown(state.AttributeTypes(context.Background()))
	for _, tc := range []struct {
		name                      string
		config, state, plan, want types.Object
	}{
		{"match by name after reordering", plan, state, plan, state},
		{"changed name remains unresolved", object(references(unresolved("host-c"))), state, object(references(unresolved("host-c"))), object(references(unresolved("host-c")))},
		{"no state", plan, null, plan, plan}, {"unknown state", plan, unknown, plan, plan},
		{"null config", null, state, plan, plan}, {"unknown config", unknown, state, plan, plan},
		{"null plan", plan, state, null, null}, {"unknown plan", plan, state, unknown, unknown},
		{"empty set stays empty", object(references()), state, object(references()), object(references())},
		{"null nested set stays null", object(types.SetNull(idname.RefObjectType)), state, object(types.SetNull(idname.RefObjectType)), object(types.SetNull(idname.RefObjectType))},
		{"unknown nested set stays unknown", object(types.SetUnknown(idname.RefObjectType)), state, object(types.SetUnknown(idname.RefObjectType)), object(types.SetUnknown(idname.RefObjectType))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.ObjectResponse{PlanValue: tc.plan}
			SourceDestObjectModifier().PlanModifyObject(context.Background(), planmodifier.ObjectRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}, resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, tc.want.Equal(resp.PlanValue), "want %s, got %s", tc.want, resp.PlanValue)
		})
	}
}
