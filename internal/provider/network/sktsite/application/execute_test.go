package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/adapter"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

// Every failure has expectations only for the calls preceding it, so an extra mutation fails the test.
func TestCreateOrderingAndFailures(t *testing.T) {
	t.Parallel()
	stages := []string{"add", "native", "range", "exchange", "interface", "general", "configuration", "read_native", "read_interface"}
	for fail := -1; fail < len(stages); fail++ {
		t.Run(func() string {
			if fail < 0 {
				return "success"
			}
			return stages[fail]
		}(), func(t *testing.T) {
			t.Parallel()
			boom := errors.New("failed")
			port := mocks.NewSocketSitePort(t)
			service := application.Service{Port: port, Retry: adapter.Retry{Wait: func(context.Context, time.Duration) error { return nil }}}
			in := application.Input{ConnectionType: "SOCKET_X1700", DesiredIndex: new("INT_5"), InterfaceIndex: "INT_5"}
			var calls []*mock.Call
			step := func(index int) error {
				if fail == index {
					return boom
				}
				return nil
			}
			calls = append(calls, port.EXPECT().Add(mock.Anything, in.Add).Return("site", step(0)).Once())
			if fail != 0 {
				calls = append(calls, port.EXPECT().NativeRange(mock.Anything, "site").Return(&application.Range{NetworkRangeID: "range"}, step(1)).Once())
				if fail != 1 {
					calls = append(calls, port.EXPECT().UpdateRange(mock.Anything, "range", in.Range).Return(step(2)).Once())
					if fail != 2 {
						calls = append(calls, port.EXPECT().Exchange(mock.Anything, "site", "INT_3", "INT_5").Return(step(3)).Once())
						if fail != 3 {
							calls = append(calls, port.EXPECT().UpdateInterface(mock.Anything, "site", "INT_5", in.Interface).Return(step(4)).Once())
							if fail != 4 {
								attempts := 1
								if fail >= 5 {
									attempts = 6
								}
								for n := 0; n < attempts; n++ {
									calls = append(calls, port.EXPECT().General(mock.Anything, "site").Return(&application.Snapshot{ID: "site"}, step(5)).Once())
									if fail != 5 {
										calls = append(calls, port.EXPECT().Configuration(mock.Anything, "site").Return(&application.Configuration{}, step(6)).Once())
										if fail != 6 {
											calls = append(calls, port.EXPECT().NativeRange(mock.Anything, "site").Return(&application.Range{}, step(7)).Once())
											if fail != 7 {
												calls = append(calls, port.EXPECT().DefaultInterface(mock.Anything, "site", "").Return(&application.Interface{}, step(8)).Once())
											}
										}
									}
								}
							}
						}
					}
				}
			}
			mock.InOrder(calls...)
			result, err := service.Create(context.Background(), in)
			if fail >= 0 {
				require.ErrorIs(t, err, boom)
				return
			}
			require.NoError(t, err)
			require.True(t, result.Found)
			require.Equal(t, "site", result.ID)
		})
	}
}
func TestCreatePendingVisibility(t *testing.T) {
	t.Parallel()
	port := mocks.NewSocketSitePort(t)
	port.EXPECT().Add(mock.Anything, mock.Anything).Return("site", nil).Once()
	port.EXPECT().NativeRange(mock.Anything, "site").Return(&application.Range{NetworkRangeID: "range"}, nil).Once()
	port.EXPECT().UpdateRange(mock.Anything, "range", mock.Anything).Return(nil).Once()
	port.EXPECT().UpdateInterface(mock.Anything, "site", mock.Anything, mock.Anything).Return(nil).Once()
	port.EXPECT().General(mock.Anything, "site").Return(nil, nil).Times(6)
	waits := 0
	service := application.Service{Port: port, Retry: adapter.Retry{Wait: func(_ context.Context, d time.Duration) error {
		require.Equal(t, 2*time.Second, d)
		waits++
		return nil
	}}}
	result, err := service.Create(context.Background(), application.Input{})
	require.NoError(t, err)
	require.True(t, result.Pending)
	require.False(t, result.Found)
	require.Equal(t, "site", result.ID)
	require.Equal(t, 6, waits)
}
func TestUpdateHAAndDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Run("gcp_ha", func(t *testing.T) {
		port := mocks.NewSocketSitePort(t)
		in := application.Input{ID: "site", RangeID: "range", IsHA: true, ConnectionType: "SOCKET_GCP1500", Read: application.ReadInput{ID: "site"}}
		first := port.EXPECT().UpdateGeneral(mock.Anything, "site", in.General).Return(nil).Once()
		second := port.EXPECT().UpdateRange(mock.Anything, "range", in.Range).Return(nil).Once()
		third := port.EXPECT().General(mock.Anything, "site").Return(nil, nil).Once()
		mock.InOrder(first, second, third)
		result, err := (application.Service{Port: port}).Update(ctx, in)
		require.NoError(t, err)
		require.False(t, result.Found)
	})
	for _, exists := range []bool{false, true} {
		t.Run(func() string {
			if exists {
				return "delete_existing"
			}
			return "delete_absent"
		}(), func(t *testing.T) {
			port := mocks.NewSocketSitePort(t)
			port.EXPECT().Exists(mock.Anything, "site").Return(exists, nil).Once()
			if exists {
				port.EXPECT().Remove(mock.Anything, "site").Return(nil).Once()
			}
			require.NoError(t, (application.Service{Port: port}).Delete(ctx, "site"))
		})
	}
}

