package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/report-history/domain"
	"github.com/jurwaCST/Jurwa/services/report-history/ports"
)

// Service holds the report-history use cases.
type Service struct {
	repo  ports.HistoryRepository
	newID func() string
}

// NewService wires the use cases to their port. newID is usually uuid.NewString.
func NewService(repo ports.HistoryRepository, newID func() string) *Service {
	return &Service{repo: repo, newID: newID}
}

// SubmittedInput is the part of a report.submitted event that report-history keeps.
type SubmittedInput struct {
	ReportID   string
	ReporterID string
	Category   domain.Category
	Location   domain.Location
	CreatedAt  time.Time
}

// RecordSubmitted stores the summary of a new report and the first timeline entry.
// TODO(Sprint 3): called by the events:report.submitted consumer on Redis Streams.
func (s *Service) RecordSubmitted(ctx context.Context, in SubmittedInput) (domain.Summary, error) {
	summary, err := domain.NewSummary(in.ReportID, in.ReporterID, in.Category, in.Location, in.CreatedAt.UTC())
	if err != nil {
		return domain.Summary{}, err
	}
	entry, err := domain.NewEntry(s.newID(), domain.EntryInput{
		ReportID:   in.ReportID,
		EventType:  domain.EventReportSubmitted,
		ToStatus:   domain.StatusOpen,
		ActorID:    in.ReporterID,
		Summary:    "Report submitted",
		OccurredAt: summary.CreatedAt,
	})
	if err != nil {
		return domain.Summary{}, err
	}
	if err := s.repo.CreateSummary(ctx, summary); err != nil {
		return domain.Summary{}, fmt.Errorf("create summary: %w", err)
	}
	if err := s.repo.AppendEntry(ctx, entry); err != nil {
		return domain.Summary{}, fmt.Errorf("append entry: %w", err)
	}
	return summary, nil
}

// Record adds one timeline entry to a known report and updates its summary.
// TODO(Sprint 3): called by the status.changed, report.reassigned and
// escalation.compensated consumers on Redis Streams.
func (s *Service) Record(ctx context.Context, in domain.EntryInput) (domain.Entry, error) {
	in.OccurredAt = in.OccurredAt.UTC()
	entry, err := domain.NewEntry(s.newID(), in)
	if err != nil {
		return domain.Entry{}, err
	}
	summary, err := s.repo.FindSummary(ctx, entry.ReportID)
	if err != nil {
		return domain.Entry{}, err
	}
	if err := s.repo.AppendEntry(ctx, entry); err != nil {
		return domain.Entry{}, fmt.Errorf("append entry: %w", err)
	}
	if err := s.repo.UpdateSummary(ctx, summary.Apply(entry)); err != nil {
		return domain.Entry{}, fmt.Errorf("update summary: %w", err)
	}
	return entry, nil
}

// ListReportsByReporter returns one page of a reporter's reports, newest first,
// and the token of the next page ("" on the last page).
func (s *Service) ListReportsByReporter(ctx context.Context, reporterID string, pageSize int32, pageToken string) ([]domain.Summary, string, error) {
	if !domain.IsUUID(reporterID) {
		return nil, "", domain.ErrInvalidReporterID
	}
	summaries, err := s.repo.ListSummariesByReporter(ctx, reporterID)
	if err != nil {
		return nil, "", fmt.Errorf("list summaries: %w", err)
	}
	summaries = domain.FilterByReporter(summaries, reporterID)
	domain.SortNewestFirst(summaries)
	return domain.Paginate(summaries, pageSize, pageToken)
}

// GetReportTimeline returns every entry of one report, oldest first, or
// domain.ErrReportNotFound when the report is unknown.
func (s *Service) GetReportTimeline(ctx context.Context, reportID string) ([]domain.Entry, error) {
	if !domain.IsUUID(reportID) {
		return nil, domain.ErrInvalidReportID
	}
	if _, err := s.repo.FindSummary(ctx, reportID); err != nil {
		return nil, err
	}
	entries, err := s.repo.ListEntries(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	domain.SortTimeline(entries)
	return entries, nil
}
