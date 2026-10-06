package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestPrepareExistingPayloadRules(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"Native", "VLAN", "Routed"} {
		t.Run(kind, func(t *testing.T) {
			port := mocks.NewNetworkRangePort(t)
			input, err := application.Prepare(context.Background(), port, kind, &application.Settings{DhcpType: "DHCP_RELAY", RelayGroupID: new("configured"), RelayGroupName: new("name"), DhcpMicrosegmentation: new(true)})
			require.NoError(t, err)
			if kind == "Routed" {
				require.Nil(t, input)
				return
			}
			require.Nil(t, input.RelayGroupID)
			require.Nil(t, input.DhcpMicrosegmentation)
		})
	}
	t.Run("resolve_name", func(t *testing.T) {
		port := mocks.NewNetworkRangePort(t)
		port.EXPECT().RelayID(mock.Anything, "relay").Return("resolved", nil).Once()
		input, err := application.Prepare(context.Background(), port, "Native", &application.Settings{DhcpType: "DHCP_RELAY", RelayGroupName: new("relay")})
		require.NoError(t, err)
		require.Equal(t, "resolved", *input.RelayGroupID)
	})
	t.Run("missing_name", func(t *testing.T) {
		port := mocks.NewNetworkRangePort(t)
		_, err := application.Prepare(context.Background(), port, "Native", &application.Settings{DhcpType: "DHCP_RELAY"})
		require.Error(t, err)
	})
}
