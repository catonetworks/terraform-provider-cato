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

func TestNullStringModifier(t *testing.T) {
	t.Parallel()
	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"value": tftypes.String}}
	plan := tftypes.NewValue(objectType, map[string]tftypes.Value{"value": tftypes.NewValue(tftypes.String, "previous")})
	for _, test := range []struct {
		name                  string
		config, planned, want types.String
		destroy               bool
	}{
		{name: "null becomes empty", config: types.StringNull(), planned: types.StringUnknown(), want: types.StringValue("")},
		{name: "configured value", config: types.StringValue("configured"), planned: types.StringUnknown(), want: types.StringValue("configured")},
		{name: "configured empty", config: types.StringValue(""), planned: types.StringUnknown(), want: types.StringValue("")},
		{name: "unknown stays unknown", config: types.StringUnknown(), planned: types.StringUnknown(), want: types.StringUnknown()},
		{name: "destroy stays null", config: types.StringNull(), planned: types.StringNull(), want: types.StringNull(), destroy: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			raw := plan
			if test.destroy {
				raw = tftypes.NewValue(objectType, nil)
			}
			req := planmodifier.StringRequest{ConfigValue: test.config, PlanValue: test.planned, Plan: tfsdk.Plan{Raw: raw}}
			resp := planmodifier.StringResponse{PlanValue: test.planned}
			modifier := NullStringModifier()
			modifier.PlanModifyString(context.Background(), req, &resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, test.want.Equal(resp.PlanValue))
		})
	}
}
