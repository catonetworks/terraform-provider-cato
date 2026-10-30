package alwaysonindex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSectionDataValidatorValidateMap(t *testing.T) {
	t.Parallel()
	newSection := func(index types.Int64) types.Object {
		return types.ObjectValueMust(SectionAttrTypes, map[string]attr.Value{
			"id": types.StringUnknown(), "section_index": index,
		})
	}
	knownSection := func(index int64) types.Object { return newSection(types.Int64Value(index)) }
	for _, tc := range []struct {
		name     string
		sections map[string]attr.Value
		summary  string
		section  string
	}{
		{name: "empty"},
		{name: "single", sections: map[string]attr.Value{"A": knownSection(1)}},
		{name: "reordered", sections: map[string]attr.Value{"A": knownSection(2), "B": knownSection(1)}},
		{name: "empty name", sections: map[string]attr.Value{"": knownSection(1)}, summary: "Invalid section name"},
		{name: "empty name with unknown value", sections: map[string]attr.Value{"": types.ObjectUnknown(SectionAttrTypes)},
			summary: "Invalid section name"},
		{name: "negative", sections: map[string]attr.Value{"A": knownSection(-1)}, summary: "Invalid section index", section: "A"},
		{name: "zero", sections: map[string]attr.Value{"A": knownSection(0)}, summary: "Invalid section index", section: "A"},
		{name: "above maximum", sections: map[string]attr.Value{"A": knownSection(2)}, summary: "Invalid section index", section: "A"},
		{name: "gap", sections: map[string]attr.Value{"A": knownSection(1), "B": knownSection(3)},
			summary: "Invalid section index", section: "B"},
		{name: "duplicate", sections: map[string]attr.Value{"A": knownSection(1), "B": knownSection(1)},
			summary: "Duplicate section index"},
		{name: "unknown index", sections: map[string]attr.Value{"A": knownSection(2), "B": newSection(types.Int64Unknown())}},
		{name: "null index", sections: map[string]attr.Value{"A": newSection(types.Int64Null())}},
		{name: "unknown object", sections: map[string]attr.Value{"A": knownSection(2), "B": types.ObjectUnknown(SectionAttrTypes)}},
		{name: "null object", sections: map[string]attr.Value{"A": types.ObjectNull(SectionAttrTypes)}},
		{name: "unknown index does not hide duplicate", sections: map[string]attr.Value{
			"A": knownSection(1), "B": knownSection(1), "C": newSection(types.Int64Unknown()),
		}, summary: "Duplicate section index"},
		{name: "unknown object does not hide out of range index", sections: map[string]attr.Value{
			"A": knownSection(3), "B": types.ObjectUnknown(SectionAttrTypes),
		}, summary: "Invalid section index", section: "A"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := validator.MapRequest{
				Path: path.Root("section_data"), ConfigValue: types.MapValueMust(SectionObjectType, tc.sections),
			}
			var resp validator.MapResponse
			sectionDataValidator{}.ValidateMap(ctx, req, &resp)
			if tc.summary == "" {
				require.Empty(t, resp.Diagnostics)
				return
			}
			require.Len(t, resp.Diagnostics, 1)
			require.Equal(t, diag.SeverityError, resp.Diagnostics[0].Severity())
			require.Equal(t, tc.summary, resp.Diagnostics[0].Summary())
			attributeDiag, ok := resp.Diagnostics[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			switch {
			case tc.summary == "Invalid section name":
				require.Equal(t, req.Path.AtMapKey(""), attributeDiag.Path())
			case tc.section != "":
				require.Equal(t, req.Path.AtMapKey(tc.section).AtName("section_index"), attributeDiag.Path())
			default:
				require.Contains(t, []path.Path{
					req.Path.AtMapKey("A").AtName("section_index"), req.Path.AtMapKey("B").AtName("section_index"),
				}, attributeDiag.Path())
			}
		})
	}
}

func TestSectionDataValidatorUnknownAndNullMap(t *testing.T) {
	t.Parallel()
	for _, value := range []types.Map{types.MapNull(SectionObjectType), types.MapUnknown(SectionObjectType)} {
		existing := diag.NewWarningDiagnostic("Existing", "Keep this diagnostic")
		resp := validator.MapResponse{Diagnostics: diag.Diagnostics{existing}}
		sectionDataValidator{}.ValidateMap(ctx, validator.MapRequest{ConfigValue: value}, &resp)
		require.Equal(t, diag.Diagnostics{existing}, resp.Diagnostics)
	}
}

