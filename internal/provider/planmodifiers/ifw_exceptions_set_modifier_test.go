package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var nameIdObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	},
}

func mustObj(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(nameIdObjType.AttrTypes, attrs)
	if diags.HasError() {
		t.Fatalf("failed to create object: %s", diags.Errors())
	}
	return obj
}

func getAttrStr(t *testing.T, obj types.Object, key string) types.String {
	t.Helper()
	v, ok := obj.Attributes()[key]
	if !ok {
		t.Fatalf("attribute %q not found", key)
	}
	s, ok := v.(types.String)
	if !ok {
		t.Fatalf("attribute %q is not types.String", key)
	}
	return s
}

// --- preserveElementId tests ---

func TestPreserveElementId_NameUnchanged(t *testing.T) {
	// Name-only reference, same name as state → preserve state id
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringUnknown(), // id not in config, marked unknown
		"name": types.StringValue("app-cursor"),
	})
	state := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("500000004"),
		"name": types.StringValue("app-cursor"),
	})

	result := m.preserveElementId(context.Background(), planned, state)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "500000004" {
		t.Errorf("expected id=%q, got %q (unknown=%v, null=%v)", "500000004", id.ValueString(), id.IsUnknown(), id.IsNull())
	}
	if name.ValueString() != "app-cursor" {
		t.Errorf("expected name=%q, got %q", "app-cursor", name.ValueString())
	}
}

func TestPreserveElementId_NameChanged(t *testing.T) {
	// Name-only reference with a NEW name (no state match → preserveElementId is not
	// called for unmatched elements, but clearStaleIDForNameRefSet handles it).
	// This test verifies the "matched by name" path still works when the entity exists
	// but the test simulates an element matched to a DIFFERENT state entry.
	// In practice, findElementByName returns nil for new names, so preserveElementId
	// is never called. The id stays unknown. Test clearStaleIDForNameRefSet instead.
}

func TestPreserveElementId_IdOnlyReference(t *testing.T) {
	// User specifies id="4" only, name is null in config but UseStateForUnknown leaked
	// name="test user_1" from state. Plan has id="4", name="test user_1".
	// preserveElementId must NOT override id with state's "2".
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),           // explicitly from config
		"name": types.StringValue("test user_1"), // leaked from UseStateForUnknown
	})
	state := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("2"),
		"name": types.StringValue("test user_1"),
	})

	result := m.preserveElementId(context.Background(), planned, state)

	id := getAttrStr(t, result, "id")
	if id.ValueString() != "4" {
		t.Errorf("expected id=%q (from config), got %q — plan id was overwritten by state", "4", id.ValueString())
	}
}

func TestPreserveElementId_IdOnlyReference_NameNull(t *testing.T) {
	// User specifies id="4" only, name is null (UseStateForUnknown didn't match).
	// preserveElementId must keep id="4" and fill name from state.
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),
		"name": types.StringNull(),
	})
	state := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("2"),
		"name": types.StringValue("old user"),
	})

	result := m.preserveElementId(context.Background(), planned, state)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "4" {
		t.Errorf("expected id=%q (from config), got %q", "4", id.ValueString())
	}
	if name.ValueString() != "old user" {
		t.Errorf("expected name=%q (from state), got %q", "old user", name.ValueString())
	}
}

func TestPreserveElementId_FirstApply(t *testing.T) {
	// First apply: plan has id=unknown, name="new-user", state element found by name
	// (same name exists in state from a previous apply). Preserve state id.
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringUnknown(),
		"name": types.StringValue("new-user"),
	})
	state := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("99"),
		"name": types.StringValue("new-user"),
	})

	result := m.preserveElementId(context.Background(), planned, state)

	id := getAttrStr(t, result, "id")
	if id.ValueString() != "99" {
		t.Errorf("expected id=%q (from state), got %q (unknown=%v)", "99", id.ValueString(), id.IsUnknown())
	}
}

// --- clearStaleIDForNameRefSet tests ---

func TestClearStaleID_NewNameEntity(t *testing.T) {
	// Name-based reference, no match in state → id should become unknown
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringNull(), // no id in config, no UseStateForUnknown
		"name": types.StringValue("brand-new-user"),
	})

	result := m.clearStaleIDForNameRefSet(context.Background(), planned, "user")
	resultObj := result.(types.Object)

	id := getAttrStr(t, resultObj, "id")
	if !id.IsUnknown() {
		t.Errorf("expected id to be unknown for new entity, got value=%q null=%v", id.ValueString(), id.IsNull())
	}
}

func TestClearStaleID_IdOnlyReference_NotCleared(t *testing.T) {
	// User specified id="4" explicitly, name leaked from UseStateForUnknown.
	// clearStaleIDForNameRefSet must NOT clear the explicit id to unknown.
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),           // explicitly from config
		"name": types.StringValue("test user_1"), // leaked from UseStateForUnknown
	})

	result := m.clearStaleIDForNameRefSet(context.Background(), planned, "user")
	resultObj := result.(types.Object)

	id := getAttrStr(t, resultObj, "id")
	if id.IsUnknown() {
		t.Error("explicit id was cleared to unknown — should have been preserved")
	}
	if id.ValueString() != "4" {
		t.Errorf("expected id=%q, got %q", "4", id.ValueString())
	}
}

func TestClearStaleID_NullName_NoOp(t *testing.T) {
	// Name is null (id-only reference) → should not touch anything
	m := ifwExceptionsSetModifier{}
	planned := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),
		"name": types.StringNull(),
	})

	result := m.clearStaleIDForNameRefSet(context.Background(), planned, "user")
	resultObj := result.(types.Object)

	id := getAttrStr(t, resultObj, "id")
	if id.ValueString() != "4" {
		t.Errorf("expected id=%q unchanged, got %q", "4", id.ValueString())
	}
}
