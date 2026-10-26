package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func activePeriod(from, until types.String, fromFlag, untilFlag types.Bool) types.Object {
	return objectWithValues(map[string]attr.Value{"effective_from": from, "expires_at": until, "use_effective_from": fromFlag, "use_expires_at": untilFlag})
}
func TestActivePeriodModifier(t *testing.T) {
	t.Parallel()
	null := types.StringNull()
	a, b := types.StringValue("2026-10-21T00:00:00Z"), types.StringValue("2026-11-21T00:00:00Z")
	u := types.BoolUnknown()
	f := types.BoolValue(false)
	yes := types.BoolValue(true)
	defaults := activePeriod(null, null, f, f)
	unknown := types.ObjectUnknown(defaults.AttributeTypes(context.Background()))
	nullObj := types.ObjectNull(defaults.AttributeTypes(context.Background()))
	configured := activePeriod(a, b, u, u)
	state := activePeriod(a, b, yes, yes)
	for _, tc := range []struct {
		name                      string
		config, state, plan, want types.Object
	}{
		{"omitted uses disabled flags", nullObj, state, unknown, defaults},
		{"first apply", configured, nullObj, configured, configured}, {"unknown state", configured, unknown, configured, configured},
		{"unchanged", configured, state, state, state},
		{"set effective date", activePeriod(a, null, u, u), defaults, unknown, activePeriod(a, null, yes, f)},
		{"set expiration date", activePeriod(null, b, u, u), defaults, unknown, activePeriod(null, b, f, yes)},
		{"remove effective date", activePeriod(null, b, u, u), state, unknown, activePeriod(null, b, f, yes)},
		{"remove expiration date", activePeriod(a, null, u, u), state, unknown, activePeriod(a, null, yes, f)},
		{"unknown config", unknown, state, unknown, unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.ObjectResponse{PlanValue: tc.plan}
			ActivePeriodModifier().PlanModifyObject(context.Background(), planmodifier.ObjectRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.True(t, tc.want.Equal(resp.PlanValue), "want %s, got %s", tc.want, resp.PlanValue)
		})
	}
}
