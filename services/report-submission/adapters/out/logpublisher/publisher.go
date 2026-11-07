// Package logpublisher is the EventPublisher used when REDIS_URL is not set: it
// writes each event as a JSON log line instead of sending it.
package logpublisher

import (
	"context"
	"log/slog"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

// Publisher logs events instead of sending them.
type Publisher struct {
	logger *slog.Logger
}

// New returns a Publisher that writes to logger.
func New(logger *slog.Logger) *Publisher {
	return &Publisher{logger: logger}
}

// ReportSubmitted logs the event.
func (p *Publisher) ReportSubmitted(ctx context.Context, r domain.Report) error {
	p.logger.InfoContext(ctx, "event published",
		"event", "report.submitted",
		"report_id", r.ID,
		"category", string(r.Category),
		"building", r.Location.Building,
	)
	return nil
}
