package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestEmptySetDefault(t *testing.T) {
	t.Parallel()
	null := types.SetNull(types.StringType)
	unknown := types.SetUnknown(types.StringType)
	value := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("configured")})
	empty := types.SetValueMust(types.StringType, []attr.Value{})
	for _, tc := range []struct {
		name                      string
		config, state, plan, want types.Set
	}{
		{"new omitted set", null, null, unknown, empty}, {"existing omitted set", null, value, value, value},
		{"configured", value, null, value, value}, {"unknown config", unknown, null, unknown, unknown}, {"unknown state", null, unknown, unknown, unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &planmodifier.SetResponse{PlanValue: tc.plan}
			EmptySetDefault(types.StringType).PlanModifySet(context.Background(), planmodifier.SetRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}, resp)
			require.False(t, resp.Diagnostics.HasError())
			require.True(t, tc.want.Equal(resp.PlanValue))
		})
	}
}
