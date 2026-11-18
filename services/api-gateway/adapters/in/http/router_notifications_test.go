package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
)

// ---- notification routes (SCRUM-147) ----

// fakeNotifications records what the gateway sends to notification.
type fakeNotifications struct {
	got *notificationv1.ListNotificationsRequest
	err error
}

func (f *fakeNotifications) ListNotifications(_ context.Context, in *notificationv1.ListNotificationsRequest) (*notificationv1.ListNotificationsResponse, error) {
	f.got = in
	if f.err != nil {
		return nil, f.err
	}
	reportID := "r1"
	return &notificationv1.ListNotificationsResponse{Notifications: []*notificationv1.Notification{
		{Id: "n1", RecipientUserId: in.GetRecipientUserId(), EventType: "status.changed", ReportId: &reportID, Subject: "Report in progress", Body: "A plumber is on the way"},
	}}, nil
}

func TestNotificationsNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/me/notifications", "", "").Code)
	assert.Nil(t, e.notifications.got)
}

func TestNotificationsAreForTheSignedInUser(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/me/notifications?recipientUserId=u-other&unreadOnly=true&pageSize=5&pageToken=abc", "student-token", "")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := e.notifications.got
	require.NotNil(t, got)
	assert.Equal(t, "u-student", got.GetRecipientUserId(), "the recipient comes from the JWT, never the query")
	assert.True(t, got.GetUnreadOnly())
	assert.Equal(t, int32(5), got.GetPageSize())
	assert.Equal(t, "abc", got.GetPageToken())

	list := decode(t, rec)["notifications"].([]any)
	require.Len(t, list, 1)
	n := list[0].(map[string]any)
	assert.Equal(t, "status.changed", n["eventType"], "lowerCamelCase JSON")
	assert.Equal(t, "u-student", n["recipientUserId"])
	assert.Equal(t, "", decode(t, rec)["nextPageToken"], "empty fields are still sent")
}

func TestNotificationsDefaults(t *testing.T) {
	e := newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/me/notifications", "head-token", "").Code)
	assert.Equal(t, "u-head", e.notifications.got.GetRecipientUserId())
	assert.False(t, e.notifications.got.GetUnreadOnly())
	assert.Zero(t, e.notifications.got.GetPageSize(), "0 means the service default")
}

func TestNotificationsAnySignedInRole(t *testing.T) {
	for _, token := range []string{"student-token", "head-token", "staff-token", "admin-token"} {
		assert.Equal(t, http.StatusOK, newEnv().do("GET", "/me/notifications", token, "").Code, token)
	}
}

func TestNotificationsBadQuery(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/me/notifications?pageSize=ten", "student-token", "").Code)
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/me/notifications?unreadOnly=maybe", "student-token", "").Code)
	assert.Nil(t, e.notifications.got, "notification is not called")
}

func TestNotificationsErrors(t *testing.T) {
	e := newEnv()
	e.notifications.err = status.Error(codes.InvalidArgument, "page_size must be 1 to 100")
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/me/notifications?pageSize=500", "student-token", "").Code)
	e.notifications.err = status.Error(codes.Unavailable, "dial notification:9018")
	assert.Equal(t, http.StatusServiceUnavailable, e.do("GET", "/me/notifications", "student-token", "").Code)
}