func TestUpdateStopsOnMutationError(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"general", "range", "exchange", "interface"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			port := mocks.NewSocketSitePort(t)
			boom := errors.New("failed")
			in := application.Input{ID: "site", RangeID: "range", CurrentIndex: "LAN1", DesiredIndex: new("LAN2"), InterfaceIndex: "LAN2"}
			outcome := func(stage string) error {
				if stage == failure {
					return boom
				}
				return nil
			}
			calls := []*mock.Call{port.EXPECT().UpdateGeneral(mock.Anything, "site", in.General).Return(outcome("general")).Once()}
			if failure != "general" {
				calls = append(calls, port.EXPECT().UpdateRange(mock.Anything, "range", in.Range).Return(outcome("range")).Once())
				if failure != "range" {
					calls = append(calls, port.EXPECT().Exchange(mock.Anything, "site", "LAN1", "LAN2").Return(outcome("exchange")).Once())
					if failure != "exchange" {
						calls = append(calls, port.EXPECT().UpdateInterface(mock.Anything, "site", "LAN2", in.Interface).Return(outcome("interface")).Once())
					}
				}
			}
			mock.InOrder(calls...)
			_, err := (application.Service{Port: port}).Update(context.Background(), in)
			require.ErrorIs(t, err, boom)
		})
	}
}
func TestProjectionValidationError(t *testing.T) {
	t.Parallel()
	port := mocks.NewSocketSitePort(t)
	boom := errors.New("projection failed")
	port.EXPECT().General(mock.Anything, "site").Return(&application.Snapshot{ID: "site"}, nil).Once()
	port.EXPECT().Configuration(mock.Anything, "site").Return(&application.Configuration{}, nil).Once()
	port.EXPECT().NativeRange(mock.Anything, "site").Return(&application.Range{}, nil).Once()
	port.EXPECT().DefaultInterface(mock.Anything, "site", "").Return(&application.Interface{}, nil).Once()
	service := application.Service{Port: port, ValidateSnapshot: func(*application.Snapshot) error { return boom }}
	_, err := service.Read(context.Background(), application.ReadInput{ID: "site"})
	require.ErrorIs(t, err, boom)
}

func TestUpdateExplicitEmptyDesiredIndex(t *testing.T) {
	t.Parallel()
	port := mocks.NewSocketSitePort(t)
	in := application.Input{ID: "site", RangeID: "range", CurrentIndex: "LAN1", DesiredIndex: new(""), Read: application.ReadInput{ID: "site"}}
	port.EXPECT().UpdateGeneral(mock.Anything, "site", in.General).Return(nil).Once()
	port.EXPECT().UpdateRange(mock.Anything, "range", in.Range).Return(nil).Once()
	port.EXPECT().Exchange(mock.Anything, "site", "LAN1", "").Return(nil).Once()
	port.EXPECT().UpdateInterface(mock.Anything, "site", "", in.Interface).Return(nil).Once()
	port.EXPECT().General(mock.Anything, "site").Return(nil, nil).Once()
	result, err := (application.Service{Port: port}).Update(context.Background(), in)
	require.NoError(t, err)
	require.False(t, result.Found)
}
