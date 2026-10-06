package adapter

import (
	"context"
	"time"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

// Retry runs only hydration attempts. SDK HTTP retries remain configured by the provider.
type Retry struct {
	Wait func(context.Context, time.Duration) error
}

func (r Retry) Run(
	ctx context.Context,
	policy application.RetryPolicy,
	attempt func() (
		application.Result,
		error,
	),
) (
	application.Result,
	error,
) {
	wait := r.Wait
	if wait == nil {
		wait = waitForRetry
	}
	var result application.Result
	for n := 0; n < policy.Attempts; n++ {
		if err := ctx.Err(); err != nil {
			return application.Result{}, err
		}
		next, err := attempt()
		result = next
		if err != nil && n == policy.Attempts-1 {
			return application.Result{}, err
		}
		if err == nil && result.Found {
			return result, nil
		}
		// Preserve the delay after an absent final response as well as delays between attempts.
		if err = wait(ctx, policy.Delay); err != nil {
			return application.Result{}, err
		}
	}
	return result, nil
}
func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
