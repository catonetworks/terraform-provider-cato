package tls

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestBuildTLSRuleIndexStateDataUsesAPIComputedFields(t *testing.T) {
	t.Parallel()

	rule := RulesRuleDataIndex{
		IndexInSection: 11,
		SectionName:    "Section C",
		RuleName:       "Rule C",
		Description:    "plan description",
		Enabled:        false,
	}

	obj, diags := buildTLSRuleIndexStateData(
		rule,
		map[string]string{"Rule C": "rule-id-c"},
		map[string]string{"Rule C": "api description"},
		map[string]bool{"Rule C": true},
	)
	require.False(t, diags.HasError())

	var got RulesRuleItemIndex
	asDiags := obj.As(context.Background(), &got, basetypes.ObjectAsOptions{})
	require.False(t, asDiags.HasError())

	require.Equal(t, "rule-id-c", got.ID.ValueString())
	require.Equal(t, "api description", got.Description.ValueString())
	require.True(t, got.Enabled.ValueBool())
}
