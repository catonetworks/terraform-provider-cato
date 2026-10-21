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

func wanException(name string, ref types.Object, ip types.List, country types.Set, deviceAttrs types.Object) types.Object {
	nested := objectWithValues(map[string]attr.Value{"host": references(ref), "ip": ip})
	return objectWithValues(map[string]attr.Value{
		"name": types.StringValue(name), "source": nested, "destination": nested, "application": objectWithValues(map[string]attr.Value{"application": references(ref)}), "service": objectWithValues(map[string]attr.Value{"standard": references(ref)}),
		"country": country, "device": country, "device_attributes": deviceAttrs,
	})
}
func TestWanExceptionsSetModifierCorrelation(t *testing.T) {
	t.Parallel()
	known := reference(types.StringValue("host-a"), types.StringValue("id-a"))
	unknownID := reference(types.StringValue("host-a"), types.StringUnknown())
	byID := reference(types.StringUnknown(), types.StringValue("id-a"))
	ip := types.ListNull(types.StringType)
	country := references()
	attrs := objectWithValues(map[string]attr.Value{"category": types.ListNull(types.StringType)})
	nullAttrs := types.ObjectNull(attrs.AttributeTypes(context.Background()))
	prior := wanException("exception-a", known, ip, types.SetNull(idname.RefObjectType), nullAttrs)
	planned := wanException("exception-a", unknownID, ip, country, attrs)
	want := wanException("exception-a", known, ip, country, nullAttrs)
	typ := prior.Type(context.Background())
	set := func(v ...attr.Value) types.Set { return types.SetValueMust(typ, v) }
	for _, tc := range []struct {
		name       string
		plan, want types.Object
	}{
		{"matching names preserve reference IDs and empty sets", planned, want},
		{"ID selector recovers computed name", wanException("exception-a", byID, ip, country, attrs), want},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.SetResponse{PlanValue: set(tc.plan)}
			WanExceptionsSetModifier().PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: set(tc.plan), PlanValue: set(tc.plan), StateValue: set(prior)}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.True(t, set(tc.want).Equal(resp.PlanValue), "want %s, got %s", set(tc.want), resp.PlanValue)
		})
	}
}
func TestWanExceptionsSetModifierNewException(t *testing.T) {
	t.Parallel()
	ref := reference(types.StringValue("host-a"), types.StringUnknown())
	resolved := reference(types.StringValue("host-a"), types.StringNull())
	country := references()
	deviceAttrs := objectWithValues(map[string]attr.Value{"category": types.ListUnknown(types.StringType)})
	resolvedAttrs := objectWithValues(map[string]attr.Value{"category": types.ListNull(types.StringType)})
	plan := wanException("new", ref, types.ListUnknown(types.StringType), country, deviceAttrs)
	want := wanException("new", resolved, types.ListNull(types.StringType), country, resolvedAttrs)
	typ := plan.Type(context.Background())
	set := func(v ...attr.Value) types.Set { return types.SetValueMust(typ, v) }
	other := wanException("other", reference(types.StringValue("other"), types.StringValue("other-id")), types.ListNull(types.StringType), country, resolvedAttrs)
	for name, state := range map[string]types.Set{"first apply": types.SetNull(typ), "empty state": set(), "new exception alongside existing": set(other), "unknown state": types.SetUnknown(typ)} {
		t.Run(name, func(t *testing.T) {
			resp := &planmodifier.SetResponse{PlanValue: set(plan)}
			WanExceptionsSetModifier().PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: set(plan), PlanValue: set(plan), StateValue: state}, resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, set(want).Equal(resp.PlanValue), "want %s, got %s", set(want), resp.PlanValue)
		})
	}
}
func TestWanExceptionsSetModifierUnresolvedAndEmptySets(t *testing.T) {
	t.Parallel()
	typ := types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType}}
	known := types.SetValueMust(typ, []attr.Value{objectWithValues(map[string]attr.Value{"name": types.StringValue("existing")})})
	null, unknown, empty := types.SetNull(typ), types.SetUnknown(typ), types.SetValueMust(typ, []attr.Value{})
	for _, tc := range []struct {
		name                string
		config, plan, state types.Set
	}{
		{"null config", null, known, known}, {"unknown config", unknown, known, known}, {"null plan", known, null, known}, {"unknown plan", known, unknown, known}, {"empty plan", empty, empty, known},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.SetResponse{PlanValue: tc.plan}
			WanExceptionsSetModifier().PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: tc.config, PlanValue: tc.plan, StateValue: tc.state}, resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, tc.plan.Equal(resp.PlanValue))
		})
	}
}
