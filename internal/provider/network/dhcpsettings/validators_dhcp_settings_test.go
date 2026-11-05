package dhcpsettings

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

func TestRelayGroupPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ref := func(id, name types.String) types.Object {
		return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": id, "name": name})
	}
	null := types.ObjectNull(idname.RefAttrTypes)
	unknown := types.ObjectUnknown(idname.RefAttrTypes)
	state := ref(types.StringValue("group-1"), types.StringValue("Example"))
	byID := ref(types.StringValue("group-1"), types.StringNull())
	byName := ref(types.StringNull(), types.StringValue("Example"))
	planID := ref(types.StringValue("group-1"), types.StringUnknown())
	planName := ref(types.StringUnknown(), types.StringValue("Example"))
	changedID := ref(types.StringValue("group-2"), types.StringUnknown())
	changedName := ref(types.StringUnknown(), types.StringValue("Other"))
	var schemaResponse resource.SchemaResponse
	NewDHCPSettingsResource().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), "%v", schemaResponse.Diagnostics)
	attributePath := path.Root("relay").AtName("group")
	attribute, diags := schemaResponse.Schema.AttributeAtPath(ctx, attributePath)
	require.False(t, diags.HasError(), "%v", diags)
	group := attribute.(schema.SingleNestedAttribute)
	require.NotEmpty(t, group.PlanModifiers)

	for _, tc := range []struct {
		name                      string
		config, plan, state, want types.Object
		destroy                   bool
		wantError                 string
	}{
		{name: "omitted on create", config: null, plan: unknown, state: null, want: unknown},
		{name: "omitted on update retains group", config: null, plan: unknown, state: state, want: state},
		{name: "unknown configuration", config: unknown, plan: unknown, state: state, want: unknown},
		{name: "configured ID on create", config: byID, plan: planID, state: null, want: planID},
		{name: "configured name on create", config: byName, plan: planName, state: null, want: planName},
		{name: "matching ID", config: byID, plan: planID, state: state, want: state},
		{name: "matching name", config: byName, plan: planName, state: state, want: state},
		{name: "changed ID", config: ref(types.StringValue("group-2"), types.StringNull()), plan: changedID, state: state, want: changedID},
		{name: "changed name", config: ref(types.StringNull(), types.StringValue("Other")), plan: changedName, state: state, want: changedName},
		{name: "both ID and name rejected", config: state, plan: state, state: null, want: state, wantError: "only one of"},
		{name: "destroy keeps null plan", config: null, plan: null, state: state, want: null, destroy: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			model := SettingsModel{ID: types.StringValue("account-123"), LeaseTime: types.Int64Value(3600), Relay: types.ObjectValueMust(RelayAttrTypes, map[string]attr.Value{
				"enabled": types.BoolValue(false), "timeout": types.Int64Value(10), "group": tc.plan,
			})}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, model).HasError())
			model.Relay = types.ObjectValueMust(RelayAttrTypes, map[string]attr.Value{
				"enabled": types.BoolValue(false), "timeout": types.Int64Value(10), "group": tc.state,
			})
			state := tfsdk.State{Schema: schemaResponse.Schema}
			require.False(t, state.Set(ctx, model).HasError())
			if tc.state.IsNull() {
				state.Raw = tftypes.NewValue(state.Raw.Type(), nil)
			}
			if tc.destroy {
				plan.Raw = tftypes.NewValue(plan.Raw.Type(), nil)
			}
			resp := planmodifier.ObjectResponse{PlanValue: tc.plan}
			for _, modifier := range group.PlanModifiers {
				modifier.PlanModifyObject(ctx, planmodifier.ObjectRequest{
					Path: attributePath, ConfigValue: tc.config, PlanValue: resp.PlanValue, StateValue: tc.state, Plan: plan, State: state,
				}, &resp)
			}
			if tc.wantError != "" {
				require.True(t, resp.Diagnostics.HasError())
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.wantError)
			} else {
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			}
			require.True(t, tc.want.Equal(resp.PlanValue), "expected %s, got %s", tc.want, resp.PlanValue)
		})
	}
}
