package oldidname

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestParseNameIDListsDeduplicateReferences(t *testing.T) {
	type reference struct{ ID, Name string }
	a := &reference{ID: "suspected_malware", Name: "Malware"}
	b := &reference{ID: "food_drinks_tobacco", Name: "Alcohol and Tobacco"}
	ref := func(id, name string) attr.Value {
		return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id": types.StringValue(id), "name": types.StringValue(name),
		})
	}
	for _, parser := range []struct {
		name  string
		parse func(context.Context, []*reference, string) types.Set
		empty types.Set
	}{
		{"null for empty", ParseNameIDList[*reference], types.SetNull(idname.RefObjectType)},
		{"set for empty", ParseNameIDListOrEmptySet[*reference], types.SetValueMust(idname.RefObjectType, nil)},
	} {
		t.Run(parser.name, func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				items []*reference
				want  types.Set
			}{
				{"nil", nil, parser.empty},
				{"empty", []*reference{}, parser.empty},
				{"unique", []*reference{a, b}, types.SetValueMust(idname.RefObjectType, []attr.Value{ref(a.ID, a.Name), ref(b.ID, b.Name)})},
				{"duplicates", []*reference{a, b, b, a}, types.SetValueMust(idname.RefObjectType, []attr.Value{ref(a.ID, a.Name), ref(b.ID, b.Name)})},
				{"nil among duplicates", []*reference{nil, a, a}, types.SetValueMust(idname.RefObjectType, []attr.Value{ref(a.ID, a.Name)})},
				{"same name different IDs", []*reference{a, {ID: "other", Name: a.Name}}, types.SetValueMust(idname.RefObjectType, []attr.Value{ref(a.ID, a.Name), ref("other", a.Name)})},
			} {
				t.Run(tc.name, func(t *testing.T) {
					got := parser.parse(context.Background(), tc.items, "test.references")
					if !got.Equal(tc.want) {
						t.Fatalf("got %s, want %s", got, tc.want)
					}
				})
			}
		})
	}
}
