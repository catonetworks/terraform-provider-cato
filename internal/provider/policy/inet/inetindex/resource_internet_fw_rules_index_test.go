package inetindex

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestBuildIfwSectionMoves_skipsProtectedSectionAndUsesItAsAnchor(t *testing.T) {
	t.Parallel()

	sections := []IfwRulesSectionDataIndex{
		{SectionIndex: 1, SectionName: "Default Outbound Internet"},
		{SectionIndex: 2, SectionName: "Custom"},
	}
	sectionIDs := map[string]string{
		"Default Outbound Internet": "system-section",
		"Custom":                    "custom-section",
	}
	protectedSectionIDs := map[string]struct{}{"system-section": {}}

	moves := buildIfwSectionMoves(sections, sectionIDs, protectedSectionIDs, "")

	require.Len(t, moves, 2)
	require.True(t, moves[0].protected)
	require.Equal(t, "system-section", moves[0].input.ID)
	require.Equal(t, "LAST_IN_POLICY", string(moves[0].input.To.Position))
	require.Nil(t, moves[0].input.To.Ref)

	require.False(t, moves[1].protected)
	require.Equal(t, "custom-section", moves[1].input.ID)
	require.Equal(t, "AFTER_SECTION", string(moves[1].input.To.Position))
	require.Equal(t, "system-section", *moves[1].input.To.Ref)
}

func TestBuildIfwSectionMoves_preservesDistinctSectionReferences(t *testing.T) {
	t.Parallel()

	sections := []IfwRulesSectionDataIndex{
		{SectionIndex: 1, SectionName: "A"},
		{SectionIndex: 2, SectionName: "B"},
		{SectionIndex: 3, SectionName: "C"},
	}
	sectionIDs := map[string]string{
		"A": "section-a",
		"B": "section-b",
		"C": "section-c",
	}

	moves := buildIfwSectionMoves(sections, sectionIDs, nil, "anchor")

	require.Len(t, moves, 3)
	require.Equal(t, "anchor", *moves[0].input.To.Ref)
	require.Equal(t, "section-a", *moves[1].input.To.Ref)
	require.Equal(t, "section-b", *moves[2].input.To.Ref)
	require.False(t, moves[0].protected)
	require.False(t, moves[1].protected)
	require.False(t, moves[2].protected)
}

func TestBuildIfwRuleIndexStateDataUsesAPIComputedFields(t *testing.T) {
	t.Parallel()

	rule := IfwRulesRuleDataIndex{
		IndexInSection: 3,
		SectionName:    "Section A",
		RuleName:       "Rule A",
		Description:    "plan description",
		Enabled:        types.BoolUnknown(),
	}

	obj, diags := buildIfwRuleIndexStateData(
		rule,
		map[string]string{"Rule A": "rule-id-a"},
		map[string]string{"Rule A": "api description"},
		map[string]bool{"Rule A": true},
	)
	require.False(t, diags.HasError())

	var got IfwRulesRuleItemIndex
	asDiags := obj.As(context.Background(), &got, basetypes.ObjectAsOptions{})
	require.False(t, asDiags.HasError())

	require.Equal(t, "rule-id-a", got.ID.ValueString())
	require.Equal(t, "api description", got.Description.ValueString())
	require.True(t, got.Enabled.ValueBool())
}

func TestIfwRulesIndexEnabledIsSettableAndNewValuesRemainUnknown(t *testing.T) {
	t.Parallel()

	var resp resource.SchemaResponse
	(&ifwRulesIndexResource{}).Schema(
		context.Background(),
		resource.SchemaRequest{},
		&resp,
	)
	require.False(t, resp.Diagnostics.HasError())

	ruleData, ok := resp.Schema.Attributes["rule_data"].(schema.MapNestedAttribute)
	require.True(t, ok)

	enabled, ok := ruleData.NestedObject.Attributes["enabled"].(schema.BoolAttribute)
	require.True(t, ok)
	require.True(t, enabled.Computed)
	require.True(t, enabled.Optional)
	require.False(t, enabled.Required)
	require.Len(t, enabled.PlanModifiers, 1)

	req := planmodifier.BoolRequest{
		ConfigValue: types.BoolNull(),
		PlanValue:   types.BoolUnknown(),
		StateValue:  types.BoolNull(),
	}
	modifierResp := planmodifier.BoolResponse{PlanValue: req.PlanValue}
	enabled.PlanModifiers[0].PlanModifyBool(context.Background(), req, &modifierResp)

	require.True(t, modifierResp.PlanValue.IsUnknown())
}
