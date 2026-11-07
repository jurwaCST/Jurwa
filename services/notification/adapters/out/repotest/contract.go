// Package repotest holds the behaviour every ports.NotificationRepository must
// have. The in-memory adapter runs it in its unit tests and the PostgreSQL
// adapter in its integration test, so both store notifications the same way.
package repotest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

const (
	alice  = "22222222-2222-4222-8222-222222222222"
	bob    = "44444444-4444-4444-8444-444444444444"
	report = "33333333-3333-4333-8333-333333333333"
)

// start is in UTC and whole microseconds, the precision PostgreSQL keeps.
var start = time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)

// ID returns a fixed UUID for a short test number.
func ID(n int) string { return fmt.Sprintf("00000000-0000-4000-8000-%012d", n) }

func note(n int, user string, minute int) domain.Notification {
	return domain.Notification{
		ID: ID(n), RecipientUserID: user, Channel: domain.InApp, EventType: domain.EventStatusChanged,
		Subject: "Report updated", Body: "Your report changed.", Status: domain.Sent,
		CreatedAt: start.Add(time.Duration(minute) * time.Minute),
	}
}

func ids(list []domain.Notification) []string {
	out := make([]string, len(list))
	for i, n := range list {
		out[i] = n.ID
	}
	return out
}

// Run checks repo behaviour. newRepo must return an empty repository.
func Run(t *testing.T, newRepo func(t *testing.T) ports.NotificationRepository) {
	ctx := context.Background()

	t.Run("every field is kept", func(t *testing.T) {
		r := newRepo(t)
		sentAt, readAt := start.Add(time.Second), start.Add(time.Hour)
		full := domain.Notification{
			ID: ID(1), RecipientUserID: alice, Channel: domain.Email, EventType: domain.EventSLALapsed,
			ReportID: report, Subject: "Escalated", Body: "Your report was escalated.",
			Status: domain.Sent, CreatedAt: start, SentAt: &sentAt, ReadAt: &readAt,
		}
		digest := note(2, alice, 1) // no report id, not read
		digest.EventType = domain.EventDigest
		require.NoError(t, r.Save(ctx, full))
		require.NoError(t, r.Save(ctx, digest))

		list, more, err := r.ListByRecipient(ctx, alice, false, 0, 10)
		require.NoError(t, err)
		assert.False(t, more)
		require.Len(t, list, 2)
		assertSame(t, digest, list[0])
		assertSame(t, full, list[1])
	})

	t.Run("only sent messages of that user, newest first", func(t *testing.T) {
		r := newRepo(t)
		failed := note(1, alice, 1)
		failed.Status = domain.Failed
		pending := note(2, alice, 2)
		pending.Status = domain.Pending
		for _, n := range []domain.Notification{failed, pending, note(3, alice, 3), note(4, bob, 4), note(5, alice, 5), note(6, alice, 5)} {
			require.NoError(t, r.Save(ctx, n))
		}
		list, more, err := r.ListByRecipient(ctx, alice, false, 0, 10)
		require.NoError(t, err)
		assert.False(t, more)
		assert.Equal(t, []string{ID(6), ID(5), ID(3)}, ids(list), "same time: the one saved later comes first")

		list, _, err = r.ListByRecipient(ctx, "55555555-5555-4555-8555-555555555555", false, 0, 10)
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("pages", func(t *testing.T) {
		r := newRepo(t)
		for i := 1; i <= 5; i++ {
			require.NoError(t, r.Save(ctx, note(i, alice, i)))
		}
		page, more, err := r.ListByRecipient(ctx, alice, false, 0, 2)
		require.NoError(t, err)
		assert.True(t, more)
		assert.Equal(t, []string{ID(5), ID(4)}, ids(page))

		page, more, err = r.ListByRecipient(ctx, alice, false, 4, 2)
		require.NoError(t, err)
		assert.False(t, more)
		assert.Equal(t, []string{ID(1)}, ids(page))

		page, more, err = r.ListByRecipient(ctx, alice, false, 10, 2)
		require.NoError(t, err)
		assert.False(t, more)
		assert.Empty(t, page)
	})

	t.Run("unread only", func(t *testing.T) {
		r := newRepo(t)
		read := note(1, alice, 1)
		readAt := start
		read.ReadAt = &readAt
		require.NoError(t, r.Save(ctx, read))
		require.NoError(t, r.Save(ctx, note(2, alice, 2)))

		list, _, err := r.ListByRecipient(ctx, alice, true, 0, 10)
		require.NoError(t, err)
		assert.Equal(t, []string{ID(2)}, ids(list))
	})

	t.Run("saving the same id again replaces it and keeps its place", func(t *testing.T) {
		r := newRepo(t)
		first := note(1, alice, 1)
		first.Status = domain.Pending
		require.NoError(t, r.Save(ctx, first))
		require.NoError(t, r.Save(ctx, note(2, alice, 1)))
		require.NoError(t, r.Save(ctx, first.MarkSent(start)))

		list, _, err := r.ListByRecipient(ctx, alice, false, 0, 10)
		require.NoError(t, err)
		assert.Equal(t, []string{ID(2), ID(1)}, ids(list), "the replaced one keeps its save order")
		assert.Equal(t, domain.Sent, list[1].Status)
	})
}

func assertSame(t *testing.T, want, got domain.Notification) {
	t.Helper()
	assert.True(t, want.CreatedAt.Equal(got.CreatedAt), "created_at")
	got.CreatedAt = want.CreatedAt
	got.SentAt = sameTime(t, want.SentAt, got.SentAt)
	got.ReadAt = sameTime(t, want.ReadAt, got.ReadAt)
	assert.Equal(t, want, got)
}

// sameTime checks two optional times are equal and returns want, so the
// struct comparison after it does not fail on time zone details.
func sameTime(t *testing.T, want, got *time.Time) *time.Time {
	t.Helper()
	if want == nil {
		assert.Nil(t, got)
		return got
	}
	if assert.NotNil(t, got) {
		assert.True(t, want.Equal(*got), "want %v, got %v", *want, *got)
	}
	return want
}
