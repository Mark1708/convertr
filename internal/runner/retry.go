package runner

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"git.mark1708.ru/me/convertr/internal/backend"
)

// RetryOpts configures exponential backoff retry behaviour.
type RetryOpts struct {
	MaxAttempts int           // total attempts (1 = no retry, 0 defaults to DefaultRetry)
	BaseDelay   time.Duration // delay before second attempt
	MaxDelay    time.Duration
	Multiplier  float64 // delay multiplier per attempt (e.g. 2.0)
}

// DefaultRetry is the default retry configuration used with ErrorPolicyRetry.
var DefaultRetry = RetryOpts{
	MaxAttempts: 3,
	BaseDelay:   200 * time.Millisecond,
	MaxDelay:    5 * time.Second,
	Multiplier:  2.0,
}

// withRetry calls fn up to opts.MaxAttempts times, retrying only on ErrConvertFail.
func withRetry(ctx context.Context, opts RetryOpts, fn func() error) error {
	if opts.MaxAttempts <= 1 {
		return fn()
	}
	delay := opts.BaseDelay
	var err error
	for i := 0; i < opts.MaxAttempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if !errors.Is(err, backend.ErrConvertFail) {
			return err
		}
		if i == opts.MaxAttempts-1 {
			break
		}
		// jitter: ±25 % of delay
		jitter := time.Duration(rand.Int63n(int64(delay)/2) - int64(delay)/4) //nolint:gosec
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay + jitter):
		}
		next := time.Duration(float64(delay) * opts.Multiplier)
		if next > opts.MaxDelay {
			next = opts.MaxDelay
		}
		delay = next
	}
	return err
}
