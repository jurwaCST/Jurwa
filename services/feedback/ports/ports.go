package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

// FeedbackRepository stores at most one feedback per report.
// adapters/out/memory, or adapters/out/postgres when DATABASE_URL is set.
type FeedbackRepository interface {
	// Create returns domain.ErrAlreadyExists when the report already has feedback.
	Create(ctx context.Context, f domain.Feedback) error
	// GetByReport returns domain.ErrNotFound when the report has no feedback.
	GetByReport(ctx context.Context, reportID string) (domain.Feedback, error)
}

// StatusChecker says whether a report is RESOLVED. The real one calls
// status-lifecycle GetStatus. It returns domain.ErrUnavailable when
// status-lifecycle cannot be reached.
type StatusChecker interface {
	IsResolved(ctx context.Context, reportID string) (bool, error)
}
