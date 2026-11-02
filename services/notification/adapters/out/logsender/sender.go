// Package logsender is a Sender that writes each notification as a JSON log
// line instead of delivering it. In-app messages are read with
// ListNotifications; real email delivery is out of scope.
package logsender

import (
	"context"
	"log/slog"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

// Sender logs notifications.
type Sender struct {
	logger *slog.Logger
}

// New returns a Sender that writes to logger.
func New(logger *slog.Logger) *Sender {
	return &Sender{logger: logger}
}

// Send logs the notification. The body is left out of the log on purpose,
// because it can contain personal details.
func (s *Sender) Send(ctx context.Context, n domain.Notification) error {
	s.logger.InfoContext(ctx, "notification sent",
		"id", n.ID,
		"recipient_user_id", n.RecipientUserID,
		"channel", n.Channel.String(),
		"event_type", n.EventType,
		"report_id", n.ReportID,
		"subject", n.Subject,
	)
	return nil
}
