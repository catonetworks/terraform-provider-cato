package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func enumStrings[T ~string](values []T) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = string(v)
	}
	return result
}
func checkStringValidator(t *testing.T, v validator.String, valid []string, skipNull bool) {
	t.Helper()
	ctx := context.Background()
	require.NotEmpty(t, v.Description(ctx))
	require.Equal(t, v.Description(ctx), v.MarkdownDescription(ctx))
	for _, value := range valid {
		t.Run(value, func(t *testing.T) {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{Path: path.Root("setting"), ConfigValue: types.StringValue(value)}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"invalid", types.StringValue("INVALID"), true}, {"empty", types.StringValue(""), true},
		{"null", types.StringNull(), !skipNull}, {"unknown", types.StringUnknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{Path: path.Root("setting"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			if tc.wantError {
				require.Contains(t, resp.Diagnostics[0].Detail(), "setting")
			}
		})
	}
}
func checkSetValidator(t *testing.T, v validator.Set, valid []string) {
	t.Helper()
	ctx := context.Background()
	items := make([]attr.Value, len(valid))
	for i, s := range valid {
		items[i] = types.StringValue(s)
	}
	require.NotEmpty(t, v.Description(ctx))
	require.Equal(t, v.Description(ctx), v.MarkdownDescription(ctx))
	for _, tc := range []struct {
		name      string
		value     types.Set
		wantError bool
	}{
		{"all valid values", types.SetValueMust(types.StringType, items), false},
		{"mixed invalid value", types.SetValueMust(types.StringType, []attr.Value{items[0], types.StringValue("INVALID")}), true},
		{"empty", types.SetValueMust(types.StringType, []attr.Value{}), false},
		{"null", types.SetNull(types.StringType), false}, {"unknown", types.SetUnknown(types.StringType), false},
		{"wrong element type", types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(1)}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.SetResponse{}
			v.ValidateSet(ctx, validator.SetRequest{Path: path.Root("setting"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}
