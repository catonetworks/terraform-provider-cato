package lfsubpolicy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestLfSubPolicyPositionReplacementModifier(t *testing.T) {
	t.Parallel()

	modifier := configuredPositionModifier{}
	first := lfSubPolicyPosition("LAST_IN_POLICY", types.StringNull())
	second := lfSubPolicyPosition("AFTER_RULE", types.StringValue("rule-123"))

	t.Run("imported null position accepts initial configuration", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: types.ObjectNull(policyrule.PositionAttrTypes),
			PlanValue:  first,
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("unchanged configured position stays in place", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  first,
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("configured position change is handled by update", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  second,
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("unknown configured position is deferred", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  types.ObjectUnknown(policyrule.PositionAttrTypes),
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("configured position cannot be removed", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  types.ObjectNull(policyrule.PositionAttrTypes),
		}, resp)

		require.True(t, resp.Diagnostics.HasError())
	})
}
