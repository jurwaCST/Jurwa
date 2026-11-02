package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeClock is a clock the test moves by hand, so no test waits 10 real seconds.
type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time          { return c.t }
func (c *fakeClock) Advance(d time.Duration) { c.t = c.t.Add(d) }

// recordSleep returns a Sleep func that records each wait instead of waiting.
func recordSleep(waits *[]time.Duration) func(context.Context, time.Duration) error {
	return func(_ context.Context, d time.Duration) error {
		*waits = append(*waits, d)
		return nil
	}
}

var (
	errUnavailable = status.Error(codes.Unavailable, "down")
	errDeadline    = status.Error(codes.DeadlineExceeded, "slow")
	errInvalid     = status.Error(codes.InvalidArgument, "bad input")
	errNotFound    = status.Error(codes.NotFound, "no rule")
)

// ---- timeout ----

func TestWithTimeoutSlowCall(t *testing.T) {
	err := WithTimeout(context.Background(), 10*time.Millisecond, func(ctx context.Context) error {
		<-ctx.Done() // a call slower than the timeout
		return ctx.Err()
	})
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}

func TestWithTimeoutFastCall(t *testing.T) {
	err := WithTimeout(context.Background(), time.Second, func(ctx context.Context) error {
		_, ok := ctx.Deadline()
		assert.True(t, ok, "the call gets a deadline")
		return nil
	})
	assert.NoError(t, err)

	err = WithTimeout(context.Background(), time.Second, func(context.Context) error { return errNotFound })
	assert.Equal(t, codes.NotFound, status.Code(err), "other errors pass through")
}

// ---- retry ----

func TestRetryRetriesOnlyTransientErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCalls int
	}{
		{name: "unavailable is retried", err: errUnavailable, wantCalls: 3},
		{name: "deadline exceeded is retried", err: errDeadline, wantCalls: 3},
		{name: "invalid argument is not retried", err: errInvalid, wantCalls: 1},
		{name: "not found is not retried", err: errNotFound, wantCalls: 1},
		{name: "plain error is not retried", err: errors.New("boom"), wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var waits []time.Duration
			p := Retry{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, Sleep: recordSleep(&waits)}
			calls := 0
			err := p.Do(context.Background(), func(context.Context) error { calls++; return tt.err })
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestRetryWaitsGrow(t *testing.T) {
	var waits []time.Duration
	p := Retry{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, Sleep: recordSleep(&waits)}
	_ = p.Do(context.Background(), func(context.Context) error { return errUnavailable })
	assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}, waits, "no wait after the last attempt")
}

