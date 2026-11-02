package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/notification/application"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

const (
	alice    = "22222222-2222-2222-2222-222222222222"
	reportID = "33333333-3333-3333-3333-333333333333"
)

type fakeSender struct{ err error }

func (f fakeSender) Send(context.Context, domain.Notification) error { return f.err }

func newHandler(sender fakeSender) *Handler {
	return NewHandler(application.NewService(memory.New(), sender))
}

func sendReq() *notificationv1.SendNotificationRequest {
	rid := reportID
	return &notificationv1.SendNotificationRequest{
		RecipientUserId: alice,
		Channel:         notificationv1.Channel_IN_APP,
		EventType:       "status.changed",
		ReportId:        &rid,
		Subject:         "Report updated",
		Body:            "Your report is now In Progress.",
	}
}

func TestSendThenList(t *testing.T) {
	h := newHandler(fakeSender{})
	ctx := context.Background()

	sent, err := h.SendNotification(ctx, sendReq())
	require.NoError(t, err)
	n := sent.GetNotification()
	assert.NotEmpty(t, n.GetId())
	assert.Equal(t, alice, n.GetRecipientUserId())
	assert.Equal(t, notificationv1.Channel_IN_APP, n.GetChannel())
	assert.Equal(t, reportID, n.GetReportId())
	assert.Equal(t, notificationv1.DeliveryStatus_SENT, n.GetDeliveryStatus())
	assert.NotNil(t, n.GetCreatedAt())
	assert.NotNil(t, n.GetSentAt())
	assert.Nil(t, n.ReadAt)

	list, err := h.ListNotifications(ctx, &notificationv1.ListNotificationsRequest{RecipientUserId: alice})
	require.NoError(t, err)
	require.Len(t, list.GetNotifications(), 1)
	assert.Equal(t, n.GetId(), list.GetNotifications()[0].GetId())
	assert.Empty(t, list.GetNextPageToken())
}

func TestSendWithoutReportID(t *testing.T) {
	req := sendReq()
	req.ReportId = nil
	req.EventType = "digest.daily"
	res, err := newHandler(fakeSender{}).SendNotification(context.Background(), req)
	require.NoError(t, err)
	assert.Nil(t, res.GetNotification().ReportId)
}

func TestErrorCodes(t *testing.T) {
	ctx := context.Background()
	h := newHandler(fakeSender{})

	bad := sendReq()
	bad.Subject = ""
	_, err := h.SendNotification(ctx, bad)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "missing subject")

	bad = sendReq()
	bad.Channel = notificationv1.Channel_CHANNEL_UNSPECIFIED
	_, err = h.SendNotification(ctx, bad)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "channel unspecified")

	_, err = newHandler(fakeSender{err: errors.New("down")}).SendNotification(ctx, sendReq())
	assert.Equal(t, codes.Unavailable, status.Code(err), "sender failure")

	_, err = h.ListNotifications(ctx, &notificationv1.ListNotificationsRequest{RecipientUserId: "x"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "recipient not a UUID")

	_, err = h.ListNotifications(ctx, &notificationv1.ListNotificationsRequest{RecipientUserId: alice, PageSize: 101})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "page size too big")

	_, err = h.ListNotifications(ctx, &notificationv1.ListNotificationsRequest{RecipientUserId: alice, PageToken: "abc"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "bad page token")
}

func TestToStatusErrorInternal(t *testing.T) {
	err := toStatusError(errors.New("db down"))
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "db down")
}

func TestToProtoReadAt(t *testing.T) {
	readAt := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	out := toProto(domain.Notification{ReadAt: &readAt})
	require.NotNil(t, out.ReadAt)
	assert.Equal(t, readAt, out.GetReadAt().AsTime())
}
