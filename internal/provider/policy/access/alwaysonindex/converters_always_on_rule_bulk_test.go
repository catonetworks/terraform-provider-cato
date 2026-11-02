package alwaysonindex

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestParsePolicy(t *testing.T) {
	t.Parallel()
	var result cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{
		"revision":{"changes":7},
		"sections":[
			{"section":{"id":"s2","name":"Second"}},
			{"section":{"id":"s1","name":"First"}},
			{"section":{"id":"s3","name":"Empty"}}
		],
		"rules":[
			{"rule":{"id":"r1","name":"Alpha","index":10,"section":{"id":"s1","name":"First"}}},
			{"rule":{"id":"r2","name":"Beta","index":20,"section":{"id":"s2","name":"Second"}}},
			{"rule":{"id":"r3","name":"Gamma","index":30,"section":{"id":"s1","name":"First"}}}
		]
	}}}}`), &result))

	var diags diag.Diagnostics
	actual := parseRulesIndex(ctx, &result, &diags)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, actual)
	require.Equal(t, types.Int64Value(7), actual.Changes)
	require.Equal(t, types.MapValueMust(types.ObjectType{AttrTypes: SectionAttrTypes}, map[string]attr.Value{
		"Second": types.ObjectValueMust(SectionAttrTypes, map[string]attr.Value{
			"id": types.StringValue("s2"), "section_index": types.Int64Value(1),
		}),
		"First": types.ObjectValueMust(SectionAttrTypes, map[string]attr.Value{
			"id": types.StringValue("s1"), "section_index": types.Int64Value(2),
		}),
		"Empty": types.ObjectValueMust(SectionAttrTypes, map[string]attr.Value{
			"id": types.StringValue("s3"), "section_index": types.Int64Value(3),
		}),
	}), actual.SectionData)
	require.Equal(t, types.MapValueMust(types.ObjectType{AttrTypes: RuleAttrTypes}, map[string]attr.Value{
		"Alpha": types.ObjectValueMust(RuleAttrTypes, map[string]attr.Value{
			"id": types.StringValue("r1"), "section_name": types.StringValue("First"), "index_in_section": types.Int64Value(1),
		}),
		"Beta": types.ObjectValueMust(RuleAttrTypes, map[string]attr.Value{
			"id": types.StringValue("r2"), "section_name": types.StringValue("Second"), "index_in_section": types.Int64Value(1),
		}),
		"Gamma": types.ObjectValueMust(RuleAttrTypes, map[string]attr.Value{
			"id": types.StringValue("r3"), "section_name": types.StringValue("First"), "index_in_section": types.Int64Value(2),
		}),
	}), actual.RuleData)
}

func TestParsePolicyEmpty(t *testing.T) {
	t.Parallel()
	for _, changes := range []int64{0, math.MaxInt64} {
		var result cato.PolicyZtnaAlwaysOnPolicy
		body := fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[],"rules":[],"revision":{"changes":%d}}}}}`, changes)
		require.NoError(t, json.Unmarshal([]byte(body), &result))
		var diags diag.Diagnostics
		actual := parseRulesIndex(ctx, &result, &diags)
		require.False(t, diags.HasError(), "%v", diags)
		require.NotNil(t, actual)
		require.Equal(t, changes, actual.Changes.ValueInt64())
		require.Equal(t, types.MapValueMust(types.ObjectType{AttrTypes: SectionAttrTypes}, map[string]attr.Value{}), actual.SectionData)
		require.Equal(t, types.MapValueMust(types.ObjectType{AttrTypes: RuleAttrTypes}, map[string]attr.Value{}), actual.RuleData)
	}
}

func TestParsePolicyInvalidResponse(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, policy string }{
		{name: "missing lists", policy: `{}`},
		{name: "missing sections", policy: `{"rules":[]}`},
		{name: "missing rules", policy: `{"sections":[]}`},
		{name: "nil section", policy: `{"sections":[null],"rules":[]}`},
		{name: "missing section ID", policy: `{"sections":[{"section":{"name":"A"}}],"rules":[]}`},
		{name: "missing section name", policy: `{"sections":[{"section":{"id":"s1"}}],"rules":[]}`},
		{name: "duplicate section name", policy: `{"sections":[{"section":{"id":"s1","name":"A"}},
			{"section":{"id":"s2","name":"A"}}],"rules":[]}`},
		{name: "duplicate section ID", policy: `{"sections":[{"section":{"id":"s1","name":"A"}},
			{"section":{"id":"s1","name":"B"}}],"rules":[]}`},
		{name: "nil rule", policy: `{"sections":[],"rules":[null]}`},
		{name: "missing rule ID", policy: `{"sections":[{"section":{"id":"s1","name":"S"}}],
			"rules":[{"rule":{"name":"A","section":{"id":"s1","name":"S"}}}]}`},
		{name: "missing rule name", policy: `{"sections":[{"section":{"id":"s1","name":"S"}}],
			"rules":[{"rule":{"id":"r1","section":{"id":"s1","name":"S"}}}]}`},
		{name: "missing parent", policy: `{"sections":[],"rules":[{"rule":{"id":"r1","name":"A","section":{"id":"s1"}}}]}`},
		{name: "duplicate rule name", policy: `{"sections":[{"section":{"id":"s1","name":"S"}}],"rules":[
			{"rule":{"id":"r1","name":"A","section":{"id":"s1","name":"S"}}},
			{"rule":{"id":"r2","name":"A","section":{"id":"s1","name":"S"}}}]}`},
		{name: "duplicate rule ID", policy: `{"sections":[{"section":{"id":"s1","name":"S"}}],"rules":[
			{"rule":{"id":"r1","name":"A","section":{"id":"s1","name":"S"}}},
			{"rule":{"id":"r1","name":"B","section":{"id":"s1","name":"S"}}}]}`},
		{name: "negative changes", policy: `{"sections":[],"rules":[],"revision":{"changes":-1}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var result cato.PolicyZtnaAlwaysOnPolicy
			require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":`+tc.policy+`}}}`), &result))
			var diags diag.Diagnostics
			require.Nil(t, parseRulesIndex(ctx, &result, &diags))
			require.True(t, diags.HasError())
			require.Equal(t, "Invalid Always On policy response", diags.Errors()[0].Summary())
		})
	}
	for _, result := range []*cato.PolicyZtnaAlwaysOnPolicy{nil, {}} {
		var diags diag.Diagnostics
		require.Nil(t, parseRulesIndex(ctx, result, &diags))
		require.True(t, diags.HasError())
	}
}

func sectionOrderModel(t *testing.T, sections map[string]Section) RulesIndex {
	t.Helper()
	value, diags := types.MapValueFrom(context.Background(), SectionObjectType, sections)
	require.False(t, diags.HasError(), "%v", diags)
	return RulesIndex{SectionData: value}
}
