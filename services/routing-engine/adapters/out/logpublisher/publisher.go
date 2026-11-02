// Package logpublisher is a temporary EventPublisher that writes each event as
// a JSON log line. It is replaced by a Redis Streams publisher in Sprint 3.
package logpublisher

import (
	"context"
	"log/slog"

	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

// Publisher logs events instead of sending them.
type Publisher struct {
	logger *slog.Logger
}

// New returns a Publisher that writes to logger.
func New(logger *slog.Logger) *Publisher {
	return &Publisher{logger: logger}
}

func (p *Publisher) PublishReportRouted(ctx context.Context, ev ports.ReportRouted) error {
	p.logger.InfoContext(ctx, "event published",
		"event", "report.routed",
		"stream", "events:report.routed",
		"report_id", ev.ReportID,
		"category", string(ev.Category),
		"category_head_id", ev.CategoryHeadID,
		"target_position", ev.TargetPosition,
		"rule_id", ev.RuleID,
	)
	return nil
}
