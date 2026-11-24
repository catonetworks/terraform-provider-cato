package ifindex

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestBulkIfRuleEnablementLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const resourceType = "cato_bulk_if_move_rule"
	const accountID = "account-123"
	enabledPath := path.Root("rule_data").AtMapKey("Rule A").AtName("enabled")

	api := mocks.NewInternetFirewallBulkPolicyClient(t)
	res := &ifwRulesIndexResource{client: &catoclient.Service{AccountId: accountID}, apiClient: api}
	server := providerserver.NewProtocol6(bulkIfRuleTestProvider{res: res})()
	var schemaResp resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceSchema := schemaResp.Schema
	tfType := resourceSchema.Type().TerraformType(ctx)
	checkDiagnostics := func(diags []*tfprotov6.Diagnostic) {
		t.Helper()
		for _, d := range diags {
			require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
		}
	}

	var sections cato_go_sdk.IfwSectionsIndexPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"internetFirewall":{"policy":{"sections":[{"section":{"id":"s1","name":"Section A"}}]}}}}`), &sections))
	var rules cato_go_sdk.IfwRulesIndexPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"internetFirewall":{"policy":{"rules":[{"rule":{"id":"r1","name":"Rule A","index":1,"enabled":false,"section":{"id":"s1","name":"Section A"}}}]}}}}`), &rules))
	var reordered cato_go_sdk.PolicyInternetFirewallReorderPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"internetFirewall":{"reorderPolicy":{"status":"SUCCESS"}}}}`), &reordered))
	api.EXPECT().PolicyInternetFirewallSectionsIndex(mock.Anything, accountID).Return(&sections, nil).Times(7)
	api.EXPECT().PolicyInternetFirewallRulesIndex(mock.Anything, accountID).Return(&rules, nil).Times(7)
	api.EXPECT().PolicyInternetFirewallReorderPolicy(mock.Anything, mock.Anything, mock.Anything, accountID).
		Return(&reordered, nil).Twice()
	api.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, accountID).Return(nil, nil).Twice()

	dynamic := func(raw tftypes.Value) *tfprotov6.DynamicValue {
		value, err := tfprotov6.NewDynamicValue(tfType, raw)
		require.NoError(t, err)
		return &value
	}
	config := func(enabled types.Bool) *tfprotov6.DynamicValue {
		ruleData, diags := types.MapValueFrom(ctx, IfwRuleIndexResourceObjectTypes, map[string]IfwRulesRuleItemIndex{
			"Rule A": {
				RuleName: types.StringValue("Rule A"), SectionName: types.StringValue("Section A"),
				IndexInSection: types.Int64Value(1), Enabled: enabled,
			},
		})
		require.False(t, diags.HasError(), "%v", diags)
		state := tfsdk.State{Schema: resourceSchema}
		diags = state.Set(ctx, IfwRulesIndex{RuleData: ruleData, SectionData: types.MapNull(IfwSectionIndexResourceObjectTypes)})
		require.False(t, diags.HasError(), "%v", diags)
		return dynamic(state.Raw)
	}
	checkEnabled := func(value *tfprotov6.DynamicValue, expected types.Bool) {
		raw, err := value.Unmarshal(tfType)
		require.NoError(t, err)
		state := tfsdk.State{Schema: resourceSchema, Raw: raw}
		var actual types.Bool
		diags := state.GetAttribute(ctx, enabledPath, &actual)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, expected, actual)
	}

	explicit := config(types.BoolValue(true))
	validation, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
		TypeName: resourceType, Config: explicit,
	})
	require.NoError(t, err)
	checkDiagnostics(validation.Diagnostics)
	require.Len(t, validation.Diagnostics, 1)
	require.Equal(t, tfprotov6.DiagnosticSeverityWarning, validation.Diagnostics[0].Severity)
	require.Contains(t, validation.Diagnostics[0].Detail, "cato_if_rule.rule.enabled")

	empty := dynamic(tftypes.NewValue(tfType, nil))
	createPlan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: resourceType, Config: explicit, ProposedNewState: explicit, PriorState: empty,
	})
	require.NoError(t, err)
	checkDiagnostics(createPlan.Diagnostics)
	checkEnabled(createPlan.PlannedState, types.BoolValue(true))
	created, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: resourceType, Config: explicit, PriorState: empty, PlannedState: createPlan.PlannedState, PlannedPrivate: createPlan.PlannedPrivate,
	})
	require.NoError(t, err)
	checkDiagnostics(created.Diagnostics)
	checkEnabled(created.NewState, types.BoolValue(true))

	read := func(state *tfprotov6.DynamicValue, private []byte) *tfprotov6.ReadResourceResponse {
		refreshed, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{
			TypeName: resourceType, CurrentState: state, Private: private,
		})
		require.NoError(t, err)
		checkDiagnostics(refreshed.Diagnostics)
		return refreshed
	}
	refreshed := read(created.NewState, created.Private)
	checkEnabled(refreshed.NewState, types.BoolValue(true))
	checkEnabled(read(created.NewState, nil).NewState, types.BoolValue(false)) // v2.0.0 state
	noChange, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: resourceType, Config: explicit, PriorState: refreshed.NewState,
		PriorPrivate: refreshed.Private, ProposedNewState: refreshed.NewState,
	})
	require.NoError(t, err)
	checkDiagnostics(noChange.Diagnostics)
	require.Equal(t, refreshed.NewState, noChange.PlannedState)

	// Terraform can reuse prior state in its proposed state when an optional,
	// computed field is omitted and there are no other changes.
	omitted := config(types.BoolNull())
	updatePlan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: resourceType, Config: omitted, PriorState: refreshed.NewState,
		PriorPrivate: refreshed.Private, ProposedNewState: refreshed.NewState,
	})
	require.NoError(t, err)
	checkDiagnostics(updatePlan.Diagnostics)
	checkEnabled(updatePlan.PlannedState, types.BoolUnknown())
	updated, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: resourceType, Config: omitted, PriorState: refreshed.NewState,
		PlannedState: updatePlan.PlannedState, PlannedPrivate: updatePlan.PlannedPrivate,
	})
	require.NoError(t, err)
	checkDiagnostics(updated.Diagnostics)
	checkEnabled(updated.NewState, types.BoolValue(false))
	require.Empty(t, updated.Private)
	rules.Policy.InternetFirewall.Policy.Rules[0].Rule.Enabled = true
	checkEnabled(read(updated.NewState, updated.Private).NewState, types.BoolValue(true))
}

type bulkIfRuleTestProvider struct {
	provider.Provider
	res resource.Resource
}

func (p bulkIfRuleTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cato"
}

func (p bulkIfRuleTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{func() resource.Resource { return p.res }}
}
