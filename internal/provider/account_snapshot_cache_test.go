package provider

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
)

func TestAccountSnapshotCacheReusesSuccessfulResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cache := newAccountSnapshotCache()
	var calls atomic.Int64

	fetch := func(context.Context) (*cato.AccountSnapshot, error) {
		calls.Add(1)
		return &cato.AccountSnapshot{}, nil
	}

	first, err := cache.get(ctx, "account|site-1|", false, fetch)
	if err != nil {
		t.Fatalf("unexpected first fetch error: %v", err)
	}
	second, err := cache.get(ctx, "account|site-1|", false, fetch)
	if err != nil {
		t.Fatalf("unexpected second fetch error: %v", err)
	}

	if first != second {
		t.Fatal("expected cached response pointer to be reused")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one API call, got %d", got)
	}
}

func TestAccountSnapshotCacheForceRefreshReplacesCachedResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cache := newAccountSnapshotCache()
	var calls atomic.Int64

	fetch := func(context.Context) (*cato.AccountSnapshot, error) {
		calls.Add(1)
		return &cato.AccountSnapshot{}, nil
	}

	first, err := cache.get(ctx, "account|site-1|", false, fetch)
	if err != nil {
		t.Fatalf("unexpected first fetch error: %v", err)
	}
	second, err := cache.get(ctx, "account|site-1|", true, fetch)
	if err != nil {
		t.Fatalf("unexpected refresh fetch error: %v", err)
	}

	if first == second {
		t.Fatal("expected force refresh to replace cached response")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected two API calls, got %d", got)
	}
}

func TestAccountSnapshotCacheDoesNotCacheErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cache := newAccountSnapshotCache()
	var calls atomic.Int64
	wantErr := errors.New("snapshot failed")

	fetch := func(context.Context) (*cato.AccountSnapshot, error) {
		if calls.Add(1) == 1 {
			return nil, wantErr
		}
		return &cato.AccountSnapshot{}, nil
	}

	if _, err := cache.get(ctx, "account|site-1|", false, fetch); !errors.Is(err, wantErr) {
		t.Fatalf("expected first fetch error %v, got %v", wantErr, err)
	}
	if _, err := cache.get(ctx, "account|site-1|", false, fetch); err != nil {
		t.Fatalf("expected retry after error to succeed, got %v", err)
	}

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected failed call not to be cached, got %d calls", got)
	}
}

func TestAccountSnapshotCacheCoalescesConcurrentRequests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cache := newAccountSnapshotCache()
	var calls atomic.Int64
	started := make(chan struct{})
	release := make(chan struct{})

	fetch := func(context.Context) (*cato.AccountSnapshot, error) {
		calls.Add(1)
		close(started)
		<-release
		return &cato.AccountSnapshot{}, nil
	}

	const waiters = 8
	var wg sync.WaitGroup
	wg.Add(waiters)
	errs := make(chan error, waiters)
	for range waiters {
		go func() {
			defer wg.Done()
			_, err := cache.get(ctx, "account|site-1|", false, fetch)
			errs <- err
		}()
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first fetch")
	}
	close(release)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected coalesced fetch error: %v", err)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one coalesced API call, got %d", got)
	}
}

func TestAccountSnapshotCacheSerializesDifferentFetches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cache := newAccountSnapshotCache()
	var active atomic.Int64
	var maxActive atomic.Int64
	release := make(chan struct{})

	fetch := func(context.Context) (*cato.AccountSnapshot, error) {
		cur := active.Add(1)
		for {
			prev := maxActive.Load()
			if cur <= prev || maxActive.CompareAndSwap(prev, cur) {
				break
			}
		}
		<-release
		active.Add(-1)
		return &cato.AccountSnapshot{}, nil
	}

	const requests = 4
	var wg sync.WaitGroup
	wg.Add(requests)
	for i := range requests {
		go func(i int) {
			defer wg.Done()
			_, err := cache.get(ctx, accountSnapshotCacheKey("account", []string{string(rune('a' + i))}, nil), false, fetch)
			if err != nil {
				t.Errorf("unexpected fetch error: %v", err)
			}
		}(i)
	}

	for range requests {
		release <- struct{}{}
	}
	wg.Wait()

	if got := maxActive.Load(); got != 1 {
		t.Fatalf("expected only one active fetch at a time, got %d", got)
	}
}

func TestAccountSnapshotCacheKeySortsIDs(t *testing.T) {
	t.Parallel()

	first := accountSnapshotCacheKey("account", []string{"site-b", "site-a"}, []string{"user-b", "user-a"})
	second := accountSnapshotCacheKey("account", []string{"site-a", "site-b"}, []string{"user-a", "user-b"})

	if first != second {
		t.Fatalf("expected key to be order-stable, got %q and %q", first, second)
	}
}
