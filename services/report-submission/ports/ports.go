package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

// ReportRepository stores reports. The in-memory adapter is used in Sprint 1;
// a MongoDB adapter replaces it in Sprint 2.
type ReportRepository interface {
	Save(ctx context.Context, report domain.Report) error
	// FindByID returns domain.ErrReportNotFound when no report has the id.
	FindByID(ctx context.Context, id string) (domain.Report, error)
	// Delete removes a report (SCRUM-163). domain.ErrReportNotFound when unknown.
	Delete(ctx context.Context, id string) error
}

// EventPublisher tells other services about new reports
// (stream events:report.submitted).
type EventPublisher interface {
	ReportSubmitted(ctx context.Context, report domain.Report) error
	// ReportWithdrawn says a report was deleted by its student, and who held it (SCRUM-163).
	ReportWithdrawn(ctx context.Context, report domain.Report, assigneeID string) error
}

// IDGenerator creates new report ids (UUIDs).
type IDGenerator interface {
	NewID() string
}
