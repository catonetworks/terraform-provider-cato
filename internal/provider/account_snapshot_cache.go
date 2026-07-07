package provider

import (
	"context"
	"slices"
	"strings"
	"sync"

	cato "github.com/catonetworks/cato-go-sdk"
)

const accountSnapshotMaxConcurrentRequests = 1

// accountSnapshotCache is a provider-process cache for AccountSnapshot responses.
//
// Terraform refresh can run many resource reads concurrently. For WAN interfaces, those reads
// often ask for the same site snapshot and can otherwise create a burst of AccountSnapshot API
// calls that intermittently fails before the plan completes. The cache lets reads proceed
// concurrently, coalesces same-key in-flight requests, and serializes actual outbound fetches.
// Mutating paths can force refresh to bypass cached data and repopulate the cache with a fresh
// snapshot for post-apply state hydration.
type accountSnapshotCache struct {
	mu       sync.Mutex
	values   map[string]*cato.AccountSnapshot
	inflight map[string]*accountSnapshotCall
	limit    chan struct{}
}

type accountSnapshotCall struct {
	done chan struct{}
	resp *cato.AccountSnapshot
	err  error
}

func newAccountSnapshotCache() *accountSnapshotCache {
	return &accountSnapshotCache{
		values:   map[string]*cato.AccountSnapshot{},
		inflight: map[string]*accountSnapshotCall{},
		limit:    make(chan struct{}, accountSnapshotMaxConcurrentRequests),
	}
}

func accountSnapshotCacheKey(accountID string, siteIDs, userIDs []string) string {
	siteIDs = sortedStringCopy(siteIDs)
	userIDs = sortedStringCopy(userIDs)

	return strings.Join([]string{
		accountID,
		strings.Join(siteIDs, ","),
		strings.Join(userIDs, ","),
	}, "|")
}

func (c *accountSnapshotCache) get(
	ctx context.Context,
	key string,
	forceRefresh bool,
	fetch func(context.Context) (*cato.AccountSnapshot, error),
) (*cato.AccountSnapshot, error) {
	if c == nil {
		return fetch(ctx)
	}

	c.mu.Lock()
	if !forceRefresh {
		if cached, ok := c.values[key]; ok {
			c.mu.Unlock()
			return cached, nil
		}
	}
	if forceRefresh {
		c.mu.Unlock()
		resp, err := c.fetchWithLimit(ctx, fetch)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.values[key] = resp
		c.mu.Unlock()
		return resp, nil
	}
	if call, ok := c.inflight[key]; ok {
		c.mu.Unlock()
		select {
		case <-call.done:
			return call.resp, call.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	call := &accountSnapshotCall{done: make(chan struct{})}
	c.inflight[key] = call
	c.mu.Unlock()

	call.resp, call.err = c.fetchWithLimit(ctx, fetch)

	c.mu.Lock()
	if call.err == nil {
		c.values[key] = call.resp
	}
	delete(c.inflight, key)
	close(call.done)
	c.mu.Unlock()

	return call.resp, call.err
}

func sortedStringCopy(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	copied := append([]string(nil), values...)
	slices.Sort(copied)
	return copied
}

func (d *catoClientData) accountSnapshot(
	ctx context.Context,
	siteIDs []string,
	userIDs []string,
	forceRefresh bool,
) (*cato.AccountSnapshot, error) {
	key := accountSnapshotCacheKey(d.AccountId, siteIDs, userIDs)
	fetch := func(ctx context.Context) (*cato.AccountSnapshot, error) {
		return d.catov2.AccountSnapshot(ctx, siteIDs, userIDs, &d.AccountId)
	}

	// forceRefresh in get already bypasses and overwrites the cached value, so no
	// separate invalidation is required here.
	return d.accountSnapshotCache.get(ctx, key, forceRefresh, fetch)
}

func (c *accountSnapshotCache) fetchWithLimit(
	ctx context.Context,
	fetch func(context.Context) (*cato.AccountSnapshot, error),
) (*cato.AccountSnapshot, error) {
	select {
	case c.limit <- struct{}{}:
		defer func() { <-c.limit }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return fetch(ctx)
}
