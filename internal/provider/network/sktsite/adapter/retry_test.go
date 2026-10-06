package adapter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

func TestHydrationRetry(t *testing.T) {
	t.Parallel()
	boom := errors.New("fetch failed")
	for _, mode := range []string{"visible", "absent", "error", "recovery"} {
		t.Run(mode, func(t *testing.T) {
			attempts, waits := 0, 0
			r := Retry{Wait: func(context.Context, time.Duration) error { waits++; return nil }}
			result, err := r.Run(context.Background(), application.CreateHydrationPolicy, func() (application.Result, error) {
				attempts++
				switch mode {
				case "visible":
					return application.Result{Found: true}, nil
				case "error":
					return application.Result{}, boom
				case "recovery":
					if attempts < 3 {
						return application.Result{}, boom
					}
					return application.Result{Found: true}, nil
				default:
					return application.Result{}, nil
				}
			})
			switch mode {
			case "visible":
				require.NoError(t, err)
				require.True(t, result.Found)
				require.Equal(t, 1, attempts)
				require.Zero(t, waits)
			case "absent":
				require.NoError(t, err)
				require.Equal(t, 6, attempts)
				require.Equal(t, 6, waits)
			case "error":
				require.ErrorIs(t, err, boom)
				require.Equal(t, 6, attempts)
				require.Equal(t, 5, waits)
			case "recovery":
				require.NoError(t, err)
				require.Equal(t, 3, attempts)
				require.Equal(t, 2, waits)
			}
		})
	}
}
func TestRetryCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (Retry{}).Run(ctx, application.CreateHydrationPolicy, func() (application.Result, error) {
		t.Fatal("attempt after cancellation")
		return application.Result{}, nil
	})
	require.ErrorIs(t, err, context.Canceled)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	_, err = (Retry{Wait: func(ctx context.Context, _ time.Duration) error { cancel(); return waitForRetry(ctx, time.Hour) }}).Run(ctx, application.CreateHydrationPolicy, func() (application.Result, error) { return application.Result{}, nil })
	require.ErrorIs(t, err, context.Canceled)
}
