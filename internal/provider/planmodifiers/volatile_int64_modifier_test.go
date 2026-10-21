package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestVolatileInt64(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		state      tftypes.Value
		plan, want types.Int64
	}{
		{"create keeps unknown", tftypes.NewValue(tftypes.String, nil), types.Int64Unknown(), types.Int64Unknown()},
		{"create preserves known plan", tftypes.NewValue(tftypes.String, nil), types.Int64Value(3), types.Int64Value(3)},
		{"update recomputes", tftypes.NewValue(tftypes.String, "resource"), types.Int64Value(3), types.Int64Unknown()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.Int64Response{PlanValue: tc.plan}
			VolatileInt64().PlanModifyInt64(context.Background(), planmodifier.Int64Request{PlanValue: tc.plan, State: tfsdk.State{Raw: tc.state}}, resp)
			require.Equal(t, tc.want, resp.PlanValue)
			require.False(t, resp.Diagnostics.HasError())
		})
	}
}
