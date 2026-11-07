package application

import (
	"context"
	"errors"
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
// EventID, when set, becomes the id of the first timeline entry.
type SubmittedInput struct {
	EventID    string
	ReportID   string
	ReporterID string
	Category   domain.Category
	Location   domain.Location
	CreatedAt  time.Time
}

// RecordSubmitted stores the summary of a new report and the first timeline entry.
// Called by the events:report.submitted consumer (adapters/in/events).
func (s *Service) RecordSubmitted(ctx context.Context, in SubmittedInput) (domain.Summary, error) {
	summary, err := domain.NewSummary(in.ReportID, in.ReporterID, in.Category, in.Location, in.CreatedAt.UTC())
	if err != nil {
		return domain.Summary{}, err
	}
	entryID := in.EventID
	if entryID == "" {
		entryID = s.newID()
	}
	entry, err := domain.NewEntry(entryID, domain.EntryInput{
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

// Record adds one timeline entry with a new id to a known report and updates
// its summary.
func (s *Service) Record(ctx context.Context, in domain.EntryInput) (domain.Entry, error) {
	return s.record(ctx, s.newID(), in)
}

// RecordEvent adds the timeline entry of one event, using the event id as the
// entry id. The same event delivered again is ignored (nil). Called by the
// report.routed, status.changed and report.reassigned consumers.
func (s *Service) RecordEvent(ctx context.Context, eventID string, in domain.EntryInput) error {
	_, err := s.record(ctx, eventID, in)
	if errors.Is(err, domain.ErrEntryExists) {
		return nil
	}
	return err
}

func (s *Service) record(ctx context.Context, id string, in domain.EntryInput) (domain.Entry, error) {
	in.OccurredAt = in.OccurredAt.UTC()
	entry, err := domain.NewEntry(id, in)
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

// ListReportsByAssignee returns one page of the reports currently assigned to
// one person, newest first, and the token of the next page ("" on the last page).
func (s *Service) ListReportsByAssignee(ctx context.Context, assigneeID string, pageSize int32, pageToken string) ([]domain.Summary, string, error) {
	if !domain.IsUUID(assigneeID) {
		return nil, "", domain.ErrInvalidAssigneeID
	}
	summaries, err := s.repo.ListSummariesByAssignee(ctx, assigneeID)
	if err != nil {
		return nil, "", fmt.Errorf("list summaries: %w", err)
	}
	summaries = domain.FilterByAssignee(summaries, assigneeID)
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
