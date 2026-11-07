package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

const (
	alice = "22222222-2222-2222-2222-222222222222"
	bob   = "44444444-4444-4444-4444-444444444444"
)

var start = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func note(id, user string, minute int) domain.Notification {
	return domain.Notification{ID: id, RecipientUserID: user, Status: domain.Sent, CreatedAt: start.Add(time.Duration(minute) * time.Minute)}
}

func TestListHidesPendingAndFailed(t *testing.T) {
	r := New()
	ctx := context.Background()
	failed := note("failed", alice, 1)
	failed.Status = domain.Failed
	pending := note("pending", alice, 2)
	pending.Status = domain.Pending
	require.NoError(t, r.Save(ctx, failed))
	require.NoError(t, r.Save(ctx, pending))
	require.NoError(t, r.Save(ctx, note("sent", alice, 3)))

	list, _, err := r.ListByRecipient(ctx, alice, false, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"sent"}, ids(list), "the user only sees messages that were delivered")

	got, ok := r.Get("failed")
	require.True(t, ok, "the failed message is still in the outbox")
	assert.Equal(t, domain.Failed, got.Status)

	_, ok = r.Get("missing")
	assert.False(t, ok)
}

func ids(list []domain.Notification) []string {
	out := make([]string, len(list))
	for i, n := range list {
		out[i] = n.ID
	}
	return out
}

func TestListOnlyThatUserNewestFirst(t *testing.T) {
	r := New()
	ctx := context.Background()
	require.NoError(t, r.Save(ctx, note("a1", alice, 1)))
	require.NoError(t, r.Save(ctx, note("b1", bob, 2)))
	require.NoError(t, r.Save(ctx, note("a2", alice, 3)))
	require.NoError(t, r.Save(ctx, note("a3", alice, 3))) // same time as a2, saved later

	list, more, err := r.ListByRecipient(ctx, alice, false, 0, 10)
	require.NoError(t, err)
	assert.False(t, more)
	assert.Equal(t, []string{"a3", "a2", "a1"}, ids(list))

	list, _, err = r.ListByRecipient(ctx, "55555555-5555-5555-5555-555555555555", false, 0, 10)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestListPages(t *testing.T) {
	r := New()
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		require.NoError(t, r.Save(ctx, note(fmt.Sprintf("n%d", i), alice, i)))
	}

	page, more, err := r.ListByRecipient(ctx, alice, false, 0, 2)
	require.NoError(t, err)
	assert.True(t, more)
	assert.Equal(t, []string{"n5", "n4"}, ids(page))

	page, more, err = r.ListByRecipient(ctx, alice, false, 4, 2)
	require.NoError(t, err)
	assert.False(t, more)
	assert.Equal(t, []string{"n1"}, ids(page))

	page, more, err = r.ListByRecipient(ctx, alice, false, 10, 2)
	require.NoError(t, err)
	assert.False(t, more)
	assert.Empty(t, page)
}

func TestListUnreadOnly(t *testing.T) {
	r := New()
	ctx := context.Background()
	read := note("read", alice, 1)
	readAt := start
	read.ReadAt = &readAt
	require.NoError(t, r.Save(ctx, read))
	require.NoError(t, r.Save(ctx, note("unread", alice, 2)))

	list, _, err := r.ListByRecipient(ctx, alice, true, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"unread"}, ids(list))
}

func TestSaveReplacesSameID(t *testing.T) {
	r := New()
	ctx := context.Background()
	n := note("a1", alice, 1)
	n.Status = domain.Pending
	require.NoError(t, r.Save(ctx, n))
	require.NoError(t, r.Save(ctx, n.MarkSent(start)))

	list, _, err := r.ListByRecipient(ctx, alice, false, 0, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, domain.Sent, list[0].Status)
}

func TestContract(t *testing.T) {
	repotest.Run(t, func(*testing.T) ports.NotificationRepository { return New() })
}
