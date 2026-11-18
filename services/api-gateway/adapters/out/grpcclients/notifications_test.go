package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
)

type fakeNotificationClient struct {
	notificationv1.NotificationServiceClient
	got *notificationv1.ListNotificationsRequest
}

func (f *fakeNotificationClient) ListNotifications(_ context.Context, in *notificationv1.ListNotificationsRequest, _ ...grpc.CallOption) (*notificationv1.ListNotificationsResponse, error) {
	f.got = in
	return &notificationv1.ListNotificationsResponse{NextPageToken: "next"}, nil
}

func TestNotificationsPassTheRequest(t *testing.T) {
	fake := &fakeNotificationClient{}
	in := &notificationv1.ListNotificationsRequest{RecipientUserId: "u1", UnreadOnly: true, PageSize: 5}
	resp, err := Notifications{C: fake}.ListNotifications(context.Background(), in)
	require.NoError(t, err)
	assert.Same(t, in, fake.got)
	assert.Equal(t, "next", resp.GetNextPageToken())
}
