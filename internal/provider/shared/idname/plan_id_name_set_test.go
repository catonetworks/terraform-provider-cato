package idname

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestPlanModifySet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ref := func(id, name types.String) types.Object {
		return types.ObjectValueMust(RefAttrTypes, map[string]attr.Value{"id": id, "name": name})
	}
	set := func(refs ...attr.Value) types.Set { return types.SetValueMust(RefObjectType, refs) }
	nameA := ref(types.StringNull(), types.StringValue("ref-a"))
	nameB := ref(types.StringNull(), types.StringValue("ref-b"))
	idA := ref(types.StringValue("id-a"), types.StringNull())
	idB := ref(types.StringValue("id-b"), types.StringNull())
	resolvedA := ref(types.StringValue("id-a"), types.StringValue("ref-a"))
	resolvedB := ref(types.StringValue("id-b"), types.StringValue("ref-b"))
	planNameA := ref(types.StringUnknown(), types.StringValue("ref-a"))
	planNameB := ref(types.StringUnknown(), types.StringValue("ref-b"))
	planIDA := ref(types.StringValue("id-a"), types.StringUnknown())
	planIDB := ref(types.StringValue("id-b"), types.StringUnknown())
	nullSet := types.SetNull(RefObjectType)
	unknownSet := types.SetUnknown(RefObjectType)
	unknownObject := types.ObjectUnknown(RefAttrTypes)
	resourceValue := func(t *testing.T, value types.Set, nullResource bool) tftypes.Value {
		t.Helper()
		object := types.ObjectValueMust(
			map[string]attr.Type{"references": value.Type(ctx)},
			map[string]attr.Value{"references": value},
		)
		raw, err := object.ToTerraformValue(ctx)
		require.NoError(t, err)
		if nullResource {
			return tftypes.NewValue(raw.Type(), nil)
		}
		return raw
	}

	for _, tc := range []struct {
		name                      string
		config, plan, state, want types.Set
		wantError                 string
		destroy                   bool
	}{
		{
			name:   "matching name and ID retain their counterparts",
			config: set(nameA, idB), plan: set(planNameA, planIDB), state: set(resolvedA, resolvedB), want: set(resolvedA, resolvedB),
		},
		{
			name:   "create by name leaves ID unknown",
			config: set(nameA), plan: set(planNameA), state: nullSet, want: set(planNameA),
		},
		{
			name:   "create by ID leaves name unknown",
			config: set(idA), plan: set(planIDA), state: nullSet, want: set(planIDA),
		},
		{
			name:   "changed name replaces the old reference",
			config: set(nameB), plan: set(planNameB), state: set(resolvedA), want: set(planNameB),
		},
		{
			name:   "changed ID replaces the old reference",
			config: set(idB), plan: set(planIDB), state: set(resolvedA), want: set(planIDB),
		},
		{
			name:   "added reference preserves existing matches",
			config: set(nameA, nameB), plan: set(planNameA, planNameB), state: set(resolvedA), want: set(resolvedA, planNameB),
		},
		{
			name:   "removed reference stays removed",
			config: set(nameB), plan: set(planNameB), state: set(resolvedA, resolvedB), want: set(resolvedB),
		},
		{
			name:   "duplicate display names allow explicit IDs",
			config: set(idA, idB), plan: set(planIDA, planIDB),
			state: set(resolvedA, ref(types.StringValue("id-b"), types.StringValue("ref-a"))),
			want:  set(resolvedA, ref(types.StringValue("id-b"), types.StringValue("ref-a"))),
		},
		{
			name:   "duplicate display names allow removing an explicit ID",
			config: set(idB), plan: set(planIDB),
			state: set(resolvedA, ref(types.StringValue("id-b"), types.StringValue("ref-a"))),
			want:  set(ref(types.StringValue("id-b"), types.StringValue("ref-a"))),
		},
		{
			name:   "unrelated duplicate display names do not block name references",
			config: set(nameB), plan: set(planNameB),
			state: set(resolvedA, ref(types.StringValue("id-c"), types.StringValue("ref-a")), resolvedB),
			want:  set(resolvedB),
		},
		{
			name:   "ambiguous name references require an explicit ID",
			config: set(nameA), plan: set(planNameA),
			state: set(resolvedA, ref(types.StringValue("id-b"), types.StringValue("ref-a")), ref(types.StringValue("id-c"), types.StringValue("ref-a"))),
			want:  set(planNameA), wantError: "Duplicate name",
		},
		{
			name:   "empty configuration produces a known empty set",
			config: set(), plan: set(), state: set(resolvedA), want: set(),
		},
		{
			name:   "null configuration produces an empty set during ordinary planning",
			config: nullSet, plan: nullSet, state: set(resolvedA), want: set(),
		},
		{
			name:   "unknown configuration remains unknown",
			config: unknownSet, plan: unknownSet, state: set(resolvedA), want: unknownSet,
		},
		{
			name:   "unknown object preserves the incoming plan",
			config: set(unknownObject), plan: set(unknownObject), state: set(resolvedA), want: set(unknownObject),
		},
		{
			name:   "unknown object after a known reference preserves the whole plan",
			config: set(nameA, unknownObject), plan: set(planNameA, unknownObject), state: set(resolvedA), want: set(planNameA, unknownObject),
		},
		{
			name:   "unknown name preserves the incoming plan",
			config: set(ref(types.StringNull(), types.StringUnknown())), plan: set(unknownObject), state: set(resolvedA), want: set(unknownObject),
		},
		{
			name:   "unknown ID preserves the incoming plan",
			config: set(ref(types.StringUnknown(), types.StringNull())), plan: set(unknownObject), state: set(resolvedA), want: set(unknownObject),
		},
		{
			name:   "destroy keeps the null plan",
			config: set(nameA), plan: nullSet, state: set(resolvedA), want: nullSet,
			destroy: true,
		},
		{
			name:   "destroy with null configuration keeps the null plan",
			config: nullSet, plan: nullSet, state: set(resolvedA), want: nullSet,
			destroy: true,
		},
		{
			name:   "destroy with empty configuration keeps the null plan",
			config: set(), plan: nullSet, state: set(resolvedA), want: nullSet,
			destroy: true,
		},
		{
			name:   "both name and ID are rejected",
			config: set(resolvedA), plan: set(resolvedA), state: nullSet, want: set(resolvedA), wantError: "only one of",
		},
		{
			name:   "missing name and ID are rejected",
			config: set(ref(types.StringNull(), types.StringNull())), plan: set(unknownObject), state: nullSet, want: set(unknownObject),
			wantError: "'name' or 'id' must be defined",
		},
		{
			name:   "empty name is rejected",
			config: set(ref(types.StringNull(), types.StringValue(""))), plan: set(unknownObject), state: nullSet, want: set(unknownObject),
			wantError: "'name' must not be empty",
		},
		{
			name:   "empty ID is rejected",
			config: set(ref(types.StringValue(""), types.StringNull())), plan: set(unknownObject), state: nullSet, want: set(unknownObject),
			wantError: "'id' must not be empty",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.SetRequest{
				Path: path.Root("references"), ConfigValue: tc.config, PlanValue: tc.plan, StateValue: tc.state,
				Plan:  tfsdk.Plan{Raw: resourceValue(t, tc.plan, tc.destroy)},
				State: tfsdk.State{Raw: resourceValue(t, tc.state, tc.state.IsNull())},
			}
			resp := planmodifier.SetResponse{PlanValue: req.PlanValue}
			SetPlanModifier().PlanModifySet(ctx, req, &resp)
			if tc.wantError != "" {
				require.True(t, resp.Diagnostics.HasError(), "expected diagnostic containing %q", tc.wantError)
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.wantError)
			} else {
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			}
			require.True(t, resp.PlanValue.Equal(tc.want), "got %s, want %s", resp.PlanValue, tc.want)
		})
	}
}
