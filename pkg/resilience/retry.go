package resilience

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Retry runs a call again when it fails with a temporary error.
type Retry struct {
	MaxAttempts int           // total attempts including the first; 0 or 1 means no retry
	BaseDelay   time.Duration // wait before the 2nd attempt; doubles each time
	// Sleep waits d or until ctx ends. nil means a real sleep; tests inject a fake.
	Sleep func(ctx context.Context, d time.Duration) error
	// Retryable decides which errors are tried again. nil means the package
	// function Retryable (Unavailable or DeadlineExceeded).
	Retryable func(error) bool
}

// Retryable reports whether err is temporary: Unavailable or DeadlineExceeded.
// Business answers such as InvalidArgument or NotFound are never retried, and
// neither is an open circuit breaker.
func Retryable(err error) bool {
	if errors.Is(err, ErrCircuitOpen) {
		return false
	}
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded:
		return true
	}
	return false
}

// RetryableUnavailableOnly is Retryable for calls that create or change
// something: only Unavailable, when the request never reached the service. After
// DeadlineExceeded the first attempt may have worked, so a retry could make a
// duplicate or a false conflict.
func RetryableUnavailableOnly(err error) bool {
	return Retryable(err) && status.Code(err) == codes.Unavailable
}

// Do calls fn until it succeeds, returns an error that is not retryable, the
// attempts run out or ctx ends. It returns the last error of fn.
func (r Retry) Do(ctx context.Context, fn func(context.Context) error) error {
	sleep := r.Sleep
	if sleep == nil {
		sleep = sleepContext
	}
	retryable := r.Retryable
	if retryable == nil {
		retryable = Retryable
	}
	attempts := max(r.MaxAttempts, 1)
	delay := r.BaseDelay

	var err error
	for attempt := 1; ; attempt++ {
		err = fn(ctx)
		if err == nil || !retryable(err) || attempt == attempts || ctx.Err() != nil {
			return err
		}
		if sleep(ctx, delay) != nil {
			return err
		}
		delay *= 2
	}
}

// sleepContext waits d, or returns early with ctx's error when ctx ends.
func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
