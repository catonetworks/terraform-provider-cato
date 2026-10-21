package wanindex

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestBuildWanRuleIndexStateDataUsesAPIComputedFields(t *testing.T) {
	t.Parallel()

	rule := WanRulesRuleDataIndex{
		IndexInSection: 7,
		SectionName:    "Section B",
		RuleName:       "Rule B",
		Description:    "plan description",
		Enabled:        false,
	}

	obj, diags := buildWanRuleIndexStateData(
		rule,
		map[string]string{"Rule B": "rule-id-b"},
		map[string]string{"Rule B": "api description"},
		map[string]bool{"Rule B": true},
	)
	require.False(t, diags.HasError())

	var got WanRulesRuleItemIndex
	asDiags := obj.As(context.Background(), &got, basetypes.ObjectAsOptions{})
	require.False(t, asDiags.HasError())

	require.Equal(t, "rule-id-b", got.ID.ValueString())
	require.Equal(t, "api description", got.Description.ValueString())
	require.True(t, got.Enabled.ValueBool())
}
