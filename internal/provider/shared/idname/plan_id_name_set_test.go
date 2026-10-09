package idname

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIDNameSetModifier(t *testing.T) {
	refType := types.ObjectType{AttrTypes: ModelTypes}
	ref := func(name, id types.String) types.Object {
		return types.ObjectValueMust(refType.AttrTypes, map[string]attr.Value{"name": name, "id": id})
	}
	set := func(refs ...attr.Value) types.Set { return types.SetValueMust(refType, refs) }
	a := ref(types.StringValue("group-a"), types.StringValue("id-a"))
	b := ref(types.StringValue("group-b"), types.StringValue("id-b"))
	tests := []struct {
		name                      string
		config, plan, state, want types.Set
	}{
		{"reordered names retain matching IDs", set(ref(types.StringValue("group-b"), types.StringNull()), ref(types.StringValue("group-a"), types.StringNull())), set(ref(types.StringValue("group-b"), types.StringUnknown()), ref(types.StringValue("group-a"), types.StringUnknown())), set(a, b), set(b, a)},
		{"changed name leaves ID unknown", set(ref(types.StringValue("group-b"), types.StringNull())), set(ref(types.StringValue("group-b"), types.StringUnknown())), set(a), set(ref(types.StringValue("group-b"), types.StringUnknown()))},
		{"changed ID leaves name unknown", set(ref(types.StringNull(), types.StringValue("id-b"))), set(ref(types.StringUnknown(), types.StringValue("id-b"))), set(a), set(ref(types.StringUnknown(), types.StringValue("id-b")))},
		{"ID-only references retain matching names", set(ref(types.StringNull(), types.StringValue("id-b")), ref(types.StringNull(), types.StringValue("id-a"))), set(ref(types.StringUnknown(), types.StringValue("id-b")), ref(types.StringUnknown(), types.StringValue("id-a"))), set(a, b), set(b, a)},
		{"first apply keeps unresolved ID", set(ref(types.StringValue("group-a"), types.StringNull())), set(ref(types.StringValue("group-a"), types.StringUnknown())), types.SetNull(refType), set(ref(types.StringValue("group-a"), types.StringUnknown()))},
		{"removed groups stay null", types.SetNull(refType), types.SetNull(refType), set(a), types.SetNull(refType)},
		{"unknown configuration stays unknown", types.SetUnknown(refType), types.SetUnknown(refType), set(a), types.SetUnknown(refType)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := planmodifier.SetResponse{PlanValue: tc.plan}
			SetPlanModifier().PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: tc.config, PlanValue: tc.plan, StateValue: tc.state}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if !resp.PlanValue.Equal(tc.want) {
				t.Fatalf("got %s, want %s", resp.PlanValue, tc.want)
			}
		})
	}
}
