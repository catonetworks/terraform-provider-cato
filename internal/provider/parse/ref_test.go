package parse

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestSchemaNameID(t *testing.T) {
	t.Run("without prefix", func(t *testing.T) {
		s := SchemaNameID("")

		if len(s) != 2 {
			t.Fatalf("expected 2 attributes, got %d", len(s))
		}
		if got := s["name"].GetDescription(); got != "name" {
			t.Fatalf("expected name description %q, got %q", "name", got)
		}
		if got := s["id"].GetDescription(); got != "ID" {
			t.Fatalf("expected id description %q, got %q", "ID", got)
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		s := SchemaNameID("group")

		if got := s["name"].GetDescription(); got != "group name" {
			t.Fatalf("expected name description %q, got %q", "group name", got)
		}
		if got := s["id"].GetDescription(); got != "group ID" {
			t.Fatalf("expected id description %q, got %q", "group ID", got)
		}
	})
}

func TestNormalizeDateTime(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no fractional seconds",
			in:   "2024-02-03T04:05:06",
			want: "2024-02-03T04:05:06Z",
		},
		{
			name: "with fractional seconds",
			in:   "2024-02-03T04:05:06.987654",
			want: "2024-02-03T04:05:06Z",
		},
		{
			name: "with timezone suffix still normalizes prefix",
			in:   "2024-02-03T04:05:06+02:00",
			want: "2024-02-03T04:05:06Z",
		},
		{
			name: "non matching text returns unchanged",
			in:   "not-a-datetime",
			want: "not-a-datetime",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeDateTime(tc.in); got != tc.want {
				t.Fatalf("unexpected normalized value\nwant: %q\ngot:  %q", tc.want, got)
			}
		})
	}
}

