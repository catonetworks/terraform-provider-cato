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

func TestAlwaysOnRuleBulkOmittedRuleDataPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	alwaysonindex.NewAlwaysOnRuleBulkResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	sections, diags := types.MapValueFrom(ctx, alwaysonindex.SectionObjectType, map[string]alwaysonindex.Section{
		"Alpha": {ID: types.StringValue("s1"), Index: types.Int64Value(1)},
	})
	require.False(t, diags.HasError(), "%v", diags)
	configuredSections, diags := types.MapValueFrom(ctx, alwaysonindex.SectionObjectType, map[string]alwaysonindex.Section{
		"Alpha": {ID: types.StringNull(), Index: types.Int64Value(1)},
	})
	require.False(t, diags.HasError(), "%v", diags)
	empty, diags := types.MapValueFrom(ctx, alwaysonindex.RuleObjectType, map[string]alwaysonindex.Rule{})
	require.False(t, diags.HasError(), "%v", diags)
	populated, diags := types.MapValueFrom(ctx, alwaysonindex.RuleObjectType, map[string]alwaysonindex.Rule{
		"First": {ID: types.StringValue("r1"), IndexInSection: types.Int64Value(1), SectionName: types.StringValue("Alpha")},
	})
	require.False(t, diags.HasError(), "%v", diags)
	dynamic := func(t *testing.T, model alwaysonindex.RulesIndex) tfprotov6.DynamicValue {
		t.Helper()
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, model).HasError())
		value, err := tfprotov6.NewDynamicValue(state.Raw.Type(), state.Raw)
		require.NoError(t, err)
		return value
	}
	for _, tc := range []struct {
		name           string
		create         bool
		prior          types.Map
		omitSections   bool
		triggerChanged bool
	}{
		{name: "omitted on create", create: true, prior: types.MapNull(alwaysonindex.RuleObjectType)},
		{name: "still omitted", prior: empty},
		{name: "publish only", prior: populated, omitSections: true},
		{name: "publish intent changes", prior: populated, omitSections: true, triggerChanged: true},
		{name: "previously populated then omitted", prior: populated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			model := alwaysonindex.RulesIndex{
				SectionData: sections, RuleData: tc.prior, Changes: types.Int64Value(0), PublishTrigger: types.StringNull(),
			}
			prior := dynamic(t, model)
			if tc.create {
				prior = tfprotov6.DynamicValue{JSON: []byte("null")}
			}
			// Terraform carries prior computed values into its proposed state when configuration is omitted.
			proposedModel := model
			if tc.triggerChanged {
				proposedModel.PublishTrigger = types.StringValue("changed")
			}
			proposed := dynamic(t, proposedModel)
			if tc.triggerChanged {
				model.PublishTrigger = proposedModel.PublishTrigger
			}
			model.SectionData = configuredSections
			if tc.omitSections {
				model.SectionData = types.MapNull(alwaysonindex.SectionObjectType)
			}
			model.RuleData = types.MapNull(alwaysonindex.RuleObjectType)
			model.Changes = types.Int64Null()
			config := dynamic(t, model)
			server := providerserver.NewProtocol6(provider.New("test")())()
			resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName: "cato_always_on_rule_bulk", PriorState: &prior, Config: &config, ProposedNewState: &proposed,
			})
			require.NoError(t, err)
			for _, diagnostic := range resp.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
			}
			require.Empty(t, resp.RequiresReplace)
			raw, err := resp.PlannedState.Unmarshal(schemaResp.Schema.Type().TerraformType(ctx))
			require.NoError(t, err)
			planned := tfsdk.Plan{Schema: schemaResp.Schema, Raw: raw}
			var actual alwaysonindex.RulesIndex
			require.False(t, planned.Get(ctx, &actual).HasError())
			if tc.create || tc.triggerChanged {
				require.True(t, actual.RuleData.IsUnknown(), "omitted rule_data is resolved from the live policy at apply")
			} else {
				require.True(t, tc.prior.Equal(actual.RuleData), "omission must preserve the previously observed rules")
			}
		})
	}
}