func TestSectionDataValidatorInvalidTypes(t *testing.T) {
	t.Parallel()
	for _, value := range []types.Map{
		types.MapValueMust(types.StringType, map[string]attr.Value{"A": types.StringValue("invalid")}),
		types.MapValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{}}, map[string]attr.Value{
			"A": types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		}),
	} {
		var resp validator.MapResponse
		sectionDataValidator{}.ValidateMap(ctx, validator.MapRequest{Path: path.Root("section_data"), ConfigValue: value}, &resp)
		require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	}
}

func TestSectionDataValidatorSchema(t *testing.T) {
	t.Parallel()
	var resp resource.SchemaResponse
	(&alwaysOnRuleBulkResource{}).Schema(ctx, resource.SchemaRequest{}, &resp)
	sections, ok := resp.Schema.Attributes["section_data"].(schema.MapNestedAttribute)
	require.True(t, ok)
	require.Contains(t, sections.Validators, sectionDataValidator{})
	v := sectionDataValidator{}
	require.NotEmpty(t, v.Description(ctx))
	require.Equal(t, v.Description(ctx), v.MarkdownDescription(ctx))
}

func TestRuleDataValidatorValidateMap(t *testing.T) {
	t.Parallel()
	newRule := func(section types.String, index types.Int64) types.Object {
		return types.ObjectValueMust(RuleAttrTypes, map[string]attr.Value{
			"id": types.StringUnknown(), "section_name": section, "index_in_section": index,
		})
	}
	knownRule := func(section string, index int64) types.Object {
		return newRule(types.StringValue(section), types.Int64Value(index))
	}
	for _, tc := range []struct {
		name    string
		rules   map[string]attr.Value
		summary string
		rule    string
		field   string
	}{
		{name: "empty"},
		{name: "empty rule name", rules: map[string]attr.Value{"": knownRule("A", 1)}, summary: "Invalid rule name"},
		{name: "single", rules: map[string]attr.Value{"R1": knownRule("A", 1)}},
		{name: "reordered", rules: map[string]attr.Value{"R1": knownRule("A", 2), "R2": knownRule("A", 1)}},
		{name: "independent sections", rules: map[string]attr.Value{"R1": knownRule("A", 1), "R2": knownRule("B", 1)}},
		{name: "empty section", rules: map[string]attr.Value{"R1": knownRule("", 1)},
			summary: "Invalid section name", rule: "R1", field: "section_name"},
		{name: "negative", rules: map[string]attr.Value{"R1": knownRule("A", -1)},
			summary: "Invalid rule index", rule: "R1", field: "index_in_section"},
		{name: "zero", rules: map[string]attr.Value{"R1": knownRule("A", 0)},
			summary: "Invalid rule index", rule: "R1", field: "index_in_section"},
		{name: "above maximum", rules: map[string]attr.Value{"R1": knownRule("A", 2)},
			summary: "Invalid rule index", rule: "R1", field: "index_in_section"},
		{name: "gap", rules: map[string]attr.Value{"R1": knownRule("A", 1), "R2": knownRule("A", 3)},
			summary: "Invalid rule index", rule: "R2", field: "index_in_section"},
		{name: "per section maximum", rules: map[string]attr.Value{"R1": knownRule("A", 1), "R2": knownRule("B", 2)},
			summary: "Invalid rule index", rule: "R2", field: "index_in_section"},
		{name: "duplicate", rules: map[string]attr.Value{"R1": knownRule("A", 1), "R2": knownRule("A", 1)},
			summary: "Duplicate rule index", field: "index_in_section"},
		{name: "unknown section", rules: map[string]attr.Value{
			"R1": knownRule("A", 2), "R2": newRule(types.StringUnknown(), types.Int64Value(1)),
		}},
		{name: "null section", rules: map[string]attr.Value{"R1": newRule(types.StringNull(), types.Int64Value(1))}},
		{name: "unknown index", rules: map[string]attr.Value{
			"R1": knownRule("A", 2), "R2": newRule(types.StringValue("A"), types.Int64Unknown()),
		}},
		{name: "null index", rules: map[string]attr.Value{"R1": newRule(types.StringValue("A"), types.Int64Null())}},
		{name: "unknown object", rules: map[string]attr.Value{"R1": knownRule("A", 2), "R2": types.ObjectUnknown(RuleAttrTypes)}},
		{name: "null object", rules: map[string]attr.Value{"R1": types.ObjectNull(RuleAttrTypes)}},
		{name: "unknown index does not hide known duplicate", rules: map[string]attr.Value{
			"R1": knownRule("A", 1), "R2": knownRule("A", 1), "R3": newRule(types.StringValue("A"), types.Int64Unknown()),
		}, summary: "Duplicate rule index", field: "index_in_section"},
		{name: "unknown section does not hide negative index", rules: map[string]attr.Value{
			"R1": knownRule("A", -1), "R2": newRule(types.StringUnknown(), types.Int64Value(1)),
		}, summary: "Invalid rule index", rule: "R1", field: "index_in_section"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := validator.MapRequest{
				Path: path.Root("rule_data"), ConfigValue: types.MapValueMust(RuleObjectType, tc.rules),
			}
			var resp validator.MapResponse
			ruleDataValidator{}.ValidateMap(ctx, req, &resp)
			if tc.summary == "" {
				require.Empty(t, resp.Diagnostics)
				return
			}
			require.Len(t, resp.Diagnostics, 1)
			require.Equal(t, diag.SeverityError, resp.Diagnostics[0].Severity())
			require.Equal(t, tc.summary, resp.Diagnostics[0].Summary())
			attributeDiag, ok := resp.Diagnostics[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			switch {
			case tc.summary == "Invalid rule name":
				require.Equal(t, req.Path.AtMapKey(""), attributeDiag.Path())
			case tc.rule != "":
				require.Equal(t, req.Path.AtMapKey(tc.rule).AtName(tc.field), attributeDiag.Path())
			default:
				require.Contains(t, []path.Path{
					req.Path.AtMapKey("R1").AtName(tc.field), req.Path.AtMapKey("R2").AtName(tc.field),
				}, attributeDiag.Path())
			}
		})
	}
}

