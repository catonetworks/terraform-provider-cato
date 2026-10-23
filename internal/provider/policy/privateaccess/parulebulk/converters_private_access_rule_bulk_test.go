package parulebulk

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPrivateAccessBulkConversionDiagnostics(t *testing.T) {
	ctx := context.Background()
	t.Run("object decoding", func(t *testing.T) {
		var diags diag.Diagnostics
		privateAccessRuleFromObject(ctx, &diags, types.ObjectUnknown(PrivateAccessBulkRuleTypes))
		require.True(t, diags.HasError())
	})
	t.Run("object encoding", func(t *testing.T) {
		original := PrivateAccessBulkRuleTypes
		PrivateAccessBulkRuleTypes = map[string]attr.Type{"name": types.BoolType}
		t.Cleanup(func() { PrivateAccessBulkRuleTypes = original })
		var diags diag.Diagnostics
		privateAccessRuleObjectFromModel(ctx, &diags, &PrivateAccessBulkRule{}, PrivateAccessBulkRule{})
		require.True(t, diags.HasError())
	})
	t.Run("map encoding", func(t *testing.T) {
		var diags diag.Diagnostics
		privateAccessRuleMapFromObjects(ctx, &diags, map[string]types.Object{
			"rule": types.ObjectNull(map[string]attr.Type{"unexpected": types.BoolType}),
		})
		require.True(t, diags.HasError())
	})
}

func TestPrivateAccessBulkHydrationStopsOnConversionError(t *testing.T) {
	t.Parallel()
	// The outer map decodes successfully, but its unknown rule cannot be decoded
	// into a Go struct. No configured SDK is needed because hydration must stop here.
	plan := &PrivateAccessRuleBulkModel{
		RuleData: types.MapValueMust(types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes}, map[string]attr.Value{
			"rule": types.ObjectUnknown(PrivateAccessBulkRuleTypes),
		}),
	}
	r := &privAccessRuleBulkResource{}
	state, rules, diags, err := r.hydratePrivAccessRuleBulkState(context.Background(), plan)
	require.ErrorIs(t, err, ErrConvertError)
	require.True(t, diags.HasError())
	require.Nil(t, state)
	require.Nil(t, rules)
}
