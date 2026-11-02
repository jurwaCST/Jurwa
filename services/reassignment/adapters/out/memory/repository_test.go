package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func a(id, report, assignee string, at time.Time) domain.Assignment {
	return domain.Assignment{ID: id, ReportID: report, AssigneeUserID: assignee, Current: true, AssignedAt: at}
}

func TestReplaceAndList(t *testing.T) {
	r := New()
	ctx := context.Background()

	_, err := r.Current(ctx, "r1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	list, err := r.List(ctx, "r1")
	require.NoError(t, err)
	assert.Empty(t, list)

	require.NoError(t, r.Replace(ctx, "", a("a1", "r1", "u1", t0)))
	assert.ErrorIs(t, r.Replace(ctx, "", a("a9", "r1", "u9", t0)), domain.ErrAlreadyAssigned)

	require.NoError(t, r.Replace(ctx, "a1", a("a2", "r1", "u2", t0.Add(time.Hour))))
	assert.ErrorIs(t, r.Replace(ctx, "a1", a("a3", "r1", "u3", t0)), domain.ErrConflict, "a1 is no longer current")
	assert.ErrorIs(t, r.Replace(ctx, "a1", a("a3", "r2", "u3", t0)), domain.ErrConflict, "r2 has no assignment")

	cur, err := r.Current(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, "a2", cur.ID)

	list, _ = r.List(ctx, "r1")
	require.Len(t, list, 2)
	assert.Equal(t, "a1", list[0].ID, "oldest first")
	assert.False(t, list[0].Current)
	assert.True(t, list[1].Current)

	// List returns a copy, so callers cannot change the store.
	list[0].AssigneeUserID = "changed"
	again, _ := r.List(ctx, "r1")
	assert.Equal(t, "u1", again[0].AssigneeUserID)
}
