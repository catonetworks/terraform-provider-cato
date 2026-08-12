package provider

import (
	"testing"

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
