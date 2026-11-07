// Package memory is an in-memory HistoryRepository. Data is lost on restart.
// It is used in Sprint 2 and in tests; PostgreSQL replaces it in Sprint 3.
package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

// Repository keeps summaries and timelines in maps. It is safe for concurrent use.
type Repository struct {
	mu        sync.RWMutex
	summaries map[string]domain.Summary
	entries   map[string][]domain.Entry // by report id, in append order
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{summaries: make(map[string]domain.Summary), entries: make(map[string][]domain.Entry)}
}

// CreateSummary stores a new summary or returns domain.ErrReportExists.
func (r *Repository) CreateSummary(_ context.Context, s domain.Summary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.summaries[s.ReportID]; ok {
		return domain.ErrReportExists
	}
	r.summaries[s.ReportID] = s
	return nil
}

// FindSummary returns the summary or domain.ErrReportNotFound.
func (r *Repository) FindSummary(_ context.Context, reportID string) (domain.Summary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.summaries[reportID]
	if !ok {
		return domain.Summary{}, domain.ErrReportNotFound
	}
	return s, nil
}

// UpdateSummary replaces a stored summary or returns domain.ErrReportNotFound.
func (r *Repository) UpdateSummary(_ context.Context, s domain.Summary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.summaries[s.ReportID]; !ok {
		return domain.ErrReportNotFound
	}
	r.summaries[s.ReportID] = s
	return nil
}

// ListSummariesByReporter returns the summaries of one reporter, in no fixed order.
func (r *Repository) ListSummariesByReporter(_ context.Context, reporterID string) ([]domain.Summary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Summary, 0)
	for _, s := range r.summaries {
		if s.ReporterID == reporterID {
			out = append(out, s)
		}
	}
	return out, nil
}

// ListSummariesByAssignee returns the summaries assigned to one person, in no fixed order.
func (r *Repository) ListSummariesByAssignee(_ context.Context, assigneeID string) ([]domain.Summary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Summary, 0)
	for _, s := range r.summaries {
		if s.AssigneeID == assigneeID {
			out = append(out, s)
		}
	}
	return out, nil
}

// AppendEntry adds an entry to the end of its report's timeline, or returns
// domain.ErrEntryExists when the report already has an entry with that id.
func (r *Repository) AppendEntry(_ context.Context, e domain.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, old := range r.entries[e.ReportID] {
		if old.ID == e.ID {
			return domain.ErrEntryExists
		}
	}
	r.entries[e.ReportID] = append(r.entries[e.ReportID], e)
	return nil
}

// ListEntries returns a copy of one report's entries, in append order.
func (r *Repository) ListEntries(_ context.Context, reportID string) ([]domain.Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.entries[reportID]), nil
}
