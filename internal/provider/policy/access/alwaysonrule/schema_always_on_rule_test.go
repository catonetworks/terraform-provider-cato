package alwaysonrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestAlwaysOnRuleSetDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	resourceSchema := alwaysOnSchema(t)
	for _, attributePath := range []path.Path{
		path.Root("rule").AtName("platform"),
		path.Root("rule").AtName("device_posture_profile"),
		path.Root("rule").AtName("source").AtName("users"),
		path.Root("rule").AtName("source").AtName("user_groups"),
	} {
		t.Run(attributePath.String(), func(t *testing.T) {
			t.Parallel()
			attribute, diags := resourceSchema.AttributeAtPath(ctx, attributePath)
			require.False(t, diags.HasError(), "%v", diags)
			require.True(t, attribute.IsOptional())
			require.True(t, attribute.IsComputed())
			require.False(t, attribute.IsRequired())
			var modifiers []planmodifier.Set
			var elementType attr.Type
			var element attr.Value
			switch setAttribute := attribute.(type) {
			case schema.SetAttribute:
				modifiers = setAttribute.PlanModifiers
				elementType = types.StringType
				element = types.StringValue("WINDOWS")
			case schema.SetNestedAttribute:
				modifiers = setAttribute.PlanModifiers
				elementType = idname.RefObjectType
				object, objectDiags := types.ObjectValueFrom(ctx, idname.RefAttrTypes, idname.Ref{
					ID: types.StringValue("reference-1"), Name: types.StringValue("Example"),
				})
				require.False(t, objectDiags.HasError(), "%v", objectDiags)
				element = object
			default:
				t.Fatalf("expected a set attribute, got %T", attribute)
			}
			null := types.SetNull(elementType)
			unknown := types.SetUnknown(elementType)
			empty := types.SetValueMust(elementType, nil)
			value := types.SetValueMust(elementType, []attr.Value{element})
			for _, tc := range []struct {
				name                      string
				config, plan, state, want types.Set
				destroy                   bool
			}{
				{name: "omitted on create", config: null, plan: unknown, state: null, want: empty},
				{name: "removed on update", config: null, plan: unknown, state: value, want: empty},
				{name: "explicit empty", config: empty, plan: empty, state: value, want: empty},
				{name: "configured value", config: value, plan: value, state: value, want: value},
				{name: "unknown config", config: unknown, plan: unknown, state: value, want: unknown},
				{name: "destroy", config: null, plan: null, state: value, want: null, destroy: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					plan := alwaysOnPlan(t, alwaysOnPlanModel(t))
					if tc.destroy {
						plan.Raw = tftypes.NewValue(plan.Raw.Type(), nil)
					}
					response := planmodifier.SetResponse{PlanValue: tc.plan}
					for _, modifier := range modifiers {
						modifier.PlanModifySet(ctx, planmodifier.SetRequest{
							Path: attributePath, ConfigValue: tc.config, PlanValue: response.PlanValue,
							StateValue: tc.state, Plan: plan,
						}, &response)
						require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
					}
					require.True(t, tc.want.Equal(response.PlanValue), "expected %s, got %s", tc.want, response.PlanValue)
				})
			}
		})
	}
}
