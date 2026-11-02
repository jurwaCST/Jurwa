package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

// NotificationRepository is the outbox of notifications.
// In-memory for now; a database in Sprint 3.
type NotificationRepository interface {
	Save(ctx context.Context, n domain.Notification) error
	// ListByRecipient returns one user's notifications, newest first,
	// skipping offset and returning at most limit. more is true when there
	// are more after this page.
	ListByRecipient(ctx context.Context, userID string, unreadOnly bool, offset, limit int) (list []domain.Notification, more bool, err error)
}

// Sender delivers one notification. Sprint 2 only logs it (in-app);
// real email delivery is out of scope.
type Sender interface {
	Send(ctx context.Context, n domain.Notification) error
}
