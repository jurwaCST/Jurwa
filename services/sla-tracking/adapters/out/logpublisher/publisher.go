// Package logpublisher is a temporary EventPublisher that writes each event as
// a JSON log line. It is replaced by a Redis Streams publisher
// (stream events:sla.lapsed) once the event bus is in place.
package logpublisher

import (
	"context"
	"log/slog"
	"time"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

// Publisher logs events instead of sending them.
type Publisher struct {
	logger *slog.Logger
}

// New returns a Publisher that writes to logger.
func New(logger *slog.Logger) *Publisher {
	return &Publisher{logger: logger}
}

func (p *Publisher) PublishSLALapsed(ctx context.Context, ev ports.SLALapsed) error {
	p.logger.InfoContext(ctx, "event published",
		"event", "sla.lapsed",
		"report_id", ev.ReportID,
		"category", ev.Category.String(),
		"deadline", ev.Deadline.Format(time.RFC3339),
		"lapsed_at", ev.LapsedAt.Format(time.RFC3339),
	)
	return nil
}
