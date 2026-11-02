package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

// Errors from the use cases. The gRPC adapter maps them to status codes.
var (
	ErrUnavailable      = errors.New("notification could not be delivered")
	ErrInvalidPageSize  = errors.New("page_size must be 1 to 100")
	ErrInvalidPageToken = errors.New("invalid page_token")
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Service holds the notification use cases.
type Service struct {
	repo   ports.NotificationRepository
	sender ports.Sender
	newID  func() string
	now    func() time.Time
}

// NewService wires the use cases to the store and the sender.
func NewService(repo ports.NotificationRepository, sender ports.Sender) *Service {
	return &Service{
		repo:   repo,
		sender: sender,
		newID:  func() string { return uuid.NewString() },
		now:    func() time.Time { return time.Now().UTC() },
	}
}

// Send checks the input, delivers the notification and stores it.
// If delivery fails it is stored as FAILED and ErrUnavailable is returned.
func (s *Service) Send(ctx context.Context, in domain.Input) (domain.Notification, error) {
	n, err := domain.NewNotification(s.newID(), in, s.now())
	if err != nil {
		return domain.Notification{}, err
	}
	if sendErr := s.sender.Send(ctx, n); sendErr != nil {
		// Keep the failed message in the outbox. If saving also fails, the
		// caller still learns that delivery failed.
		_ = s.repo.Save(ctx, n.MarkFailed())
		return domain.Notification{}, fmt.Errorf("%w: %v", ErrUnavailable, sendErr)
	}
	n = n.MarkSent(s.now())
	if err := s.repo.Save(ctx, n); err != nil {
		return domain.Notification{}, fmt.Errorf("save notification: %w", err)
	}
	return n, nil
}

// List returns one user's notifications, newest first. pageToken is the
// number of items already returned ("" for the first page). The returned
// token is "" when there are no more pages.
func (s *Service) List(ctx context.Context, userID string, unreadOnly bool, pageSize int32, pageToken string) ([]domain.Notification, string, error) {
	if err := domain.ValidateUserID(userID); err != nil {
		return nil, "", err
	}
	size := int(pageSize)
	switch {
	case size == 0:
		size = defaultPageSize
	case size < 0 || size > maxPageSize:
		return nil, "", ErrInvalidPageSize
	}
	offset := 0
	if pageToken != "" {
		v, err := strconv.Atoi(pageToken)
		if err != nil || v < 0 {
			return nil, "", ErrInvalidPageToken
		}
		offset = v
	}
	list, more, err := s.repo.ListByRecipient(ctx, userID, unreadOnly, offset, size)
	if err != nil {
		return nil, "", fmt.Errorf("list notifications: %w", err)
	}
	next := ""
	if more {
		next = strconv.Itoa(offset + len(list))
	}
	return list, next, nil
}
