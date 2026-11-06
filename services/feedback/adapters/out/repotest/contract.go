// Package repotest holds the behaviour every ports.FeedbackRepository must
// have. The in-memory adapter runs it in its unit tests and the PostgreSQL
// adapter in its integration test, so both store feedback the same way.
package repotest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/feedback/domain"
	"github.com/jurwaCST/Jurwa/services/feedback/ports"
)

const (
	report  = "33333333-3333-4333-8333-333333333333"
	student = "22222222-2222-4222-8222-222222222222"
)

// created is in UTC and whole microseconds, the precision PostgreSQL keeps.
var created = time.Date(2026, 11, 6, 9, 30, 0, 123000, time.UTC)

func feedback(id string, rating int32, comment string) domain.Feedback {
	return domain.Feedback{ID: id, ReportID: report, StudentUserID: student, Rating: rating, Comment: comment, CreatedAt: created}
}

// Run checks repo behaviour. newRepo must return an empty repository.
func Run(t *testing.T, newRepo func(t *testing.T) ports.FeedbackRepository) {
	ctx := context.Background()

	t.Run("saved feedback is found by report with every field", func(t *testing.T) {
		r := newRepo(t)
		want := feedback("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 4, "Fixed in two days, thank you")
		require.NoError(t, r.Create(ctx, want))

		got, err := r.GetByReport(ctx, report)
		require.NoError(t, err)
		assert.True(t, want.CreatedAt.Equal(got.CreatedAt))
		got.CreatedAt = want.CreatedAt
		assert.Equal(t, want, got)
	})

	t.Run("an empty comment stays empty", func(t *testing.T) {
		r := newRepo(t)
		require.NoError(t, r.Create(ctx, feedback("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", 5, "")))
		got, err := r.GetByReport(ctx, report)
		require.NoError(t, err)
		assert.Equal(t, "", got.Comment)
	})

	t.Run("a second feedback for the same report is refused and the first is kept", func(t *testing.T) {
		r := newRepo(t)
		require.NoError(t, r.Create(ctx, feedback("cccccccc-cccc-4ccc-8ccc-cccccccccccc", 5, "")))

		err := r.Create(ctx, feedback("dddddddd-dddd-4ddd-8ddd-dddddddddddd", 1, "changed my mind"))
		assert.ErrorIs(t, err, domain.ErrAlreadyExists)

		got, err := r.GetByReport(ctx, report)
		require.NoError(t, err)
		assert.Equal(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc", got.ID)
		assert.Equal(t, int32(5), got.Rating)
	})

	t.Run("a report with no feedback is not found", func(t *testing.T) {
		_, err := newRepo(t).GetByReport(ctx, report)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("a report id that is not a UUID is not found", func(t *testing.T) {
		_, err := newRepo(t).GetByReport(ctx, "not-a-uuid")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
