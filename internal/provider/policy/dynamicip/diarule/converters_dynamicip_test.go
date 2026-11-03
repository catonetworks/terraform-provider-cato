package diarule

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestEmptySourceRemainsNull(t *testing.T) {
	var rule cato_go_sdk.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule
	require.NoError(t, json.Unmarshal([]byte(`{"source":{"user":[],"usersGroup":[]}}`), &rule))
	state, diags := hydrateDynamicIPAllocationRuleStateFromClient(context.Background(), &rule)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, state.Source.IsNull())
}
