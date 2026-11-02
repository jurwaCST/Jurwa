package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

const (
	r1 = "11111111-1111-1111-1111-111111111111"
	r2 = "22222222-2222-2222-2222-222222222222"
	r3 = "33333333-3333-3333-3333-333333333333"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

// fakePolicy returns the windows in the map, ErrNoPolicy for others, or err when set.
type fakePolicy struct {
	windows map[domain.Category]time.Duration
	err     error
}

func (f fakePolicy) Window(_ context.Context, c domain.Category) (time.Duration, error) {
	if f.err != nil {
		return 0, f.err
	}
	w, ok := f.windows[c]
	if !ok {
		return 0, ports.ErrNoPolicy
	}
	return w, nil
}

// fakePublisher records events, or fails when err is set.
type fakePublisher struct {
	events []ports.SLALapsed
	err    error
}

func (f *fakePublisher) PublishSLALapsed(_ context.Context, ev ports.SLALapsed) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, ev)
	return nil
}

// clock is a settable time source.
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

type fixture struct {
	svc   *Service
	repo  *memory.Repository
	pub   *fakePublisher
	clock *clock
}

func newFixture() fixture {
	repo := memory.New()
	pub := &fakePublisher{}
	c := &clock{t: t0}
	policy := fakePolicy{windows: map[domain.Category]time.Duration{
		domain.Hostel:        48 * time.Hour,
		domain.AcademicBlock: 24 * time.Hour,
	}}
	return fixture{svc: NewService(repo, policy, pub, c.now), repo: repo, pub: pub, clock: c}
}

func TestStartTimer(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	tm, err := f.svc.StartTimer(ctx, r1, domain.Hostel)
	require.NoError(t, err)
	assert.Equal(t, domain.Running, tm.State)
	assert.Equal(t, t0, tm.StartedAt)
	assert.Equal(t, t0.Add(48*time.Hour), tm.Deadline, "deadline uses the sla-policy window")

	got, err := f.svc.GetTimer(ctx, r1)
	require.NoError(t, err)
	assert.Equal(t, tm, got)

	_, err = f.svc.StartTimer(ctx, r1, domain.Hostel)
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestStartTimer_Errors(t *testing.T) {
	tests := []struct {
		name     string
		reportID string
		category domain.Category
		policy   ports.PolicyClient
		wantErr  error
	}{
		{"report id missing", "", domain.Hostel, nil, ErrInvalidArgument},
		{"report id not a UUID", "abc", domain.Hostel, nil, ErrInvalidArgument},
		{"category unspecified", r1, domain.CategoryUnspecified, nil, domain.ErrInvalidCategory},
		{"no window for category", r1, domain.Others, nil, ErrNoPolicy},
		{"sla-policy down", r1, domain.Hostel, fakePolicy{err: errors.New("connection refused")}, ErrPolicyUnavailable},
		{"sla-policy returns zero window", r1, domain.Hostel, fakePolicy{windows: map[domain.Category]time.Duration{domain.Hostel: 0}}, domain.ErrInvalidWindow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			if tt.policy != nil {
				f.svc = NewService(f.repo, tt.policy, f.pub, f.clock.now)
			}
			_, err := f.svc.StartTimer(context.Background(), tt.reportID, tt.category)
			assert.ErrorIs(t, err, tt.wantErr)

			_, err = f.svc.GetTimer(context.Background(), r1)
			assert.ErrorIs(t, err, domain.ErrNotFound, "nothing is saved on error")
		})
	}
}

func TestStopTimer(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	_, err := f.svc.StopTimer(ctx, r1)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.svc.StopTimer(ctx, "bad")
	assert.ErrorIs(t, err, ErrInvalidArgument)

	_, err = f.svc.StartTimer(ctx, r1, domain.Hostel)
	require.NoError(t, err)

	tm, err := f.svc.StopTimer(ctx, r1)
	require.NoError(t, err)
	assert.Equal(t, domain.Stopped, tm.State)

	// A stopped timer is never due, even long after its deadline.
	f.clock.t = t0.Add(100 * time.Hour)
	due, err := f.svc.ListDueTimers(ctx, time.Time{}, 0)
	require.NoError(t, err)
	assert.Empty(t, due)

	// Stopping twice is safe.
	_, err = f.svc.StopTimer(ctx, r1)
	assert.NoError(t, err)
}

