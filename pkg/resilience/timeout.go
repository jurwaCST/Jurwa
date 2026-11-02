package resilience

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithTimeout runs fn with a context that ends after d. When the call is slower
// than d it returns a gRPC DeadlineExceeded error. d <= 0 means no timeout.
func WithTimeout(ctx context.Context, d time.Duration, fn func(context.Context) error) error {
	if d <= 0 {
		return fn(ctx)
	}
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	err := fn(ctx)
	if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return status.Errorf(codes.DeadlineExceeded, "call timed out after %s", d)
	}
	return err
}
