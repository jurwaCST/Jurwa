// Package logpublisher is a temporary EventPublisher that writes each event as
// a JSON log line. It is replaced by a Redis Streams publisher
// (stream events:report.reassigned) once the event bus is in place.
package logpublisher

import (
	"context"
	"log/slog"
	"time"

	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

// Publisher logs events instead of sending them.
type Publisher struct {
	logger *slog.Logger
}

// New returns a Publisher that writes to logger.
func New(logger *slog.Logger) *Publisher {
	return &Publisher{logger: logger}
}

func (p *Publisher) PublishReportReassigned(ctx context.Context, ev ports.ReportReassigned) error {
	p.logger.InfoContext(ctx, "event published",
		"event", "report.reassigned",
		"assignment_id", ev.AssignmentID,
		"report_id", ev.ReportID,
		"from_user_id", ev.FromUserID,
		"to_user_id", ev.ToUserID,
		"actor_id", ev.ActorID,
		"reason", ev.Reason,
		"assigned_at", ev.AssignedAt.Format(time.RFC3339),
	)
	return nil
}
