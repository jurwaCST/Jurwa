package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/feedback/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

const (
	reportID  = "33333333-3333-3333-3333-333333333333"
	studentID = "22222222-2222-2222-2222-222222222222"
)

// fakeStatus says whether a report is RESOLVED, or fails, and counts calls.
type fakeStatus struct {
	resolved bool
	err      error
	calls    int
}

func (f *fakeStatus) IsResolved(context.Context, string) (bool, error) {
	f.calls++
	return f.resolved, f.err
}

// failingRepo is a store that always fails.
type failingRepo struct{}

func (failingRepo) Create(context.Context, domain.Feedback) error { return errors.New("db down") }
func (failingRepo) GetByReport(context.Context, string) (domain.Feedback, error) {
	return domain.Feedback{}, errors.New("db down")
}

func newTestService(status *fakeStatus) *Service {
	svc := NewService(memory.New(), status)
	n := 0
	svc.newID = func() string { n++; return fmt.Sprintf("00000000-0000-4000-8000-%012d", n) }
	svc.now = func() time.Time { return time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC) }
	return svc
}

func TestSubmitWhenResolved(t *testing.T) {
	svc := newTestService(&fakeStatus{resolved: true})
	ctx := context.Background()

	f, err := svc.Submit(ctx, reportID, studentID, 5, "Fixed quickly")
	require.NoError(t, err)
	assert.Equal(t, int32(5), f.Rating)
	assert.Equal(t, "00000000-0000-4000-8000-000000000001", f.ID)

	got, err := svc.Get(ctx, reportID)
	require.NoError(t, err)
	assert.Equal(t, f, got)
}

func TestSubmitWhenNotResolvedIsFailedPrecondition(t *testing.T) {
	svc := newTestService(&fakeStatus{resolved: false})
	_, err := svc.Submit(context.Background(), reportID, studentID, 5, "")
	assert.ErrorIs(t, err, domain.ErrNotResolved)

	_, err = svc.Get(context.Background(), reportID)
	assert.ErrorIs(t, err, domain.ErrNotFound, "nothing was stored")
}

func TestSecondFeedbackIsAlreadyExists(t *testing.T) {
	status := &fakeStatus{resolved: true}
	svc := newTestService(status)
	ctx := context.Background()
	_, err := svc.Submit(ctx, reportID, studentID, 5, "")
	require.NoError(t, err)

	_, err = svc.Submit(ctx, reportID, studentID, 1, "changed my mind")
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	assert.Equal(t, 1, status.calls, "status-lifecycle is not asked again when feedback already exists")

	got, err := svc.Get(ctx, reportID)
	require.NoError(t, err)
	assert.Equal(t, int32(5), got.Rating, "the first rating is kept")
}

func TestBadInputDoesNotCallStatusLifecycle(t *testing.T) {
	status := &fakeStatus{resolved: true}
	svc := newTestService(status)
	_, err := svc.Submit(context.Background(), reportID, studentID, 6, "")
	assert.ErrorIs(t, err, domain.ErrInvalidRating)
	assert.Equal(t, 0, status.calls)
}

func TestStatusLifecycleDown(t *testing.T) {
	svc := newTestService(&fakeStatus{err: fmt.Errorf("%w: down", domain.ErrUnavailable)})
	_, err := svc.Submit(context.Background(), reportID, studentID, 5, "")
	assert.ErrorIs(t, err, domain.ErrUnavailable)
}

func TestGetUnknownReportIsNotFound(t *testing.T) {
	_, err := newTestService(&fakeStatus{}).Get(context.Background(), reportID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetBadReportID(t *testing.T) {
	_, err := newTestService(&fakeStatus{}).Get(context.Background(), "abc")
	assert.ErrorIs(t, err, domain.ErrInvalidReportID)
}

func TestStoreErrors(t *testing.T) {
	svc := NewService(failingRepo{}, &fakeStatus{resolved: true})
	ctx := context.Background()

	_, err := svc.Submit(ctx, reportID, studentID, 5, "")
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrNotFound)

	_, err = svc.Get(ctx, reportID)
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrNotFound)
}

// raceRepo looks empty on GetByReport, but Create finds that another call
// has just added feedback for the same report.
type raceRepo struct{}

func (raceRepo) GetByReport(context.Context, string) (domain.Feedback, error) {
	return domain.Feedback{}, domain.ErrNotFound
}
func (raceRepo) Create(context.Context, domain.Feedback) error { return domain.ErrAlreadyExists }

func TestTwoSubmitsAtTheSameTime(t *testing.T) {
	svc := NewService(raceRepo{}, &fakeStatus{resolved: true})
	_, err := svc.Submit(context.Background(), reportID, studentID, 5, "")
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestNewServiceDefaults(t *testing.T) {
	svc := NewService(memory.New(), &fakeStatus{resolved: true})
	f, err := svc.Submit(context.Background(), reportID, studentID, 3, "")
	require.NoError(t, err)
	assert.NotEmpty(t, f.ID)
	assert.False(t, f.CreatedAt.IsZero())
}
