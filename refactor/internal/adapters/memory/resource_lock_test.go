package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCoordinatorCancellationAndAccountIsolation(t *testing.T) {
	t.Parallel()
	var coordinator ResourceLock
	release, err := coordinator.Acquire(context.Background(), "account-a")
	require.NoError(t, err)
	otherRelease, err := coordinator.Acquire(context.Background(), "account-b")
	require.NoError(t, err)
	otherRelease()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		unlock, acquireErr := coordinator.Acquire(ctx, "account-a")
		if acquireErr == nil {
			unlock()
		}
		result <- acquireErr
	}()
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("lock wait ignored cancellation")
	}
	release()
	release() // release is safe to call twice
	require.Empty(t, coordinator.entries)
}

func TestCoordinatorSerializesWritersAndReclaimsEntries(t *testing.T) {
	t.Parallel()
	var coordinator ResourceLock
	var active atomic.Int32
	var group sync.WaitGroup
	const workers = 32
	for range workers {
		group.Go(func() {
			release, err := coordinator.Acquire(context.Background(), "account-a")
			if err != nil {
				t.Error(err)
				return
			}
			if active.Add(1) != 1 {
				t.Error("concurrent writer entered")
			}
			active.Add(-1)
			release()
		})
	}
	group.Wait()
	require.Empty(t, coordinator.entries)
}
