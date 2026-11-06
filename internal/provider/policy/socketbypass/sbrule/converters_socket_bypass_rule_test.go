package sbrule

import (
	"context"
	"encoding/json"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestStateConversionsPreserveConfiguredEmptyFilters(t *testing.T) {
	ctx := context.Background()
	var remote cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule
	require.NoError(t, json.Unmarshal([]byte(`{"id":"rule-id","name":"rule","source":{"ip":["198.18.0.230"]},"destination":{"ip":["192.0.2.230"]}}`), &remote))
	hydrated, diags := hydrateSocketBypassRuleStateFromClient(ctx, &remote)
	require.False(t, diags.HasError(), "%v", diags)
	hydratedObj, diags := types.ObjectValueFrom(ctx, SocketBypassRuleRuleAttrTypes, hydrated)
	require.False(t, diags.HasError(), "%v", diags)
	attrs := hydratedObj.Attributes()
	for _, name := range []string{"site", "source", "destination", "service"} {
		objectType := SocketBypassRuleRuleAttrTypes[name].(types.ObjectType)
		empty := emptyFilterObject(objectType.AttrTypes).Attributes()
		if name == "source" || name == "destination" {
			empty["ip"] = attrs[name].(types.Object).Attributes()["ip"]
		}
		attrs[name] = types.ObjectValueMust(objectType.AttrTypes, empty)
	}
	attrs["exception"] = types.ListValueMust(types.ObjectType{AttrTypes: SocketBypassExceptionAttrTypes}, nil)
	var prior SocketBypassRuleRulePlan
	diags = types.ObjectValueMust(SocketBypassRuleRuleAttrTypes, attrs).As(ctx, &prior, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), "%v", diags)
	for _, phase := range []string{"create", "read", "update"} {
		t.Run(phase, func(t *testing.T) {
			actual := hydrated
			var result types.Object
			var diagnostics diag.Diagnostics
			switch phase {
			case "create":
				result, diagnostics = convertCreateInput(ctx, &actual, prior)
			case "read":
				result, diagnostics = convertReadState(ctx, prior, &actual)
			case "update":
				result, diagnostics = convertUpdateInput2(ctx, prior, &actual)
			}
			require.False(t, diagnostics.HasError(), "%v", diagnostics)
			for _, name := range []string{"site", "source", "destination", "service", "exception"} {
				require.True(t, attrs[name].Equal(result.Attributes()[name]), "%s changed: %s", name, result.Attributes()[name])
			}
		})
	}
}

func emptyFilterObject(attrTypes map[string]attr.Type) types.Object {
	values := make(map[string]attr.Value, len(attrTypes))
	for name, attrType := range attrTypes {
		switch typed := attrType.(type) {
		case types.ListType:
			values[name] = types.ListValueMust(typed.ElemType, nil)
		case types.SetType:
			values[name] = types.SetValueMust(typed.ElemType, nil)
		case types.ObjectType:
			values[name] = emptyFilterObject(typed.AttrTypes)
		default:
			values[name] = types.StringNull()
		}
	}
	return types.ObjectValueMust(attrTypes, values)
}
