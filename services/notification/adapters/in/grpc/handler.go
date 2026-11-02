package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
	"github.com/jurwaCST/Jurwa/services/notification/application"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

// Handler implements NotificationService from proto/notification/v1.
type Handler struct {
	notificationv1.UnimplementedNotificationServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) SendNotification(ctx context.Context, req *notificationv1.SendNotificationRequest) (*notificationv1.SendNotificationResponse, error) {
	// domain.Channel uses the same numbers as notificationv1.Channel.
	n, err := h.app.Send(ctx, domain.Input{
		RecipientUserID: req.GetRecipientUserId(),
		Channel:         domain.Channel(req.GetChannel()),
		EventType:       req.GetEventType(),
		ReportID:        req.GetReportId(),
		Subject:         req.GetSubject(),
		Body:            req.GetBody(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &notificationv1.SendNotificationResponse{Notification: toProto(n)}, nil
}

func (h *Handler) ListNotifications(ctx context.Context, req *notificationv1.ListNotificationsRequest) (*notificationv1.ListNotificationsResponse, error) {
	list, next, err := h.app.List(ctx, req.GetRecipientUserId(), req.GetUnreadOnly(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatusError(err)
	}
	out := make([]*notificationv1.Notification, len(list))
	for i, n := range list {
		out[i] = toProto(n)
	}
	return &notificationv1.ListNotificationsResponse{Notifications: out, NextPageToken: next}, nil
}

func toProto(n domain.Notification) *notificationv1.Notification {
	out := &notificationv1.Notification{
		Id:              n.ID,
		RecipientUserId: n.RecipientUserID,
		Channel:         notificationv1.Channel(n.Channel),
		EventType:       n.EventType,
		Subject:         n.Subject,
		Body:            n.Body,
		DeliveryStatus:  notificationv1.DeliveryStatus(n.Status),
		CreatedAt:       timestamppb.New(n.CreatedAt),
	}
	if n.ReportID != "" {
		rid := n.ReportID
		out.ReportId = &rid
	}
	if n.SentAt != nil {
		out.SentAt = timestamppb.New(*n.SentAt)
	}
	if n.ReadAt != nil {
		out.ReadAt = timestamppb.New(*n.ReadAt)
	}
	return out
}

// toStatusError maps errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidRecipient),
		errors.Is(err, domain.ErrInvalidChannel),
		errors.Is(err, domain.ErrInvalidEventType),
		errors.Is(err, domain.ErrInvalidReportID),
		errors.Is(err, domain.ErrInvalidSubject),
		errors.Is(err, domain.ErrInvalidBody),
		errors.Is(err, application.ErrInvalidPageSize),
		errors.Is(err, application.ErrInvalidPageToken):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, application.ErrUnavailable):
		return status.Error(codes.Unavailable, application.ErrUnavailable.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
