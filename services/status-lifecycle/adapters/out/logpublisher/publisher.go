// Package logpublisher is a temporary EventPublisher that writes each event as
// a JSON log line. It is replaced by a Redis Streams publisher in Sprint 2.
package logpublisher

import (
	"context"
	"log/slog"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

// Publisher logs events instead of sending them.
type Publisher struct {
	logger *slog.Logger
}

// New returns a Publisher that writes to logger.
func New(logger *slog.Logger) *Publisher {
	return &Publisher{logger: logger}
}

func (p *Publisher) PublishStatusChanged(ctx context.Context, ev ports.StatusChanged) error {
	p.logger.InfoContext(ctx, "event published",
		"event", "status.changed",
		"report_id", ev.ReportID,
		"from_status", ev.From.String(),
		"to_status", ev.To.String(),
		"actor_id", ev.ActorID,
		"version", ev.Version,
		"reason", ev.Reason,
	)
	return nil
}