func TestRuleDataValidatorUnknownAndNullMap(t *testing.T) {
	t.Parallel()
	for _, value := range []types.Map{types.MapNull(RuleObjectType), types.MapUnknown(RuleObjectType)} {
		existing := diag.NewWarningDiagnostic("Existing", "Keep this diagnostic")
		resp := validator.MapResponse{Diagnostics: diag.Diagnostics{existing}}
		ruleDataValidator{}.ValidateMap(ctx, validator.MapRequest{ConfigValue: value}, &resp)
		require.Equal(t, diag.Diagnostics{existing}, resp.Diagnostics)
	}
}

func TestRuleDataValidatorInvalidTypes(t *testing.T) {
	t.Parallel()
	for _, value := range []types.Map{
		types.MapValueMust(types.StringType, map[string]attr.Value{"R1": types.StringValue("invalid")}),
		types.MapValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{}}, map[string]attr.Value{
			"R1": types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		}),
	} {
		var resp validator.MapResponse
		ruleDataValidator{}.ValidateMap(ctx, validator.MapRequest{Path: path.Root("rule_data"), ConfigValue: value}, &resp)
		require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	}
}

func TestRuleDataValidatorSchema(t *testing.T) {
	t.Parallel()
	var resp resource.SchemaResponse
	(&alwaysOnRuleBulkResource{}).Schema(ctx, resource.SchemaRequest{}, &resp)
	rules, ok := resp.Schema.Attributes["rule_data"].(schema.MapNestedAttribute)
	require.True(t, ok)
	require.Contains(t, rules.Validators, ruleDataValidator{})
	v := ruleDataValidator{}
	require.NotEmpty(t, v.Description(ctx))
	require.Equal(t, v.Description(ctx), v.MarkdownDescription(ctx))
}

