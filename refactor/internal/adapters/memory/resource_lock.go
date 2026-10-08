package memory

import (
	"context"
	"sync"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

var _ network.RevisionCoordinator = (*ResourceLock)(nil)

type entry struct {
	token chan struct{}
	users int
}

// ResourceLock is a context-aware, process-local lock for private-access policies.
// Share one instance across participating resources/provider aliases for the same
// API endpoint. External processes and nonparticipating writers are not locked.
// Its zero value is ready for use.
type ResourceLock struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func NewResourceLock() *ResourceLock {
	return &ResourceLock{}
}

func (c *ResourceLock) Acquire(ctx context.Context, accountID string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.entries == nil {
		c.entries = make(map[string]*entry)
	}
	lock := c.entries[accountID]
	if lock == nil {
		lock = &entry{token: make(chan struct{}, 1)}
		c.entries[accountID] = lock
	}
	lock.users++
	c.mu.Unlock()

	select {
	case <-ctx.Done():
		c.drop(accountID, lock)
		return nil, ctx.Err()
	case lock.token <- struct{}{}:
	}
	var once sync.Once
	release := func() {
		once.Do(func() {
			<-lock.token
			c.drop(accountID, lock)
		})
	}
	if err := ctx.Err(); err != nil {
		release()
		return nil, err
	}
	return release, nil
}

func (c *ResourceLock) drop(accountID string, lock *entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	lock.users--
	if lock.users == 0 {
		delete(c.entries, accountID)
	}
}
