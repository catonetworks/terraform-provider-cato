package idname

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestPlanModifyList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ref := func(id, name types.String) types.Object {
		return types.ObjectValueMust(RefAttrTypes, map[string]attr.Value{"id": id, "name": name})
	}
	list := func(refs ...attr.Value) types.List { return types.ListValueMust(RefObjectType, refs) }
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
	nullList := types.ListNull(RefObjectType)
	unknownList := types.ListUnknown(RefObjectType)
	unknownObject := types.ObjectUnknown(RefAttrTypes)
	nullObject := types.ObjectNull(RefAttrTypes)
	invalidList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("invalid reference")})
	resourceValue := func(t *testing.T, value types.List, nullResource bool) tftypes.Value {
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
		config, plan, state, want types.List
		wantError                 string
		destroy                   bool
	}{
		{
			name:   "reordered references follow configuration order",
			config: list(nameB, nameA), plan: list(planNameB, planNameA), state: list(resolvedA, resolvedB), want: list(resolvedB, resolvedA),
		},
		{
			name:   "name and ID aliases retain both list positions",
			config: list(nameA, idA), plan: list(planNameA, planIDA), state: list(resolvedA), want: list(resolvedA, resolvedA),
		},
		{
			name:   "repeated references are preserved on creation",
			config: list(nameA, nameA), plan: list(planNameA, planNameA), state: nullList, want: list(planNameA, planNameA),
		},
		{
			name:   "repeated state references remain valid on subsequent plans",
			config: list(nameA, nameA), plan: list(planNameA, planNameA), state: list(resolvedA, resolvedA), want: list(resolvedA, resolvedA),
		},
		{
			name:   "null state elements are discarded",
			config: list(nameA), plan: list(planNameA), state: list(nullObject, resolvedA), want: list(resolvedA),
		},
		{
			name:   "null configured references are discarded while preserving order",
			config: list(nullObject, nameA, nullObject, nameB, nullObject),
			plan:   list(nullObject, planNameA, nullObject, planNameB, nullObject), state: list(resolvedB, resolvedA), want: list(resolvedA, resolvedB),
		},
		{
			name:   "all null configured references produce a known empty list",
			config: list(nullObject, nullObject), plan: list(nullObject, nullObject), state: nullList, want: list(),
		},
		{
			name:   "all null configuration does not require state lookups",
			config: list(nullObject), plan: list(nullObject),
			state: list(resolvedA, ref(types.StringValue("id-b"), types.StringValue("ref-a"))), want: list(),
		},
		{
			name:   "all null state leaves counterparts unresolved",
			config: list(nameA), plan: list(planNameA), state: list(nullObject, nullObject), want: list(planNameA),
		},
		{
			name:   "nulls and duplicate references preserve the surviving order and count",
			config: list(nullObject, nameB, nullObject, idA, nameB), plan: list(nullObject, planNameB, nullObject, planIDA, planNameB),
			state: list(nullObject, resolvedB, resolvedA, resolvedB), want: list(resolvedB, resolvedA, resolvedB),
		},
		{
			name:   "unknown elements prevent null filtering from modifying the plan",
			config: list(nullObject, unknownObject, nameA), plan: list(nullObject, unknownObject, planNameA),
			state: list(resolvedA), want: list(nullObject, unknownObject, planNameA),
		},
		{
			name:   "unknown configuration leaves a known incoming plan unchanged",
			config: unknownList, plan: list(resolvedB), state: list(resolvedA), want: list(resolvedB),
		},
		{
			name:   "unknown object before a known reference preserves the whole plan",
			config: list(unknownObject, nameA), plan: list(unknownObject, planNameA), state: list(resolvedA), want: list(unknownObject, planNameA),
		},
		{
			name:   "unknown field after a known reference preserves the whole plan",
			config: list(nameA, ref(types.StringUnknown(), types.StringNull())),
			plan:   list(planNameA, unknownObject), state: list(resolvedA), want: list(planNameA, unknownObject),
		},
		{
			name:   "unknown state leaves counterparts unresolved",
			config: list(nameA, idB), plan: list(planNameA, planIDB), state: unknownList, want: list(planNameA, planIDB),
		},
		{
			name:   "incomplete state entries do not supply counterparts",
			config: list(nameA, idB), plan: list(planNameA, planIDB),
			state: list(nameA, idB), want: list(planNameA, planIDB),
		},
		{
			name:   "ambiguous state names report diagnostics",
			config: list(nameA), plan: list(planNameA), state: list(resolvedA, ref(types.StringValue("id-b"), types.StringValue("ref-a"))),
			want: list(planNameA), wantError: "Duplicate name",
		},
		{
			name:   "ambiguous state IDs report diagnostics",
			config: list(idA), plan: list(planIDA), state: list(resolvedA, ref(types.StringValue("id-a"), types.StringValue("ref-b"))),
			want: list(planIDA), wantError: "Duplicate id",
		},
		{
			name:   "configuration conversion errors reach the response",
			config: invalidList, plan: list(planNameA), state: nullList, want: list(planNameA), wantError: "cannot",
		},
		{
			name:   "state conversion errors reach the response",
			config: list(nameA), plan: list(planNameA), state: invalidList, want: list(planNameA), wantError: "cannot",
		},
		{
			name:   "matching name and ID retain their counterparts",
			config: list(nameA, idB), plan: list(planNameA, planIDB), state: list(resolvedA, resolvedB), want: list(resolvedA, resolvedB),
		},
		{
			name:   "create by name leaves ID unknown",
			config: list(nameA), plan: list(planNameA), state: nullList, want: list(planNameA),
		},
		{
			name:   "create by ID leaves name unknown",
			config: list(idA), plan: list(planIDA), state: nullList, want: list(planIDA),
		},
		{
			name:   "changed name replaces the old reference",
			config: list(nameB), plan: list(planNameB), state: list(resolvedA), want: list(planNameB),
		},
		{
			name:   "changed ID replaces the old reference",
			config: list(idB), plan: list(planIDB), state: list(resolvedA), want: list(planIDB),
		},
		{
			name:   "added reference preserves existing matches",
			config: list(nameA, nameB), plan: list(planNameA, planNameB), state: list(resolvedA), want: list(resolvedA, planNameB),
		},
		{
			name:   "removed reference stays removed",
			config: list(nameB), plan: list(planNameB), state: list(resolvedA, resolvedB), want: list(resolvedB),
		},
		{
			name:   "empty configuration produces a known empty list",
			config: list(), plan: list(), state: list(resolvedA), want: list(),
		},
		{
			name:   "null configuration produces an empty list during ordinary planning",
			config: nullList, plan: nullList, state: list(resolvedA), want: list(),
		},
		{
			name:   "unknown configuration remains unknown",
			config: unknownList, plan: unknownList, state: list(resolvedA), want: unknownList,
		},
		{
			name:   "unknown object preserves the incoming plan",
			config: list(unknownObject), plan: list(unknownObject), state: list(resolvedA), want: list(unknownObject),
		},
		{
			name:   "unknown object after a known reference preserves the whole plan",
			config: list(nameA, unknownObject), plan: list(planNameA, unknownObject), state: list(resolvedA), want: list(planNameA, unknownObject),
		},
		{
			name:   "unknown name preserves the incoming plan",
			config: list(ref(types.StringNull(), types.StringUnknown())), plan: list(unknownObject), state: list(resolvedA), want: list(unknownObject),
		},
		{
			name:   "unknown ID preserves the incoming plan",
			config: list(ref(types.StringUnknown(), types.StringNull())), plan: list(unknownObject), state: list(resolvedA), want: list(unknownObject),
		},
		{
			name:   "destroy keeps the null plan",
			config: list(nameA), plan: nullList, state: list(resolvedA), want: nullList,
			destroy: true,
		},
		{
			name:   "destroy with null configuration keeps the null plan",
			config: nullList, plan: nullList, state: list(resolvedA), want: nullList,
			destroy: true,
		},
		{
			name:   "destroy with empty configuration keeps the null plan",
			config: list(), plan: nullList, state: list(resolvedA), want: nullList,
			destroy: true,
		},
		{
			name:   "both name and ID are rejected",
			config: list(resolvedA), plan: list(resolvedA), state: nullList, want: list(resolvedA), wantError: "only one of",
		},
		{
			name:   "missing name and ID are rejected",
			config: list(ref(types.StringNull(), types.StringNull())), plan: list(unknownObject), state: nullList, want: list(unknownObject),
			wantError: "'name' or 'id' must be defined",
		},
		{
			name:   "empty name is rejected",
			config: list(ref(types.StringNull(), types.StringValue(""))), plan: list(unknownObject), state: nullList, want: list(unknownObject),
			wantError: "'name' must not be empty",
		},
		{
			name:   "empty ID is rejected",
			config: list(ref(types.StringValue(""), types.StringNull())), plan: list(unknownObject), state: nullList, want: list(unknownObject),
			wantError: "'id' must not be empty",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.ListRequest{
				Path: path.Root("references"), ConfigValue: tc.config, PlanValue: tc.plan, StateValue: tc.state,
				Plan:  tfsdk.Plan{Raw: resourceValue(t, tc.plan, tc.destroy)},
				State: tfsdk.State{Raw: resourceValue(t, tc.state, tc.state.IsNull())},
			}
			resp := planmodifier.ListResponse{PlanValue: req.PlanValue}
			require.NotPanics(t, func() {
				ListPlanModifier().PlanModifyList(ctx, req, &resp)
			})
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

func TestListPlanModifierDescriptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	modifier := ListPlanModifier()
	require.NotEmpty(t, modifier.Description(ctx))
	require.Equal(t, modifier.Description(ctx), modifier.MarkdownDescription(ctx))
}

func TestListPlanModifierHelpersRejectInvalidReferences(t *testing.T) {
	t.Parallel()
	modifier := idNameListPlanModifier{}
	for _, tc := range []struct {
		name    string
		handler func(*Ref, map[string]string, string, *diag.Diagnostics) Ref
		input   *Ref
	}{
		{name: "nil name reference", handler: modifier.handleName},
		{name: "nil ID reference", handler: modifier.handleID},
		{
			name: "ID reference containing a name", handler: modifier.handleID,
			input: &Ref{ID: types.StringValue("id-a"), Name: types.StringValue("ref-a")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diagnostics diag.Diagnostics
			require.NotPanics(t, func() {
				tc.handler(tc.input, nil, "references", &diagnostics)
			})
			require.True(t, diagnostics.HasError(), "invalid references must produce diagnostics")
		})
	}
}
