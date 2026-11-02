package checker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deadline = time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)

// fakeTimers hands out the queued batches one call at a time.
type fakeTimers struct {
	batches [][]Lapsed
	err     error
	calls   int
	limits  []int
}

func (f *fakeTimers) LapseDue(_ context.Context, limit int) ([]Lapsed, error) {
	f.calls++
	f.limits = append(f.limits, limit)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.batches) == 0 {
		return nil, nil
	}
	b := f.batches[0]
	f.batches = f.batches[1:]
	return b, nil
}

// fakeStatus escalates every report except those listed in skip or fail.
type fakeStatus struct {
	skip      map[string]bool
	fail      map[string]bool
	escalated []string
	reasons   []string
}

func (f *fakeStatus) Escalate(_ context.Context, reportID, reason string) error {
	if f.skip[reportID] {
		return fmt.Errorf("report resolved meanwhile: %w", ErrNotEscalatable)
	}
	if f.fail[reportID] {
		return errors.New("status-lifecycle unavailable")
	}
	f.escalated = append(f.escalated, reportID)
	f.reasons = append(f.reasons, reason)
	return nil
}

func lapsed(ids ...string) []Lapsed {
	out := make([]Lapsed, 0, len(ids))
	for _, id := range ids {
		out = append(out, Lapsed{ReportID: id, Deadline: deadline})
	}
	return out
}

func TestRun_NothingDue(t *testing.T) {
	timers := &fakeTimers{}
	status := &fakeStatus{}
	res, err := New(timers, status).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, Result{}, res)
	assert.Equal(t, 1, timers.calls)
	assert.Empty(t, status.escalated)
}

func TestRun_EscalatesEachLapsedReport(t *testing.T) {
	timers := &fakeTimers{batches: [][]Lapsed{lapsed("r1", "r2", "r3", "r4")}}
	status := &fakeStatus{skip: map[string]bool{"r2": true}, fail: map[string]bool{"r3": true}}

	res, err := New(timers, status).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, Result{Lapsed: 4, Escalated: 2, Skipped: 1, Errors: 1}, res)
	assert.Equal(t, []string{"r1", "r4"}, status.escalated, "one failure does not stop the others")
	assert.Equal(t, "SLA lapsed: deadline was 2026-11-03T09:00:00Z", status.reasons[0])
}

func TestRun_FetchesBatchesUntilAShortOne(t *testing.T) {
	full := make([]string, BatchSize)
	for i := range full {
		full[i] = fmt.Sprintf("r%d", i)
	}
	timers := &fakeTimers{batches: [][]Lapsed{lapsed(full...), lapsed("last")}}
	status := &fakeStatus{}

	res, err := New(timers, status).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, BatchSize+1, res.Escalated)
	assert.Equal(t, 2, timers.calls, "a short batch means nothing is left")
	assert.Equal(t, []int{BatchSize, BatchSize}, timers.limits)
}

func TestRun_StopsAfterMaxBatches(t *testing.T) {
	full := make([]string, BatchSize)
	for i := range full {
		full[i] = fmt.Sprintf("r%d", i)
	}
	timers := &fakeTimers{}
	for i := 0; i < MaxBatches+3; i++ {
		timers.batches = append(timers.batches, lapsed(full...))
	}
	res, err := New(timers, &fakeStatus{}).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, MaxBatches, timers.calls, "one run is bounded; the next run continues")
	assert.Equal(t, MaxBatches*BatchSize, res.Lapsed)
}

func TestRun_SlaTrackingDown(t *testing.T) {
	timers := &fakeTimers{err: errors.New("connection refused")}
	res, err := New(timers, &fakeStatus{}).Run(context.Background())
	assert.ErrorContains(t, err, "connection refused")
	assert.Equal(t, Result{}, res)
}

func TestRun_SlaTrackingFailsOnSecondBatch(t *testing.T) {
	full := make([]string, BatchSize)
	for i := range full {
		full[i] = fmt.Sprintf("r%d", i)
	}
	timers := &fakeTimers{batches: [][]Lapsed{lapsed(full...)}}
	status := &fakeStatus{}
	c := New(timers, status)

	// First call succeeds, then sla-tracking fails.
	c.timers = &failAfterFirst{inner: timers}
	res, err := c.Run(context.Background())
	assert.Error(t, err)
	assert.Equal(t, BatchSize, res.Escalated, "work done before the failure is still reported")
}

type failAfterFirst struct {
	inner *fakeTimers
	done  bool
}

func (f *failAfterFirst) LapseDue(ctx context.Context, limit int) ([]Lapsed, error) {
	if f.done {
		return nil, errors.New("sla-tracking went down")
	}
	f.done = true
	return f.inner.LapseDue(ctx, limit)
}
