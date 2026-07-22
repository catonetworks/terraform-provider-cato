package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestIfwRulesIndexEnabledIsComputedOnly(t *testing.T) {
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
	require.False(t, enabled.Optional)
	require.False(t, enabled.Required)
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

func TestBuildTLSRuleIndexStateDataUsesAPIComputedFields(t *testing.T) {
	t.Parallel()

	rule := TLSRulesRuleDataIndex{
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

	var got TLSRulesRuleItemIndex
	asDiags := obj.As(context.Background(), &got, basetypes.ObjectAsOptions{})
	require.False(t, asDiags.HasError())

	require.Equal(t, "rule-id-c", got.ID.ValueString())
	require.Equal(t, "api description", got.Description.ValueString())
	require.True(t, got.Enabled.ValueBool())
}
