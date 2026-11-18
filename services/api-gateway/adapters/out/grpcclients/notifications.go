package grpcclients

import (
	"context"

	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
)

// Notifications implements ports.Notifications.
type Notifications struct {
	C notificationv1.NotificationServiceClient
}

func (n Notifications) ListNotifications(ctx context.Context, in *notificationv1.ListNotificationsRequest) (*notificationv1.ListNotificationsResponse, error) {
	return n.C.ListNotifications(ctx, in)
}