func TestValidateSections(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		sectionIDs map[string]string
		current    map[string]int64
		target     map[string]int64
		want       diag.Diagnostics
	}{
		{name: "nil maps"},
		{
			name: "empty maps", sectionIDs: map[string]string{},
			current: map[string]int64{}, target: map[string]int64{},
		},
		{
			name: "single section", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]int64{"A": 1}, target: map[string]int64{"A": 1},
		},
		{
			name: "reordered sections use independent indexes", sectionIDs: map[string]string{"A": "s1", "B": "s2"},
			current: map[string]int64{"A": 1, "B": 2}, target: map[string]int64{"A": 2, "B": 1},
		},
		{
			name: "current count mismatch", sectionIDs: map[string]string{"A": "s1"},
			target: map[string]int64{"A": 1},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Number of section does not match",
				"Number of sections current: 0, target: 1")},
		},
		{
			name: "target count mismatch", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]int64{"A": 1},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected always-on policy sections",
				"Expected: 0 sections, got: 1")},
		},
		{
			name: "current missing section", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]int64{"B": 1}, target: map[string]int64{"A": 1},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Section ID not found", "ID of section 'B' not found in TF state")},
		},
		{
			name: "target missing section", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]int64{"A": 1}, target: map[string]int64{"B": 1},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Section ID not found", "ID of section 'B' not found in TF state")},
		},
		{
			name: "section ID value is not validated", sectionIDs: map[string]string{"A": ""},
			current: map[string]int64{"A": 1}, target: map[string]int64{"A": 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			validateSections(tc.sectionIDs, tc.current, tc.target, &diags)
			require.Equal(t, tc.want, diags)
		})
	}
}

func TestValidateSectionsInvalidIndexes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		indexes map[string]int64
		summary string
		detail  string
	}{
		{"negative", map[string]int64{"A": -1, "B": 2}, "Invalid section index", "section index must be between 1 and 2, got -1"},
		{"zero", map[string]int64{"A": 0, "B": 2}, "Invalid section index", "section index must be between 1 and 2, got 0"},
		{"above maximum", map[string]int64{"A": 3, "B": 2}, "Invalid section index", "section index must be between 1 and 2, got 3"},
		{"duplicate", map[string]int64{"A": 1, "B": 1}, "Duplicate section index", "section index '1' was used multiple times"},
	} {
		for _, side := range []string{"current", "target"} {
			t.Run(side+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				current, target := map[string]int64{"A": 1, "B": 2}, map[string]int64{"A": 1, "B": 2}
				if side == "current" {
					current = tc.indexes
				} else {
					target = tc.indexes
				}
				var diags diag.Diagnostics
				validateSections(map[string]string{"A": "s1", "B": "s2"}, current, target, &diags)
				require.Len(t, diags, 1)
				require.Equal(t, diag.SeverityError, diags[0].Severity())
				require.Equal(t, tc.summary, diags[0].Summary())
				if tc.name == "duplicate" {
					// Map iteration determines the order of the reported section names.
					require.Contains(t, []string{tc.detail + " ('A', 'B')", tc.detail + " ('B', 'A')"}, diags[0].Detail())
				} else {
					require.Equal(t, tc.detail, diags[0].Detail())
				}
			})
		}
	}
}

