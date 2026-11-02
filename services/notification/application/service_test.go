package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

const (
	alice    = "22222222-2222-2222-2222-222222222222"
	reportID = "33333333-3333-3333-3333-333333333333"
)

// fakeSender counts calls and can be told to fail.
type fakeSender struct {
	calls int
	err   error
}

func (f *fakeSender) Send(context.Context, domain.Notification) error {
	f.calls++
	return f.err
}

// failingRepo is a store whose writes and reads always fail.
type failingRepo struct{}

func (failingRepo) Save(context.Context, domain.Notification) error { return errors.New("db down") }
func (failingRepo) ListByRecipient(context.Context, string, bool, int, int) ([]domain.Notification, bool, error) {
	return nil, false, errors.New("db down")
}

func newTestService(sender *fakeSender) (*Service, *time.Time) {
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	svc := NewService(memory.New(), sender)
	n := 0
	svc.newID = func() string { n++; return fmt.Sprintf("00000000-0000-4000-8000-%012d", n) }
	svc.now = func() time.Time { return now }
	return svc, &now
}

func validInput() domain.Input {
	return domain.Input{
		RecipientUserID: alice,
		Channel:         domain.InApp,
		EventType:       domain.EventStatusChanged,
		ReportID:        reportID,
		Subject:         "Report updated",
		Body:            "Your report is now In Progress.",
	}
}

func TestSendCallsSenderOnceAndStoresAsSent(t *testing.T) {
	sender := &fakeSender{}
	svc, now := newTestService(sender)
	ctx := context.Background()

	n, err := svc.Send(ctx, validInput())
	require.NoError(t, err)
	assert.Equal(t, 1, sender.calls, "Send is called once per SendNotification")
	assert.Equal(t, domain.Sent, n.Status)
	require.NotNil(t, n.SentAt)
	assert.Equal(t, *now, *n.SentAt)

	list, _, err := svc.List(ctx, alice, false, 0, "")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, domain.Sent, list[0].Status)
}

func TestSendBadInputDoesNotCallSender(t *testing.T) {
	sender := &fakeSender{}
	svc, _ := newTestService(sender)
	in := validInput()
	in.Body = ""

	_, err := svc.Send(context.Background(), in)
	assert.ErrorIs(t, err, domain.ErrInvalidBody)
	assert.Equal(t, 0, sender.calls)
}

func TestSenderFailureIsUnavailableAndNotStoredAsSent(t *testing.T) {
	sender := &fakeSender{err: errors.New("smtp down")}
	svc, _ := newTestService(sender)
	ctx := context.Background()

	_, err := svc.Send(ctx, validInput())
	assert.ErrorIs(t, err, ErrUnavailable)
	assert.Equal(t, 1, sender.calls)

	list, _, err := svc.List(ctx, alice, false, 0, "")
	require.NoError(t, err)
	require.Len(t, list, 1, "the failed message stays in the outbox")
	assert.Equal(t, domain.Failed, list[0].Status, "it is stored as FAILED, not SENT")
	assert.Nil(t, list[0].SentAt)
}

func TestSendStoreError(t *testing.T) {
	sender := &fakeSender{}
	svc := NewService(failingRepo{}, sender)
	_, err := svc.Send(context.Background(), validInput())
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrUnavailable)

	sender.err = errors.New("smtp down")
	_, err = svc.Send(context.Background(), validInput())
	assert.ErrorIs(t, err, ErrUnavailable, "the sender error is reported even if saving FAILED also fails")
}

func TestListNewestFirstAndOnlyThatUser(t *testing.T) {
	svc, now := newTestService(&fakeSender{})
	ctx := context.Background()

	first := validInput()
	first.Subject = "first"
	_, err := svc.Send(ctx, first)
	require.NoError(t, err)

	*now = now.Add(time.Minute)
	second := validInput()
	second.Subject = "second"
	_, err = svc.Send(ctx, second)
	require.NoError(t, err)

	other := validInput()
	other.RecipientUserID = "44444444-4444-4444-4444-444444444444"
	_, err = svc.Send(ctx, other)
	require.NoError(t, err)

	list, next, err := svc.List(ctx, alice, false, 0, "")
	require.NoError(t, err)
	assert.Empty(t, next)
	require.Len(t, list, 2)
	assert.Equal(t, "second", list[0].Subject)
	assert.Equal(t, "first", list[1].Subject)
}

func TestListPaging(t *testing.T) {
	svc, now := newTestService(&fakeSender{})
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		*now = now.Add(time.Minute)
		_, err := svc.Send(ctx, validInput())
		require.NoError(t, err)
	}

	page1, next, err := svc.List(ctx, alice, false, 2, "")
	require.NoError(t, err)
	assert.Len(t, page1, 2)
	assert.Equal(t, "2", next)

	page2, next, err := svc.List(ctx, alice, false, 2, next)
	require.NoError(t, err)
	assert.Len(t, page2, 1)
	assert.Empty(t, next)
}

func TestListRules(t *testing.T) {
	svc, _ := newTestService(&fakeSender{})
	ctx := context.Background()
	tests := []struct {
		name     string
		user     string
		pageSize int32
		token    string
		wantErr  error
	}{
		{"user not a UUID", "gayley", 0, "", domain.ErrInvalidRecipient},
		{"page size 0 uses the default", alice, 0, "", nil},
		{"page size 1", alice, 1, "", nil},
		{"page size 100 (limit)", alice, 100, "", nil},
		{"page size 101", alice, 101, "", ErrInvalidPageSize},
		{"negative page size", alice, -1, "", ErrInvalidPageSize},
		{"page token not a number", alice, 0, "abc", ErrInvalidPageToken},
		{"negative page token", alice, 0, "-1", ErrInvalidPageToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.List(ctx, tt.user, false, tt.pageSize, tt.token)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestListStoreError(t *testing.T) {
	_, _, err := NewService(failingRepo{}, &fakeSender{}).List(context.Background(), alice, false, 0, "")
	assert.Error(t, err)
}

func TestNewServiceDefaults(t *testing.T) {
	svc := NewService(memory.New(), &fakeSender{})
	n, err := svc.Send(context.Background(), validInput())
	require.NoError(t, err)
	assert.NotEmpty(t, n.ID, "a real UUID is generated")
	assert.False(t, n.CreatedAt.IsZero())
}
