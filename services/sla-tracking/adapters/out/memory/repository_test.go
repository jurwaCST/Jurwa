package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func timer(id string, deadline time.Time, state domain.State) domain.Timer {
	return domain.Timer{ReportID: id, Category: domain.Hostel, StartedAt: now, Deadline: deadline, State: state}
}

func TestCreateGetUpdate(t *testing.T) {
	r := New()
	ctx := context.Background()

	_, err := r.Get(ctx, "r1")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	require.NoError(t, r.Create(ctx, timer("r1", now, domain.Running)))
	assert.ErrorIs(t, r.Create(ctx, timer("r1", now, domain.Running)), domain.ErrAlreadyExists)

	got, err := r.Get(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, domain.Running, got.State)

	got.State = domain.Stopped
	require.NoError(t, r.Update(ctx, got))
	got, _ = r.Get(ctx, "r1")
	assert.Equal(t, domain.Stopped, got.State)

	assert.ErrorIs(t, r.Update(ctx, timer("missing", now, domain.Running)), domain.ErrNotFound)
}

func TestListDue(t *testing.T) {
	r := New()
	ctx := context.Background()
	require.NoError(t, r.Create(ctx, timer("late-2", now.Add(-time.Hour), domain.Running)))
	require.NoError(t, r.Create(ctx, timer("late-1", now.Add(-2*time.Hour), domain.Running)))
	require.NoError(t, r.Create(ctx, timer("exact", now, domain.Running)))
	require.NoError(t, r.Create(ctx, timer("future", now.Add(time.Second), domain.Running)))
	require.NoError(t, r.Create(ctx, timer("stopped", now.Add(-3*time.Hour), domain.Stopped)))
	require.NoError(t, r.Create(ctx, timer("lapsed", now.Add(-3*time.Hour), domain.Lapsed)))

	due, err := r.ListDue(ctx, now, 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"late-1", "late-2", "exact"}, ids(due), "running and due, earliest deadline first")

	due, _ = r.ListDue(ctx, now, 2)
	assert.Equal(t, []string{"late-1", "late-2"}, ids(due), "limit is applied")
}

func TestListDue_SameDeadlineOrderedByID(t *testing.T) {
	r := New()
	ctx := context.Background()
	for i := 3; i >= 1; i-- {
		require.NoError(t, r.Create(ctx, timer(fmt.Sprintf("r%d", i), now, domain.Running)))
	}
	due, _ := r.ListDue(ctx, now, 10)
	assert.Equal(t, []string{"r1", "r2", "r3"}, ids(due))
}

func ids(ts []domain.Timer) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		out = append(out, t.ReportID)
	}
	return out
}
