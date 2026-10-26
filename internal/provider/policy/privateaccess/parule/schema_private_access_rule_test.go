package parule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestPrivateAccessReferenceSetDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaResponse := resource.SchemaResponse{}
	NewPrivAccessRuleResource().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), "%v", schemaResponse.Diagnostics)
	null := types.SetNull(idname.RefObjectType)
	unknown := types.SetUnknown(idname.RefObjectType)
	empty := types.SetValueMust(idname.RefObjectType, nil)
	value := types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id": types.StringValue("reference-1"), "name": types.StringValue("Example"),
		}),
	})
	paths := []path.Path{
		path.Root("countries"),
		path.Root("devices"),
		path.Root("source").AtName("users"),
		path.Root("source").AtName("user_groups"),
		path.Root("tracking").AtName("alert").AtName("mailing_list"),
		path.Root("tracking").AtName("alert").AtName("subscription_group"),
		path.Root("tracking").AtName("alert").AtName("webhook"),
	}
	for _, attributePath := range paths {
		t.Run(attributePath.String(), func(t *testing.T) {
			t.Parallel()
			attribute, diags := schemaResponse.Schema.AttributeAtPath(ctx, attributePath)
			require.False(t, diags.HasError(), "%v", diags)
			setAttribute, ok := attribute.(schema.SetNestedAttribute)
			require.True(t, ok, "expected a nested reference set, got %T", attribute)
			require.True(t, setAttribute.Optional)
			require.True(t, setAttribute.Computed, "defaults must be allowed to replace omitted values")
			for _, tc := range []struct {
				name                      string
				config, plan, state, want types.Set
				destroy                   bool
			}{
				{name: "omitted on create", config: null, plan: unknown, state: null, want: empty},
				{name: "removed on update", config: null, plan: unknown, state: value, want: empty},
				{name: "configured reference", config: value, plan: value, state: value, want: value},
				{name: "unknown configuration", config: unknown, plan: unknown, state: value, want: unknown},
				{name: "destroy", config: null, plan: null, state: value, want: null, destroy: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					raw := tftypes.NewValue(tftypes.String, "resource")
					if tc.destroy {
						raw = tftypes.NewValue(tftypes.String, nil)
					}
					response := planmodifier.SetResponse{PlanValue: tc.plan}
					for _, modifier := range setAttribute.PlanModifiers {
						modifier.PlanModifySet(ctx, planmodifier.SetRequest{
							Path: attributePath, ConfigValue: tc.config, PlanValue: response.PlanValue,
							StateValue: tc.state, Plan: tfsdk.Plan{Raw: raw},
						}, &response)
						require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
					}
					require.True(t, tc.want.Equal(response.PlanValue), "expected %s, got %s", tc.want, response.PlanValue)
				})
			}
		})
	}
}
