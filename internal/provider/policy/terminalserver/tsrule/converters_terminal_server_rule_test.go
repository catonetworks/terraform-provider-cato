package tsrule

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

func TestStateConversionsPreserveConfiguredEmptyFilters(t *testing.T) {
	for _, emptyRefs := range []types.Set{types.SetNull(idname.RefObjectType), types.SetValueMust(idname.RefObjectType, nil)} {
		ctx := context.Background()
		var remote cato_go_sdk.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Rules_Rule
		require.NoError(t, json.Unmarshal([]byte(`{"id":"remote-id","name":"remote-name"}`), &remote))
		hydrated, diags := hydrateTerminalServerRuleStateFromClient(ctx, &remote)
		require.False(t, diags.HasError(), "%v", diags)
		configured := map[string]attr.Value{"exclude_traffic": emptyRefs}
		for _, phase := range []string{"create", "read", "update"} {
			t.Run(phase, func(t *testing.T) {
				hydratedObj, d := types.ObjectValueFrom(ctx, TerminalServerRuleRuleAttrTypes, hydrated)
				require.False(t, d.HasError(), "%v", d)
				attrs := hydratedObj.Attributes()
				for name, value := range configured {
					attrs[name] = value
				}
				var prior TerminalServerRuleRulePlan
				d = types.ObjectValueMust(TerminalServerRuleRuleAttrTypes, attrs).As(ctx, &prior, basetypes.ObjectAsOptions{})
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
}
