package strule

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestEmptySourceAndDNSExclusionRemainNull(t *testing.T) {
	var rule cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule
	require.NoError(t, json.Unmarshal([]byte(`{"source":{"user":[],"usersGroup":[]},"dnsExclusion":{"domain":[]}}`), &rule))
	state, diags := hydrateSplitTunnelRuleStateFromClient(context.Background(), &rule)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, state.Source.IsNull())
	require.True(t, state.DNSExclusion.IsNull())
}
