package alwaysonindex_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonindex"
)

// Exercise framework planning with unchanged ordering and no pending draft changes.
func TestAlwaysOnRuleBulkPublishTriggerPlan(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name           string
		prior, trigger types.String
		changed        bool
	}{
		{name: "content changed", prior: types.StringValue("hash-before"), trigger: types.StringValue("hash-after"), changed: true},
		{name: "unchanged", prior: types.StringValue("hash-before"), trigger: types.StringValue("hash-before")},
		{name: "added to existing state", prior: types.StringNull(), trigger: types.StringValue("hash-after"), changed: true},
		{name: "omitted", prior: types.StringNull(), trigger: types.StringNull()},
		{name: "removed", prior: types.StringValue("hash-before"), trigger: types.StringNull(), changed: true},
		{name: "unknown until apply", prior: types.StringValue("hash-before"), trigger: types.StringUnknown(), changed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var schemaResp resource.SchemaResponse
			alwaysonindex.NewAlwaysOnRuleBulkResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			sections, diags := types.MapValueFrom(ctx, alwaysonindex.SectionObjectType, map[string]alwaysonindex.Section{
				"Alpha": {ID: types.StringValue("s1"), Index: types.Int64Value(1)},
			})
			require.False(t, diags.HasError(), "%v", diags)
			rules, diags := types.MapValueFrom(ctx, alwaysonindex.RuleObjectType, map[string]alwaysonindex.Rule{
				"First": {ID: types.StringValue("r1"), IndexInSection: types.Int64Value(1), SectionName: types.StringValue("Alpha")},
			})
			require.False(t, diags.HasError(), "%v", diags)
			model := alwaysonindex.RulesIndex{
				SectionData: sections, RuleData: rules, Changes: types.Int64Value(0), PublishTrigger: tc.prior,
			}
			state := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, state.Set(ctx, model).HasError())
			prior, err := tfprotov6.NewDynamicValue(state.Raw.Type(), state.Raw)
			require.NoError(t, err)

			model.PublishTrigger = tc.trigger
			proposed := tfsdk.Plan{Schema: schemaResp.Schema}
			require.False(t, proposed.Set(ctx, model).HasError())
			proposedValue, err := tfprotov6.NewDynamicValue(proposed.Raw.Type(), proposed.Raw)
			require.NoError(t, err)
			config := tfsdk.Config{Schema: schemaResp.Schema}
			model.Changes = types.Int64Null()
			model.SectionData, diags = types.MapValueFrom(ctx, alwaysonindex.SectionObjectType, map[string]alwaysonindex.Section{
				"Alpha": {ID: types.StringNull(), Index: types.Int64Value(1)},
			})
			require.False(t, diags.HasError(), "%v", diags)
			model.RuleData, diags = types.MapValueFrom(ctx, alwaysonindex.RuleObjectType, map[string]alwaysonindex.Rule{
				"First": {ID: types.StringNull(), IndexInSection: types.Int64Value(1), SectionName: types.StringValue("Alpha")},
			})
			require.False(t, diags.HasError(), "%v", diags)
			configState := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, configState.Set(ctx, model).HasError())
			config.Raw = configState.Raw
			configValue, err := tfprotov6.NewDynamicValue(config.Raw.Type(), config.Raw)
			require.NoError(t, err)

			server := providerserver.NewProtocol6(provider.New("test")())()
			resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName: "cato_always_on_rule_bulk", PriorState: &prior, Config: &configValue, ProposedNewState: &proposedValue,
			})
			require.NoError(t, err)
			for _, diagnostic := range resp.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
			}
			require.Empty(t, resp.RequiresReplace)
			plannedRaw, err := resp.PlannedState.Unmarshal(state.Raw.Type())
			require.NoError(t, err)
			require.Equal(t, tc.changed, !plannedRaw.Equal(state.Raw))
			planned := tfsdk.Plan{Schema: schemaResp.Schema, Raw: plannedRaw}
			var actual alwaysonindex.RulesIndex
			require.False(t, planned.Get(ctx, &actual).HasError())
			require.Equal(t, tc.trigger, actual.PublishTrigger)
			require.Equal(t, types.Int64Value(0), actual.Changes)
		})
	}
}
