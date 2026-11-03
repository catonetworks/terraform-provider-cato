package sbrule

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestEmptySiteAndServiceRemainNull(t *testing.T) {
	var rule cato_go_sdk.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"site":[],"group":[]},"service":{"simple":[],"custom":[]}}`), &rule))
	state, diags := hydrateSocketBypassRuleStateFromClient(context.Background(), &rule)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, state.Site.IsNull())
	require.True(t, state.Service.IsNull())
}
