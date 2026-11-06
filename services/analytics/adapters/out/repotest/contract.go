// Package repotest holds the behaviour every ports.ProjectionStore must have.
// The in-memory adapter runs it in its unit tests and the PostgreSQL adapter
// in its integration test, so both build the same reports from events.
package repotest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/analytics/domain"
	"github.com/jurwaCST/Jurwa/services/analytics/ports"
)

// t0 is in UTC and whole microseconds, the precision PostgreSQL keeps.
var t0 = time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)

func submitted(id, report string, c domain.Category, at time.Time) domain.Event {
	return domain.Event{ID: id, Type: domain.ReportSubmitted, ReportID: report, Category: c, OccurredAt: at}
}

func changed(id, report string, to domain.Status, at time.Time) domain.Event {
	return domain.Event{ID: id, Type: domain.StatusChanged, ReportID: report, ToStatus: to, OccurredAt: at}
}

// Run checks store behaviour. newStore must return an empty store.
func Run(t *testing.T, newStore func(t *testing.T) ports.ProjectionStore) {
	ctx := context.Background()

	apply := func(t *testing.T, s ports.ProjectionStore, e domain.Event) bool {
		t.Helper()
		ok, err := s.Apply(ctx, e)
		require.NoError(t, err)
		return ok
	}

	t.Run("empty store has no reports", func(t *testing.T) {
		reports, err := newStore(t).Reports(ctx)
		require.NoError(t, err)
		assert.Empty(t, reports)
	})

	t.Run("reports are built from events and sorted by id", func(t *testing.T) {
		s := newStore(t)
		assert.True(t, apply(t, s, submitted("e1", "hostel-1", domain.Hostel, t0)))
		assert.True(t, apply(t, s, submitted("e2", "academic-1", domain.AcademicBlock, t0.Add(time.Hour))))
		assert.True(t, apply(t, s, changed("e3", "hostel-1", domain.Resolved, t0.Add(6*time.Hour))))
		assert.True(t, apply(t, s, changed("e4", "academic-1", domain.Escalated, t0.Add(2*time.Hour))))
		assert.True(t, apply(t, s, changed("e5", "academic-1", domain.InProgress, t0.Add(3*time.Hour))))

		reports, err := s.Reports(ctx)
		require.NoError(t, err)
		require.Len(t, reports, 2)
		assertReport(t, domain.Report{ID: "academic-1", Category: domain.AcademicBlock, Status: domain.InProgress, CreatedAt: t0.Add(time.Hour), WasEscalated: true}, reports[0])
		assertReport(t, domain.Report{ID: "hostel-1", Category: domain.Hostel, Status: domain.Resolved, CreatedAt: t0, ResolvedAt: t0.Add(6 * time.Hour)}, reports[1])
	})

	t.Run("reopening a resolved report clears its resolved time", func(t *testing.T) {
		s := newStore(t)
		apply(t, s, submitted("e1", "r1", domain.Others, t0))
		apply(t, s, changed("e2", "r1", domain.Resolved, t0.Add(time.Hour)))
		apply(t, s, changed("e3", "r1", domain.Assigned, t0.Add(2*time.Hour)))

		reports, err := s.Reports(ctx)
		require.NoError(t, err)
		require.Len(t, reports, 1)
		assert.Equal(t, domain.Assigned, reports[0].Status)
		assert.True(t, reports[0].ResolvedAt.IsZero())
	})

	t.Run("a repeated event is applied once", func(t *testing.T) {
		s := newStore(t)
		assert.True(t, apply(t, s, submitted("e1", "r1", domain.Hostel, t0)))
		assert.False(t, apply(t, s, submitted("e1", "r1", domain.Hostel, t0)))
		assert.True(t, apply(t, s, changed("e2", "r1", domain.Escalated, t0)))
		assert.False(t, apply(t, s, changed("e2", "r1", domain.Escalated, t0)))
	})

	t.Run("unusable events change nothing", func(t *testing.T) {
		s := newStore(t)
		apply(t, s, submitted("e1", "r1", domain.Hostel, t0))
		for name, e := range map[string]domain.Event{
			"no event id":                  submitted("", "r2", domain.Hostel, t0),
			"no report id":                 submitted("x1", "", domain.Hostel, t0),
			"unknown category":             submitted("x2", "r3", domain.Category(9), t0),
			"submitted twice, new event":   submitted("x3", "r1", domain.Others, t0),
			"unknown status":               changed("x4", "r1", domain.Status(9), t0),
			"status for an unknown report": changed("x5", "nobody", domain.Resolved, t0),
		} {
			assert.False(t, apply(t, s, e), name)
		}
		reports, err := s.Reports(ctx)
		require.NoError(t, err)
		require.Len(t, reports, 1)
		assertReport(t, domain.Report{ID: "r1", Category: domain.Hostel, Status: domain.Open, CreatedAt: t0}, reports[0])
	})

	t.Run("an event that was not usable can be used later", func(t *testing.T) {
		s := newStore(t)
		early := changed("e2", "r1", domain.Resolved, t0.Add(time.Hour))
		assert.False(t, apply(t, s, early), "report not known yet")
		apply(t, s, submitted("e1", "r1", domain.Hostel, t0))
		assert.True(t, apply(t, s, early), "the same event id is not marked as seen")
	})
}

func assertReport(t *testing.T, want, got domain.Report) {
	t.Helper()
	assert.True(t, want.CreatedAt.Equal(got.CreatedAt), "created_at: want %v, got %v", want.CreatedAt, got.CreatedAt)
	assert.True(t, want.ResolvedAt.Equal(got.ResolvedAt), "resolved_at: want %v, got %v", want.ResolvedAt, got.ResolvedAt)
	got.CreatedAt, got.ResolvedAt = want.CreatedAt, want.ResolvedAt
	assert.Equal(t, want, got)
}
