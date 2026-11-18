package ports

import (
	"context"
	"errors"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

// NotificationRepository is the outbox of notifications.
// adapters/out/memory, or adapters/out/postgres when DATABASE_URL is set.
type NotificationRepository interface {
	Save(ctx context.Context, n domain.Notification) error
	// Create adds n only when its id is new and says whether it did. An id
	// already there is left as it is (in PostgreSQL: ON CONFLICT DO NOTHING),
	// so a message made from an event delivered twice is stored once.
	Create(ctx context.Context, n domain.Notification) (created bool, err error)
	// Exists says whether a notification with this id is stored.
	Exists(ctx context.Context, id string) (bool, error)
	// ListByRecipient returns one user's SENT notifications, newest first,
	// skipping offset and returning at most limit. more is true when there
	// are more after this page.
	ListByRecipient(ctx context.Context, userID string, unreadOnly bool, offset, limit int) (list []domain.Notification, more bool, err error)
}

// Sender delivers one notification. Sprint 2 only logs it (in-app);
// real email delivery is out of scope.
type Sender interface {
	Send(ctx context.Context, n domain.Notification) error
}

// Report is what the event consumer needs from a report: who sent it and where.
type Report struct {
	ID         string
	ReporterID string
	Building   string
	Wing       string
	Area       string
}

// Errors from Reports. Not found cannot be helped by trying again;
// unavailable can.
var (
	ErrReportNotFound     = errors.New("report not found")
	ErrReportsUnavailable = errors.New("report-submission unavailable")
)

// Reports finds a report in report-submission (adapters/out/reportclient).
// status.changed does not name the student, so the consumer asks here.
type Reports interface {
	GetReport(ctx context.Context, reportID string) (Report, error)
}
