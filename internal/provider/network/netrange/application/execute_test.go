package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/netrange/application"
)

func TestCreateOrderingAndFailures(t *testing.T) {
	t.Parallel()
	for _, fail := range []string{"", "relay", "interface", "add", "fetch", "hydrate_interface", "hydrate_relay"} {
		t.Run(fail, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			port := mocks.NewNetworkRangePort(t)
			service := application.Service{Port: port}
			kind, name := "VLAN", "relay"
			boom := errors.New("failed")
			in := application.Input{SiteID: "site", InterfaceIndex: "INT_5", RangeType: &kind, DhcpSettings: &dhcp.Settings{DhcpType: "DHCP_RELAY", RelayGroupName: &name}}
			read := application.ReadInput{SiteID: "site", InterfaceIndex: "INT_5", ResolveInterface: true, ResolveRelayName: true}
			var calls []*mock.Call
			step := func(stage string) error {
				if fail == stage {
					return boom
				}
				return nil
			}
			relay := port.EXPECT().RelayID(mock.Anything, "relay").Return("relay-id", step("relay")).Once()
			calls = append(calls, relay)
			if fail != "relay" {
				iface := port.EXPECT().InterfaceID(mock.Anything, "site", "INT_5").Return("interface-id", step("interface")).Once()
				calls = append(calls, iface)
				if fail != "interface" {
					add := port.EXPECT().Add(mock.Anything, "interface-id", mock.MatchedBy(func(i application.Input) bool {
						return i.DhcpSettings.RelayGroupID != nil && *i.DhcpSettings.RelayGroupID == "relay-id"
					})).Return("range-id", step("add")).Once()
					calls = append(calls, add)
					if fail != "add" {
						snapshot := &application.Snapshot{NetworkRangeID: "range-id", DhcpSettings: &dhcp.Settings{DhcpType: "DHCP_RELAY", RelayGroupID: new("relay-id")}}
						fetch := port.EXPECT().Fetch(mock.Anything, "range-id").Return(snapshot, step("fetch")).Once()
						calls = append(calls, fetch)
						if fail != "fetch" {
							iface := port.EXPECT().InterfaceID(mock.Anything, "site", "INT_5").Return("interface-id", step("hydrate_interface")).Once()
							calls = append(calls, iface)
							if fail != "hydrate_interface" {
								lookup := port.EXPECT().RelayName(mock.Anything, "relay-id").Return(new("relay"), step("hydrate_relay")).Once()
								calls = append(calls, lookup)
							}
						}
					}
				}
			}
			mock.InOrder(calls...)
			got, err := service.Create(ctx, in, read)
			if fail != "" {
				require.ErrorIs(t, err, boom)
				return
			}
			require.NoError(t, err)
			require.True(t, got.Found)
			require.Equal(t, "interface-id", got.Snapshot.InterfaceID)
			require.Equal(t, "relay", *got.Snapshot.DhcpSettings.RelayGroupName)
		})
	}
}
func TestAbsentOutcomes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Run("read", func(t *testing.T) {
		port := mocks.NewNetworkRangePort(t)
		port.EXPECT().Fetch(mock.Anything, "id").Return(nil, nil).Once()
		result, err := (application.Service{Port: port}).Read(ctx, application.ReadInput{ID: "id"})
		require.NoError(t, err)
		require.False(t, result.Found)
	})
	t.Run("update", func(t *testing.T) {
		port := mocks.NewNetworkRangePort(t)
		port.EXPECT().Update(mock.Anything, mock.Anything).Return(apperr.ErrAbsent).Once()
		result, err := (application.Service{Port: port}).Update(ctx, application.Input{ID: "id"}, application.ReadInput{ID: "id"})
		require.NoError(t, err)
		require.False(t, result.Found)
	})
	t.Run("delete", func(t *testing.T) {
		port := mocks.NewNetworkRangePort(t)
		port.EXPECT().Remove(mock.Anything, "id").Return(apperr.ErrAbsent).Once()
		require.NoError(t, (application.Service{Port: port}).Delete(ctx, "id"))
	})
}

func TestCreateExplicitEmptyInterfaceID(t *testing.T) {
	t.Parallel()
	port := mocks.NewNetworkRangePort(t)
	// An explicitly empty ID is forwarded just as before; it must not trigger index resolution.
	in := application.Input{InterfaceID: new("")}
	port.EXPECT().Add(mock.Anything, "", in).Return("range", nil).Once()
	port.EXPECT().Fetch(mock.Anything, "range").Return(nil, nil).Once()
	result, err := (application.Service{Port: port}).Create(context.Background(), in, application.ReadInput{})
	require.NoError(t, err)
	require.False(t, result.Found)
}
