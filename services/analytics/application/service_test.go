package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

// failingStore always fails.
type failingStore struct{}

func (failingStore) Apply(context.Context, domain.Event) (bool, error) {
	return false, errors.New("db down")
}
func (failingStore) Reports(context.Context) ([]domain.Report, error) {
	return nil, errors.New("db down")
}

func TestRecordThenSummaryAndBreakdown(t *testing.T) {
	svc := NewService(memory.New())
	ctx := context.Background()

	events := []domain.Event{
		{ID: "e1", Type: domain.ReportSubmitted, ReportID: "r1", Category: domain.Hostel, OccurredAt: t0},
		{ID: "e2", Type: domain.ReportSubmitted, ReportID: "r2", Category: domain.AcademicBlock, OccurredAt: t0},
		{ID: "e3", Type: domain.StatusChanged, ReportID: "r2", ToStatus: domain.Resolved, OccurredAt: t0.Add(3 * time.Hour)},
	}
	for _, e := range events {
		applied, err := svc.Record(ctx, e)
		require.NoError(t, err)
		assert.True(t, applied)
	}
	applied, err := svc.Record(ctx, events[0])
	require.NoError(t, err)
	assert.False(t, applied, "a repeated event id is ignored")

	s, err := svc.Summary(ctx, domain.Period{})
	require.NoError(t, err)
	assert.Equal(t, domain.Counts{Total: 2, Open: 1, Resolved: 1}, s.Counts)
	assert.Equal(t, 3.0, s.AverageResolutionHours)

	b, err := svc.Breakdown(ctx, domain.Period{})
	require.NoError(t, err)
	require.Len(t, b, 3)
	assert.Equal(t, int64(1), b[0].Counts.Total, "hostel")
	assert.Equal(t, int64(1), b[1].Counts.Resolved, "academic block")
	assert.Equal(t, int64(0), b[2].Counts.Total, "others")
}

func TestBadPeriodIsRejected(t *testing.T) {
	svc := NewService(memory.New())
	from := t0.Add(time.Hour)
	to := t0
	_, err := svc.Summary(context.Background(), domain.Period{From: &from, To: &to})
	assert.ErrorIs(t, err, domain.ErrInvalidPeriod)
	_, err = svc.Breakdown(context.Background(), domain.Period{From: &from, To: &to})
	assert.ErrorIs(t, err, domain.ErrInvalidPeriod)
}

func TestStoreErrors(t *testing.T) {
	svc := NewService(failingStore{})
	ctx := context.Background()
	_, err := svc.Record(ctx, domain.Event{ID: "e1"})
	assert.Error(t, err)
	_, err = svc.Summary(ctx, domain.Period{})
	assert.Error(t, err)
	_, err = svc.Breakdown(ctx, domain.Period{})
	assert.Error(t, err)
}
