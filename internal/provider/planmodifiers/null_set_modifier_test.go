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

func TestNullSetModifier(t *testing.T) {
	t.Parallel()
	null := types.SetNull(types.StringType)
	unknown := types.SetUnknown(types.StringType)
	value := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("configured")})
	empty := types.SetValueMust(types.StringType, []attr.Value{})
	for _, tc := range []struct {
		name               string
		config, plan, want types.Set
		destroy            bool
	}{
		{"omitted plans empty", null, unknown, empty, false}, {"configured stays", value, value, value, false},
		{"unknown config", unknown, unknown, unknown, false}, {"destroy stays null", null, null, null, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := tftypes.NewValue(tftypes.String, "resource")
			if tc.destroy {
				raw = tftypes.NewValue(tftypes.String, nil)
			}
			resp := &planmodifier.SetResponse{PlanValue: tc.plan}
			NullSetModifier(types.StringType).PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: tc.config, PlanValue: tc.plan, Plan: tfsdk.Plan{Raw: raw}}, resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, tc.want.Equal(resp.PlanValue))
		})
	}
}
