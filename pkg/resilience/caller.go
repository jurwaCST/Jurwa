package resilience

import "context"

// Caller is one call that Retry may run again. Retry.Do takes a function, and a
// Caller's method value (c.Call) is such a function. The interface lets tests
// check with a gomock mock exactly how many times a call was made.
// Regenerate the mock with: go generate ./pkg/resilience/
//
//go:generate mockgen -source=caller.go -destination=mocks/caller_mock.go -package=mocks
type Caller interface {
	Call(ctx context.Context) error
}