func TestNormalizeDateTimePtr(t *testing.T) {
	t.Run("nil stays nil", func(t *testing.T) {
		if got := NormalizeDateTimePtr(nil); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("value is normalized and returned", func(t *testing.T) {
		in := "2024-02-03T04:05:06.123"
		got := NormalizeDateTimePtr(&in)
		if got == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *got != "2024-02-03T04:05:06Z" {
			t.Fatalf("unexpected normalized pointer value: %q", *got)
		}
	})
}

func TestIDNameModifierDescriptions(t *testing.T) {
	const descr = "Once set, the value of this attribute in state will not change."
	m := idNamePlanModifier{}
	ctx := context.Background()
	want := descr

	if got := m.Description(ctx); got != want {
		t.Fatalf("unexpected description\nwant: %q\ngot:  %q", want, got)
	}
	if got := m.MarkdownDescription(ctx); got != want {
		t.Fatalf("unexpected markdown description\nwant: %q\ngot:  %q", want, got)
	}
}

func TestIDNamePlanModifierPlanModifyObject(t *testing.T) {
	t.Run("unknown config value is ignored", func(t *testing.T) {
		m := idNamePlanModifier{}
		resp := &planmodifier.ObjectResponse{}

		m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			Path:        path.Root("ref"),
			ConfigValue: types.ObjectUnknown(IDNameRefModelTypes),
			StateValue:  types.ObjectNull(IDNameRefModelTypes),
		}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
		}
		if !resp.PlanValue.IsNull() {
			t.Fatalf("expected no plan value change, got %#v", resp.PlanValue)
		}
	})

	t.Run("requires exactly one of name or id", func(t *testing.T) {
		t.Run("none configured", func(t *testing.T) {
			m := idNamePlanModifier{}
			resp := &planmodifier.ObjectResponse{}

			m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
				Path:        path.Root("ref"),
				ConfigValue: newRefObject(t, types.StringNull(), types.StringNull()),
				StateValue:  types.ObjectNull(IDNameRefModelTypes),
			}, resp)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected diagnostics")
			}
		})

		t.Run("both configured", func(t *testing.T) {
			m := idNamePlanModifier{}
			resp := &planmodifier.ObjectResponse{}

			m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
				Path:        path.Root("ref"),
				ConfigValue: newRefObject(t, types.StringValue("name-1"), types.StringValue("id-1")),
				StateValue:  types.ObjectNull(IDNameRefModelTypes),
			}, resp)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected diagnostics")
			}
		})
	})

	t.Run("name configured and unchanged keeps state", func(t *testing.T) {
		m := idNamePlanModifier{}
		resp := &planmodifier.ObjectResponse{}
		state := newRefObject(t, types.StringValue("n1"), types.StringValue("id-1"))

		m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			Path:        path.Root("ref"),
			ConfigValue: newRefObject(t, types.StringValue("n1"), types.StringNull()),
			StateValue:  state,
		}, resp)

		assertObjectsEqual(t, state, resp.PlanValue)
	})

	t.Run("name configured and changed sets unknown id", func(t *testing.T) {
		m := idNamePlanModifier{}
		resp := &planmodifier.ObjectResponse{}

		m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			Path:        path.Root("ref"),
			ConfigValue: newRefObject(t, types.StringValue("n2"), types.StringNull()),
			StateValue:  newRefObject(t, types.StringValue("n1"), types.StringValue("id-1")),
		}, resp)

		plan := mustAsRefModel(t, resp.PlanValue)
		if plan.Name.ValueString() != "n2" {
			t.Fatalf("expected name %q, got %q", "n2", plan.Name.ValueString())
		}
		if !plan.ID.IsUnknown() {
			t.Fatalf("expected id to be unknown, got %#v", plan.ID)
		}
	})

	t.Run("id configured and unchanged keeps state", func(t *testing.T) {
		m := idNamePlanModifier{}
		resp := &planmodifier.ObjectResponse{}
		state := newRefObject(t, types.StringValue("n1"), types.StringValue("id-1"))

		m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			Path:        path.Root("ref"),
			ConfigValue: newRefObject(t, types.StringNull(), types.StringValue("id-1")),
			StateValue:  state,
		}, resp)

		assertObjectsEqual(t, state, resp.PlanValue)
	})

	t.Run("id configured and changed sets unknown name", func(t *testing.T) {
		m := idNamePlanModifier{}
		resp := &planmodifier.ObjectResponse{}

		m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			Path:        path.Root("ref"),
			ConfigValue: newRefObject(t, types.StringNull(), types.StringValue("id-2")),
			StateValue:  newRefObject(t, types.StringValue("n1"), types.StringValue("id-1")),
		}, resp)

		plan := mustAsRefModel(t, resp.PlanValue)
		if plan.ID.ValueString() != "id-2" {
			t.Fatalf("expected id %q, got %q", "id-2", plan.ID.ValueString())
		}
		if !plan.Name.IsUnknown() {
			t.Fatalf("expected name to be unknown, got %#v", plan.Name)
		}
	})

	t.Run("null config with state sets null plan", func(t *testing.T) {
		m := idNamePlanModifier{}
		resp := &planmodifier.ObjectResponse{}

		m.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			Path:        path.Root("ref"),
			ConfigValue: types.ObjectNull(IDNameRefModelTypes),
			StateValue:  newRefObject(t, types.StringValue("n1"), types.StringValue("id-1")),
		}, resp)

		if !resp.PlanValue.IsNull() {
			t.Fatalf("expected null plan value, got %#v", resp.PlanValue)
		}
	})
}

func newRefObject(t *testing.T, name, id types.String) types.Object {
	t.Helper()

	obj, diags := types.ObjectValue(IDNameRefModelTypes, map[string]attr.Value{
		"name": name,
		"id":   id,
	})
	if diags.HasError() {
		t.Fatalf("unexpected object creation diagnostics: %+v", diags)
	}
	return obj
}

func mustAsRefModel(t *testing.T, v types.Object) IDNameRefModel {
	t.Helper()

	var out IDNameRefModel
	diags := v.As(context.Background(), &out, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected object conversion diagnostics: %+v", diags)
	}
	return out
}

func assertObjectsEqual(t *testing.T, want, got types.Object) {
	t.Helper()
	if !want.Equal(got) {
		t.Fatalf("objects differ\nwant: %#v\ngot:  %#v", want, got)
	}
}
