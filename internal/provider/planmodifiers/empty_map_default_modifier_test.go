package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestEmptyMapDefault(t *testing.T) {
	t.Parallel()

	null := types.MapNull(types.StringType)
	unknown := types.MapUnknown(types.StringType)
	value := types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringValue("configured")})
	empty := types.MapValueMust(types.StringType, map[string]attr.Value{})
	objectType := types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType}}
	for _, tc := range []struct {
		name                      string
		config, state, plan, want types.Map
	}{
		{"new omitted map", null, null, unknown, empty},
		{"existing omitted map", null, value, value, empty},
		{"existing empty map", null, empty, empty, empty},
		{"configured map", value, null, value, value},
		{"configured empty map", empty, null, empty, empty},
		{"unknown config", unknown, null, unknown, unknown},
		{"unknown config with existing map", unknown, value, unknown, unknown},
		{"unknown state", null, unknown, unknown, empty},
		{"object elements", types.MapNull(objectType), types.MapNull(objectType), types.MapUnknown(objectType), types.MapValueMust(objectType, map[string]attr.Value{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.MapResponse{PlanValue: tc.plan}
			EmptyMapDefault(tc.config.ElementType(context.Background())).PlanModifyMap(context.Background(), planmodifier.MapRequest{
				Plan:        tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})},
				ConfigValue: tc.config,
				StateValue:  tc.state,
				PlanValue:   tc.plan,
			}, resp)
			require.Empty(t, resp.Diagnostics)
			require.True(t, tc.want.Equal(resp.PlanValue), "want %s, got %s", tc.want, resp.PlanValue)
		})
	}

	t.Run("destroy preserves null", func(t *testing.T) {
		t.Parallel()
		resp := &planmodifier.MapResponse{PlanValue: null}
		EmptyMapDefault(types.StringType).PlanModifyMap(context.Background(), planmodifier.MapRequest{
			Plan:        tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, nil)},
			ConfigValue: null, StateValue: value, PlanValue: null,
		}, resp)
		require.Empty(t, resp.Diagnostics)
		require.True(t, resp.PlanValue.IsNull())
	})
}
