package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestLfSubPolicyAtSchemaIsImportSafe(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&lfSubPolicyResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "unexpected schema diagnostics: %+v", resp.Diagnostics)

	at, ok := resp.Schema.Attributes["at"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	require.True(t, at.Optional)
	require.False(t, at.Required)
	require.Len(t, at.PlanModifiers, 1)
}

func TestLfSubPolicyPositionReplacementModifier(t *testing.T) {
	t.Parallel()

	modifier := requiresReplaceForConfiguredPosition{}
	first := lfSubPolicyPosition("LAST_IN_POLICY", types.StringNull())
	second := lfSubPolicyPosition("AFTER_RULE", types.StringValue("rule-123"))

	t.Run("imported null position accepts initial configuration", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: types.ObjectNull(PositionAttrTypes),
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

	t.Run("configured position change requires replacement", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  second,
		}, resp)

		require.True(t, resp.RequiresReplace)
	})

	t.Run("unknown configured position requires replacement", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  types.ObjectUnknown(PositionAttrTypes),
		}, resp)

		require.True(t, resp.RequiresReplace)
	})
}

func lfSubPolicyPosition(position string, ref types.String) types.Object {
	return types.ObjectValueMust(PositionAttrTypes, map[string]attr.Value{
		"position": types.StringValue(position),
		"ref":      ref,
	})
}
