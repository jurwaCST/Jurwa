package resilience

import (
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrCircuitOpen is returned without calling the service while the breaker is
// open. Its code is Unavailable, so callers treat it like a service that is down.
var ErrCircuitOpen = status.Error(codes.Unavailable, "circuit breaker is open")

// State is the state of a Breaker.
type State int

const (
	Closed   State = iota // calls go through
	Open                  // calls are rejected
	HalfOpen              // one trial call may go through
)

func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return "unknown"
}

// Breaker stops calling a service that keeps failing, so a caller fails fast
// instead of waiting on every request. It is safe for concurrent use.
type Breaker struct {
	threshold int           // failures in a row that open the breaker
	openFor   time.Duration // how long it stays open before a trial call
	now       func() time.Time

	mu       sync.Mutex
	state    State
	failures int
	openedAt time.Time
	trial    bool // a half-open trial call is in flight
}

// NewBreaker returns a closed breaker. now is time.Now outside tests.
func NewBreaker(threshold int, openFor time.Duration, now func() time.Time) *Breaker {
	return &Breaker{threshold: threshold, openFor: openFor, now: now}
}

// State returns the current state. An open breaker whose wait is over reports HalfOpen.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.currentState()
}

func (b *Breaker) currentState() State {
	if b.state == Open && !b.now().Before(b.openedAt.Add(b.openFor)) {
		return HalfOpen
	}
	return b.state
}

// Allow returns nil when a call may go ahead, or ErrCircuitOpen. A caller that
// was allowed must report the result with Record.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.currentState() {
	case Closed:
		return nil
	case HalfOpen:
		if b.trial {
			return ErrCircuitOpen
		}
		b.state, b.trial = HalfOpen, true
		return nil
	default:
		return ErrCircuitOpen
	}
}

// Record reports the result of an allowed call.
func (b *Breaker) Record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == Open {
		return // a late result from before the breaker opened
	}
	if !isFailure(err) {
		b.state, b.failures, b.trial = Closed, 0, false
		return
	}
	b.failures++
	if b.state == HalfOpen || b.failures >= b.threshold {
		b.state, b.openedAt, b.trial = Open, b.now(), false
	}
}

// isFailure reports whether err means the service is unhealthy. Answers such as
// NotFound or InvalidArgument show the service is working, so they do not count.
func isFailure(err error) bool {
	if err == nil {
		return false
	}
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.ResourceExhausted:
		return true
	}
	return false
}
