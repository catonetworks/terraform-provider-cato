package socketlannetworkrule

import (
	"context"
	"encoding/json"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestHydrateSocketLanNetworkRuleEmptyReferences(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, response string }{
		{name: "omitted API collections", response: `{}`},
		{name: "empty API collections", response: `{
			"site":{"site":[],"group":[]},
			"source":{"host":[],"group":[],"systemGroup":[],"networkInterface":[],"globalIpRange":[],"floatingSubnet":[],"siteNetworkSubnet":[]},
			"destination":{"host":[],"group":[],"systemGroup":[],"networkInterface":[],"globalIpRange":[],"floatingSubnet":[],"siteNetworkSubnet":[]}
		}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var rule cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
			require.NoError(t, json.Unmarshal([]byte(tc.response), &rule))
			empty := types.SetValueMust(idname.RefObjectType, nil)
			for name, object := range map[string]types.Object{
				"site":        hydrateSocketLanSiteState(ctx, &rule),
				"source":      hydrateSocketLanSourceState(ctx, rule.Source),
				"destination": hydrateSocketLanDestinationState(ctx, rule.Destination),
			} {
				for field, value := range object.Attributes() {
					if value.Type(ctx).Equal(types.SetType{ElemType: idname.RefObjectType}) {
						require.True(t, empty.Equal(value), "%s.%s must match the planned empty set, got %s", name, field, value)
					}
				}
			}
		})
	}
}
