package strule

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestEmptySourceAndDNSExclusionRemainNull(t *testing.T) {
	var rule cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule
	require.NoError(t, json.Unmarshal([]byte(`{"source":{"user":[],"usersGroup":[]},"dnsExclusion":{"domain":[]}}`), &rule))
	state, diags := hydrateSplitTunnelRuleStateFromClient(context.Background(), &rule)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, state.Source.IsNull())
	require.True(t, state.DNSExclusion.IsNull())
}

func TestStateConversionsPreserveConfiguredEmptyFilters(t *testing.T) {
	ctx := context.Background()
	var remote cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule
	require.NoError(t, json.Unmarshal([]byte(`{"id":"remote-id","name":"remote-name","destination":{"domain":["example.com"]}}`), &remote))
	hydrated, diags := hydrateSplitTunnelRuleStateFromClient(ctx, &remote)
	require.False(t, diags.HasError(), "%v", diags)
	emptyRefs := types.SetValueMust(idname.RefObjectType, nil)
	source := types.ObjectValueMust(SplitTunnelSourceAttrTypes, map[string]attr.Value{"user": emptyRefs, "users_group": emptyRefs})
	configured := map[string]attr.Value{
		"country":       emptyRefs,
		"platform":      types.ListValueMust(types.StringType, nil),
		"source":        source,
		"destination":   types.ObjectValueMust(SplitTunnelDestinationAttrTypes, map[string]attr.Value{"application": emptyRefs, "global_ip_range": emptyRefs, "domain": types.ListValueMust(types.StringType, nil), "fqdn": types.ListValueMust(types.StringType, nil)}),
		"dns_exclusion": types.ObjectValueMust(SplitTunnelDNSExclusionAttrTypes, map[string]attr.Value{"domain": types.ListValueMust(types.StringType, nil)}),
	}
	// Preserve the API's populated domain while testing empty sibling selectors.
	destination := configured["destination"].(types.Object).Attributes()
	destination["domain"] = hydrated.Destination.Attributes()["domain"]
	configured["destination"] = types.ObjectValueMust(SplitTunnelDestinationAttrTypes, destination)
	for _, phase := range []string{"create", "read", "update"} {
		t.Run(phase, func(t *testing.T) {
			hydratedObj, d := types.ObjectValueFrom(ctx, SplitTunnelRuleRuleAttrTypes, hydrated)
			require.False(t, d.HasError(), "%v", d)
			attrs := hydratedObj.Attributes()
			for name, value := range configured {
				attrs[name] = value
			}
			var prior SplitTunnelRuleRulePlan
			d = types.ObjectValueMust(SplitTunnelRuleRuleAttrTypes, attrs).As(ctx, &prior, basetypes.ObjectAsOptions{})
			require.False(t, d.HasError(), "%v", d)
			actual := hydrated
			var obj types.Object
			var diagnostics diag.Diagnostics
			switch phase {
			case "create":
				obj, diagnostics = convertCreateInput(ctx, &actual, prior)
			case "read":
				obj, diagnostics = convertReadState(ctx, prior, &actual)
			case "update":
				obj, diagnostics = convertUpdateInput2(ctx, prior, &actual)
			}
			require.False(t, diagnostics.HasError(), "%v", diagnostics)
			for name, expected := range configured {
				require.True(t, expected.Equal(obj.Attributes()[name]), "%s: expected %s, got %s", name, expected, obj.Attributes()[name])
			}
			require.Equal(t, "remote-id", obj.Attributes()["id"].(types.String).ValueString())
		})
	}
}
