package provider

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

// --- correlateIfwSetElement tests ---

func TestCorrelateIfwSetElement_NameUnchanged(t *testing.T) {
	// Name-based reference, same name in plan and API.
	// After plan modifier backfills id from state, plan has id="500000004", name="app-cursor".
	// API returns id="500000004", name="app-cursor".
	// Result: both values preserved from plan (they match API anyway).
	plan := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("500000004"),
		"name": types.StringValue("app-cursor"),
	})
	response := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("500000004"),
		"name": types.StringValue("app-cursor"),
	})

	result := correlateIfwSetElement(context.Background(), plan, response)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "500000004" {
		t.Errorf("expected id=%q, got %q", "500000004", id.ValueString())
	}
	if name.ValueString() != "app-cursor" {
		t.Errorf("expected name=%q, got %q", "app-cursor", name.ValueString())
	}
}

func TestCorrelateIfwSetElement_NameChanged(t *testing.T) {
	// Name-based reference, name changed from "Jeff" to "test user_1".
	// Plan has name="test user_1", id=unknown.
	// API returns id="99", name="test user_1".
	// Result: use API's id, plan's name.
	plan := mustObj(t, map[string]attr.Value{
		"id":   types.StringUnknown(),
		"name": types.StringValue("test user_1"),
	})
	response := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("99"),
		"name": types.StringValue("test user_1"),
	})

	result := correlateIfwSetElement(context.Background(), plan, response)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "99" {
		t.Errorf("expected id=%q (from API), got %q", "99", id.ValueString())
	}
	if name.ValueString() != "test user_1" {
		t.Errorf("expected name=%q (from plan), got %q", "test user_1", name.ValueString())
	}
}

func TestCorrelateIfwSetElement_IdOnlyReference(t *testing.T) {
	// Id-based reference: user sets id="4", name is unknown (UseStateForUnknown removed from
	// exception name fields; IfwExceptionsSetModifier handles backfilling instead).
	// API returns id="4", name="test user_1".
	// Result: preserve plan id="4", use API name (plan name was unknown → accepts any value).
	plan := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),
		"name": types.StringUnknown(),
	})
	response := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),
		"name": types.StringValue("test user_1"),
	})

	result := correlateIfwSetElement(context.Background(), plan, response)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "4" {
		t.Errorf("expected id=%q (from plan), got %q", "4", id.ValueString())
	}
	if name.ValueString() != "test user_1" {
		t.Errorf("expected name=%q (from API, plan was unknown), got %q", "test user_1", name.ValueString())
	}
}

func TestCorrelateIfwSetElement_IdOnlyReference_NameNull(t *testing.T) {
	// Id-based reference: user sets id="4", name is null in plan (no leak).
	// API returns id="4", name="resolved user".
	// Result: preserve plan's id="4", keep name=null.
	plan := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),
		"name": types.StringNull(),
	})
	response := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("4"),
		"name": types.StringValue("resolved user"),
	})

	result := correlateIfwSetElement(context.Background(), plan, response)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "4" {
		t.Errorf("expected id=%q (from plan), got %q", "4", id.ValueString())
	}
	if !name.IsNull() {
		t.Errorf("expected name=null (id-only ref), got %q", name.ValueString())
	}
}

func TestCorrelateIfwSetElement_NewElement_FirstApply(t *testing.T) {
	// First apply for a new element: plan has id=unknown, name="new-group".
	// API returns id="777", name="new-group".
	// Result: use API's id, plan's name.
	plan := mustObj(t, map[string]attr.Value{
		"id":   types.StringUnknown(),
		"name": types.StringValue("new-group"),
	})
	response := mustObj(t, map[string]attr.Value{
		"id":   types.StringValue("777"),
		"name": types.StringValue("new-group"),
	})

	result := correlateIfwSetElement(context.Background(), plan, response)

	id := getAttrStr(t, result, "id")
	name := getAttrStr(t, result, "name")
	if id.ValueString() != "777" {
		t.Errorf("expected id=%q (from API), got %q", "777", id.ValueString())
	}
	if name.ValueString() != "new-group" {
		t.Errorf("expected name=%q (from plan), got %q", "new-group", name.ValueString())
	}
}