func TestValidateRules(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		sectionIDs map[string]string
		current    map[string]map[string]int64
		target     map[string]map[string]int64
		want       diag.Diagnostics
	}{
		{name: "nil maps"},
		{
			name: "empty maps", sectionIDs: map[string]string{},
			current: map[string]map[string]int64{}, target: map[string]map[string]int64{},
		},
		{
			name: "empty section", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"A": nil}, target: map[string]map[string]int64{"A": {}},
		},
		{
			name: "reordered rules with per section indexes", sectionIDs: map[string]string{"A": "s1", "B": "s2"},
			current: map[string]map[string]int64{"A": {"R1": 1, "R2": 2}, "B": {"R3": 1}},
			target:  map[string]map[string]int64{"A": {"R1": 2, "R2": 1}, "B": {"R3": 1}},
		},
		{
			name: "fewer current sections", target: map[string]map[string]int64{"A": {"R1": 1}},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Number of sections does not match",
				"Current sections: 0 target: 1")},
		},
		{
			name: "fewer target sections", current: map[string]map[string]int64{"A": {"R1": 1}},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Number of sections does not match",
				"Current sections: 1 target: 0")},
		},
		{
			name: "current section missing from target", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"Missing": {"R1": 1}},
			target:  map[string]map[string]int64{"A": {"R1": 1}},
			want:    diag.Diagnostics{diag.NewErrorDiagnostic("Section not found", "section 'Missing' not found in the target")},
		},
		{
			name: "target section differs from current", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"A": {"R1": 1}},
			target:  map[string]map[string]int64{"Missing": {"R1": 1}},
			want:    diag.Diagnostics{diag.NewErrorDiagnostic("Section not found", "section 'A' not found in the target")},
		},
		{
			name: "section ID missing", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"Missing": {"R1": 1}},
			target:  map[string]map[string]int64{"Missing": {"R1": 1}},
			want:    diag.Diagnostics{diag.NewErrorDiagnostic("Section ID not found", "ID of section 'Missing' not found in the state")},
		},
		{
			name: "fewer current rules", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"A": {}},
			target:  map[string]map[string]int64{"A": {"R1": 1}},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Number of rules does not match",
				"Rules section 'A' current: 0, target: 1")},
		},
		{
			name: "unexpected current rules", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"A": {"R1": 1}},
			target:  map[string]map[string]int64{"A": {}},
			want: diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected always-on rules",
				"Unexpected always-on rules found in section 'A' expected: 0 rules, got: 1")},
		},
		{
			name: "rule missing from target", sectionIDs: map[string]string{"A": "s1"},
			current: map[string]map[string]int64{"A": {"R1": 1}},
			target:  map[string]map[string]int64{"A": {"R2": 1}},
			want:    diag.Diagnostics{diag.NewErrorDiagnostic("Rule not found", "Rule 'R1' in section 'A' not found in target")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			validateRules(tc.sectionIDs, tc.current, tc.target, &diags)
			require.Equal(t, tc.want, diags)
		})
	}
}

func TestValidateRulesInvalidIndexes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		indexes map[string]int64
		summary string
		detail  string
	}{
		{"negative", map[string]int64{"R1": -1, "R2": 2}, "Invalid rule index", "Rule index must be between 1 and 2, got -1 for rule R1"},
		{"zero", map[string]int64{"R1": 0, "R2": 2}, "Invalid rule index", "Rule index must be between 1 and 2, got 0 for rule R1"},
		{"above maximum", map[string]int64{"R1": 3, "R2": 2}, "Invalid rule index", "Rule index must be between 1 and 2, got 3 for rule R1"},
		{"duplicate", map[string]int64{"R1": 1, "R2": 1}, "Duplicate rule index", "Duplicate rule index 1, for rule "},
	} {
		for _, side := range []string{"current", "target"} {
			t.Run(side+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				current := map[string]map[string]int64{"A": {"R1": 1, "R2": 2}}
				target := map[string]map[string]int64{"A": {"R1": 1, "R2": 2}}
				if side == "current" {
					current["A"] = tc.indexes
				} else {
					target["A"] = tc.indexes
				}
				var diags diag.Diagnostics
				validateRules(map[string]string{"A": "s1"}, current, target, &diags)
				require.Len(t, diags, 1)
				require.Equal(t, diag.SeverityError, diags[0].Severity())
				require.Equal(t, tc.summary, diags[0].Summary())
				if tc.name == "duplicate" {
					// Map iteration determines which rule is reported as the duplicate.
					require.Contains(t, []string{tc.detail + "R1", tc.detail + "R2"}, diags[0].Detail())
				} else {
					require.Equal(t, tc.detail, diags[0].Detail())
				}
			})
		}
	}
}

func TestValidatorsPreserveDiagnostics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		validate func(*diag.Diagnostics)
		summary  string
	}{
		{
			name: "sections",
			validate: func(diags *diag.Diagnostics) {
				validateSections(map[string]string{"A": "s1"}, nil, map[string]int64{"A": 1}, diags)
			},
			summary: "Number of section does not match",
		},
		{
			name: "rules",
			validate: func(diags *diag.Diagnostics) {
				validateRules(nil, map[string]map[string]int64{"A": {}}, nil, diags)
			},
			summary: "Number of sections does not match",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			existing := diag.NewWarningDiagnostic("Existing warning", "Keep this diagnostic")
			diags := diag.Diagnostics{existing}
			tc.validate(&diags)
			require.Len(t, diags, 2)
			require.Equal(t, existing, diags[0])
			require.Equal(t, tc.summary, diags[1].Summary())
			require.Equal(t, diag.SeverityError, diags[1].Severity())
		})
	}
}
