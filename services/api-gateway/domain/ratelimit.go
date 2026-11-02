package domain

import (
	"sync"
	"time"
)

// Limiter is an in-memory token bucket per key (the client IP). Each bucket
// holds at most perSecond tokens and refills at perSecond tokens a second, so a
// client gets perSecond requests in a burst, then perSecond a second. Redis
// counters (ratelimit:*) replace it in Sprint 3 so all gateway pods share limits.
type Limiter struct {
	rate float64 // tokens per second, also the bucket size
	now  func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

// maxBuckets bounds memory: past it, buckets that are full again are dropped.
const maxBuckets = 10000

// NewLimiter returns a limiter of perSecond requests a second per key. now is
// time.Now outside tests.
func NewLimiter(perSecond int, now func() time.Time) *Limiter {
	return &Limiter{rate: float64(perSecond), now: now, buckets: make(map[string]*bucket)}
}

// Allow takes one token from key's bucket. It returns false when the bucket is empty.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()

	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) >= maxBuckets {
			l.dropFull(now)
		}
		b = &bucket{tokens: l.rate, last: now}
		l.buckets[key] = b
	}
	b.tokens = min(l.rate, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// dropFull removes buckets that have refilled completely; they carry no state.
func (l *Limiter) dropFull(now time.Time) {
	for k, b := range l.buckets {
		if b.tokens+now.Sub(b.last).Seconds()*l.rate >= l.rate {
			delete(l.buckets, k)
		}
	}
}