func TestGetTimer_Errors(t *testing.T) {
	f := newFixture()
	_, err := f.svc.GetTimer(context.Background(), "")
	assert.ErrorIs(t, err, ErrInvalidArgument)
	_, err = f.svc.GetTimer(context.Background(), r1)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestListDueTimers_LapsesAndPublishes(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_, _ = f.svc.StartTimer(ctx, r1, domain.Hostel)        // deadline t0+48h
	_, _ = f.svc.StartTimer(ctx, r2, domain.AcademicBlock) // deadline t0+24h
	_, _ = f.svc.StartTimer(ctx, r3, domain.Hostel)
	_, _ = f.svc.StopTimer(ctx, r3) // resolved in time

	// Before any deadline: nothing is due.
	due, err := f.svc.ListDueTimers(ctx, time.Time{}, 0)
	require.NoError(t, err)
	assert.Empty(t, due)
	assert.Empty(t, f.pub.events)

	// 30 hours later only r2 (24 hour window) has lapsed.
	f.clock.t = t0.Add(30 * time.Hour)
	due, err = f.svc.ListDueTimers(ctx, time.Time{}, 0)
	require.NoError(t, err)
	require.Len(t, due, 1)
	assert.Equal(t, r2, due[0].ReportID)
	assert.Equal(t, domain.Lapsed, due[0].State)

	require.Len(t, f.pub.events, 1)
	assert.Equal(t, ports.SLALapsed{
		ReportID: r2, Category: domain.AcademicBlock,
		Deadline: t0.Add(24 * time.Hour), LapsedAt: t0.Add(30 * time.Hour),
	}, f.pub.events[0])

	got, _ := f.svc.GetTimer(ctx, r2)
	assert.Equal(t, domain.Lapsed, got.State, "saved as LAPSED")

	// The next run does not escalate r2 again.
	due, _ = f.svc.ListDueTimers(ctx, time.Time{}, 0)
	assert.Empty(t, due)
	assert.Len(t, f.pub.events, 1)
}

func TestListDueTimers_NowAndLimit(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_, _ = f.svc.StartTimer(ctx, r1, domain.AcademicBlock)
	_, _ = f.svc.StartTimer(ctx, r2, domain.AcademicBlock)
	f.clock.t = t0.Add(100 * time.Hour)

	// An explicit now earlier than the deadlines finds nothing.
	due, err := f.svc.ListDueTimers(ctx, t0.Add(time.Hour), 0)
	require.NoError(t, err)
	assert.Empty(t, due)

	due, err = f.svc.ListDueTimers(ctx, t0.Add(24*time.Hour), 1)
	require.NoError(t, err)
	assert.Len(t, due, 1, "limit 1")
}

func TestListDueTimers_Errors(t *testing.T) {
	tests := []struct {
		name  string
		now   time.Time
		limit int32
	}{
		{"now in the future", t0.Add(time.Minute), 0},
		{"negative limit", time.Time{}, -1},
		{"limit above 500", time.Time{}, 501},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			_, err := f.svc.ListDueTimers(context.Background(), tt.now, tt.limit)
			assert.ErrorIs(t, err, ErrInvalidArgument)
		})
	}

	// Boundaries that are allowed.
	f := newFixture()
	for _, limit := range []int32{1, 500} {
		_, err := f.svc.ListDueTimers(context.Background(), t0, limit)
		assert.NoError(t, err, limit)
	}
}

func TestListDueTimers_PublishFails(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_, _ = f.svc.StartTimer(ctx, r2, domain.AcademicBlock)
	f.clock.t = t0.Add(30 * time.Hour)
	f.pub.err = errors.New("redis down")

	_, err := f.svc.ListDueTimers(ctx, time.Time{}, 0)
	assert.ErrorContains(t, err, "redis down")

	// The timer stays RUNNING so the next run retries it.
	got, _ := f.svc.GetTimer(ctx, r2)
	assert.Equal(t, domain.Running, got.State)

	f.pub.err = nil
	due, err := f.svc.ListDueTimers(ctx, time.Time{}, 0)
	require.NoError(t, err)
	assert.Len(t, due, 1)
}

// failingRepo fails every call.
type failingRepo struct{ ports.TimerRepository }

func (failingRepo) Get(context.Context, string) (domain.Timer, error) {
	return domain.Timer{}, errors.New("store down")
}
func (failingRepo) ListDue(context.Context, time.Time, int) ([]domain.Timer, error) {
	return nil, errors.New("store down")
}

// updateFailsRepo works except for Update.
type updateFailsRepo struct{ *memory.Repository }

func (updateFailsRepo) Update(context.Context, domain.Timer) error { return errors.New("store down") }

func TestStoreErrors(t *testing.T) {
	ctx := context.Background()
	policy := fakePolicy{windows: map[domain.Category]time.Duration{domain.Hostel: time.Hour}}

	svc := NewService(failingRepo{}, policy, &fakePublisher{}, func() time.Time { return t0 })
	_, err := svc.StopTimer(ctx, r1)
	assert.ErrorContains(t, err, "store down")
	_, err = svc.ListDueTimers(ctx, time.Time{}, 0)
	assert.ErrorContains(t, err, "store down")

	c := &clock{t: t0}
	repo := updateFailsRepo{memory.New()}
	svc = NewService(repo, policy, &fakePublisher{}, c.now)
	_, err = svc.StartTimer(ctx, r1, domain.Hostel)
	require.NoError(t, err)
	_, err = svc.StopTimer(ctx, r1)
	assert.ErrorContains(t, err, "store down")

	c.t = t0.Add(2 * time.Hour)
	_, err = svc.ListDueTimers(ctx, time.Time{}, 0)
	assert.ErrorContains(t, err, "store down")
}
