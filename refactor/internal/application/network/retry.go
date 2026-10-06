package network

import (
	"context"
	"errors"
	"time"
)

// RetryPolicy bounds revision-conflict retries and published-view polling.
// Transport failures are not blindly replayed: a mutation may have succeeded.
type RetryPolicy struct {
	Attempts int
	Delay    time.Duration
}

func DefaultRetryPolicy() RetryPolicy {
	const attempts = 3
	return RetryPolicy{Attempts: attempts, Delay: time.Second}
}

func (r RetryPolicy) conflict(ctx context.Context, operation func() error) error {
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := operation()
		if !errors.Is(err, ErrRevisionConflict) || attempt+1 >= r.Attempts {
			return err
		}
		if err := r.wait(ctx); err != nil {
			return err
		}
	}
}

func (r RetryPolicy) wait(ctx context.Context) error {
	timer := time.NewTimer(r.Delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
