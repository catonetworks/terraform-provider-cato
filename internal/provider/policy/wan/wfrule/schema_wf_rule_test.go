package wfrule_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
)

func TestWanFwRulePlansSourceGroupByName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	wfrule.NewWanFwRuleResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	value := func(description, groupName, groupID string) *tfprotov6.DynamicValue {
		return &tfprotov6.DynamicValue{JSON: fmt.Appendf(nil, `{
			"at": {"position": "LAST_IN_POLICY"},
			"rule": {
				"name": "example-rule", "description": %q, "enabled": true,
				"action": "ALLOW", "direction": "BOTH",
				"source": {"users_group": [{"name": %q, "id": %s}]},
				"destination": {}, "application": {},
				"tracking": {"event": {"enabled": true}}
			}
		}`, description, groupName, groupID)}
	}
	prior := value("previous", "old-group", `"old-id"`)

	for _, tc := range []struct {
		name, groupName, wantBy, wantInput string
	}{
		{name: "group changes", groupName: "new-group", wantBy: "NAME", wantInput: "new-group"},
		{name: "same group retains ID", groupName: "old-group", wantBy: "ID", wantInput: "old-id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			config := value("updated", tc.groupName, "null")
			server := providerserver.NewProtocol6(provider.New("test")())()
			resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName: "cato_wf_rule", PriorState: prior, Config: config, ProposedNewState: config,
			})
			require.NoError(t, err)
			for _, diagnostic := range resp.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
			}
			require.Empty(t, resp.RequiresReplace)
			raw, err := resp.PlannedState.Unmarshal(schemaResp.Schema.Type().TerraformType(ctx))
			require.NoError(t, err)
			planned := tfsdk.Plan{Schema: schemaResp.Schema, Raw: raw}
			var plan wfrule.Rule
			diags := planned.Get(ctx, &plan)
			require.False(t, diags.HasError(), "%v", diags)

			input, diags := wfrule.HydrateWanRuleAPI(ctx, plan)
			require.False(t, diags.HasError(), "%v", diags)
			groups := input.Update.Rule.Source.UsersGroup
			require.Len(t, groups, 1)
			require.Equal(t, tc.wantBy, string(groups[0].By))
			require.Equal(t, tc.wantInput, groups[0].Input)
		})
	}
}
