package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

// HistoryRepository stores report summaries and timeline entries. The in-memory
// adapter is used in Sprint 2; a PostgreSQL adapter (docs/design/schema/report-history.sql)
// replaces it in Sprint 3.
type HistoryRepository interface {
	// CreateSummary returns domain.ErrReportExists when the report is already stored.
	CreateSummary(ctx context.Context, s domain.Summary) error
	// FindSummary returns domain.ErrReportNotFound when no report has the id.
	FindSummary(ctx context.Context, reportID string) (domain.Summary, error)
	UpdateSummary(ctx context.Context, s domain.Summary) error
	// ListSummariesByReporter returns the summaries of one reporter, in any order.
	ListSummariesByReporter(ctx context.Context, reporterID string) ([]domain.Summary, error)
	AppendEntry(ctx context.Context, e domain.Entry) error
	// ListEntries returns the entries of one report in the order they were appended.
	ListEntries(ctx context.Context, reportID string) ([]domain.Entry, error)
}