func TestRetryStopsOnSuccess(t *testing.T) {
	var waits []time.Duration
	p := Retry{MaxAttempts: 3, BaseDelay: time.Millisecond, Sleep: recordSleep(&waits)}
	calls := 0
	err := p.Do(context.Background(), func(context.Context) error {
		calls++
		if calls == 1 {
			return errUnavailable
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestRetryStopsWhenCallerGivesUp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := Retry{MaxAttempts: 3, BaseDelay: time.Millisecond}.Do(ctx, func(context.Context) error { calls++; return errUnavailable })
	assert.Error(t, err)
	assert.Equal(t, 1, calls, "no retry after the caller's context is done")
}

func TestRetryDefaults(t *testing.T) {
	calls := 0
	// Zero values mean 1 attempt, so a zero Retry behaves like no retry.
	err := Retry{}.Do(context.Background(), func(context.Context) error { calls++; return errUnavailable })
	assert.Equal(t, errUnavailable, err)
	assert.Equal(t, 1, calls)
}

func TestSleepContext(t *testing.T) {
	assert.NoError(t, sleepContext(context.Background(), time.Millisecond))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, sleepContext(ctx, time.Hour), context.Canceled)
}

// ---- circuit breaker ----

func newTestBreaker() (*Breaker, *fakeClock) {
	clock := &fakeClock{t: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)}
	return NewBreaker(5, 10*time.Second, clock.Now), clock
}

func fail(b *Breaker, n int) {
	for i := 0; i < n; i++ {
		b.Record(errUnavailable)
	}
}

func TestBreakerOpensAfterFiveFailuresInARow(t *testing.T) {
	b, _ := newTestBreaker()
	fail(b, 4)
	assert.Equal(t, Closed, b.State())
	assert.NoError(t, b.Allow())

	fail(b, 1)
	assert.Equal(t, Open, b.State())
	assert.ErrorIs(t, b.Allow(), ErrCircuitOpen, "calls are rejected while open")
	assert.Equal(t, codes.Unavailable, status.Code(b.Allow()))
}

func TestBreakerSuccessResetsTheCount(t *testing.T) {
	b, _ := newTestBreaker()
	fail(b, 4)
	b.Record(nil)
	fail(b, 4)
	assert.Equal(t, Closed, b.State(), "failures must be in a row")
}

func TestBreakerIgnoresBusinessErrors(t *testing.T) {
	b, _ := newTestBreaker()
	for i := 0; i < 10; i++ {
		b.Record(errNotFound)
		b.Record(errInvalid)
	}
	assert.Equal(t, Closed, b.State(), "NotFound and InvalidArgument are answers, not failures")
}

func TestBreakerHalfOpenThenCloses(t *testing.T) {
	b, clock := newTestBreaker()
	fail(b, 5)

	clock.Advance(9 * time.Second)
	assert.ErrorIs(t, b.Allow(), ErrCircuitOpen, "still open before 10 seconds")

	clock.Advance(time.Second)
	assert.Equal(t, HalfOpen, b.State())
	require.NoError(t, b.Allow(), "one trial call is let through")
	assert.ErrorIs(t, b.Allow(), ErrCircuitOpen, "only one trial at a time")

	b.Record(nil)
	assert.Equal(t, Closed, b.State(), "one success closes it")
	assert.NoError(t, b.Allow())
}

func TestBreakerHalfOpenThenReopens(t *testing.T) {
	b, clock := newTestBreaker()
	fail(b, 5)
	clock.Advance(10 * time.Second)
	require.NoError(t, b.Allow())

	b.Record(errDeadline)
	assert.Equal(t, Open, b.State(), "one failure re-opens it")
	assert.ErrorIs(t, b.Allow(), ErrCircuitOpen)

	clock.Advance(10 * time.Second)
	assert.Equal(t, HalfOpen, b.State(), "and it waits another 10 seconds")
}

func TestStateString(t *testing.T) {
	assert.Equal(t, "closed", Closed.String())
	assert.Equal(t, "open", Open.String())
	assert.Equal(t, "half-open", HalfOpen.String())
	assert.Equal(t, "unknown", State(9).String())
}

// ---- interceptor ----

// fakeInvoker returns the errors in order, then nil, and counts the calls.
type fakeInvoker struct {
	errs  []error
	calls int
}

func (f *fakeInvoker) invoke(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
	f.calls++
	if f.calls <= len(f.errs) {
		return f.errs[f.calls-1]
	}
	return nil
}

func testConfig() (Config, *fakeClock) {
	b, clock := newTestBreaker()
	var waits []time.Duration
	return Config{
		Timeout: time.Second,
		Retry:   Retry{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, Sleep: recordSleep(&waits)},
		Breaker: b,
	}, clock
}

func TestInterceptorRetriesThenSucceeds(t *testing.T) {
	cfg, _ := testConfig()
	inv := &fakeInvoker{errs: []error{errUnavailable, errUnavailable}}
	err := UnaryClientInterceptor(cfg)(context.Background(), "/svc/M", nil, nil, nil, inv.invoke)
	assert.NoError(t, err)
	assert.Equal(t, 3, inv.calls)
	assert.Equal(t, Closed, cfg.Breaker.State())
}

func TestInterceptorDoesNotRetryBusinessErrors(t *testing.T) {
	cfg, _ := testConfig()
	inv := &fakeInvoker{errs: []error{errNotFound}}
	err := UnaryClientInterceptor(cfg)(context.Background(), "/svc/M", nil, nil, nil, inv.invoke)
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Equal(t, 1, inv.calls)
}

func TestInterceptorOpensBreakerAndStopsCalling(t *testing.T) {
	cfg, clock := testConfig()
	down := &fakeInvoker{errs: []error{errUnavailable, errUnavailable, errUnavailable, errUnavailable, errUnavailable, errUnavailable}}
	ic := UnaryClientInterceptor(cfg)

	_ = ic(context.Background(), "/svc/M", nil, nil, nil, down.invoke) // 3 failures
	_ = ic(context.Background(), "/svc/M", nil, nil, nil, down.invoke) // 2 more: opens, 3rd attempt rejected
	assert.Equal(t, Open, cfg.Breaker.State())
	assert.Equal(t, 5, down.calls)

	err := ic(context.Background(), "/svc/M", nil, nil, nil, down.invoke)
	assert.ErrorIs(t, err, ErrCircuitOpen)
	assert.Equal(t, 5, down.calls, "open breaker: the service is not called and not retried")

	clock.Advance(10 * time.Second)
	up := &fakeInvoker{}
	assert.NoError(t, ic(context.Background(), "/svc/M", nil, nil, nil, up.invoke))
	assert.Equal(t, Closed, cfg.Breaker.State())
}

func TestInterceptorAppliesTimeout(t *testing.T) {
	cfg, _ := testConfig()
	cfg.Timeout = 10 * time.Millisecond
	cfg.Retry.MaxAttempts = 1
	slow := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		<-ctx.Done()
		return status.FromContextError(ctx.Err()).Err()
	}
	err := UnaryClientInterceptor(cfg)(context.Background(), "/svc/M", nil, nil, nil, slow)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 2*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.Retry.MaxAttempts)
	require.NotNil(t, cfg.Breaker)
	assert.Equal(t, Closed, cfg.Breaker.State())
	assert.NotSame(t, cfg.Breaker, DefaultConfig().Breaker, "each client gets its own breaker")
}
