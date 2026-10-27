// Package memory is an in-memory ReportRepository. Data is lost on restart.
// It is used in Sprint 1 and in tests; MongoDB replaces it in Sprint 2.
package memory

import (
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

// Repository keeps reports in a map. It is safe for concurrent use.
type Repository struct {
	mu      sync.RWMutex
	reports map[string]domain.Report
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{reports: make(map[string]domain.Report)}
}

// Save stores the report, replacing any report with the same id.
func (r *Repository) Save(_ context.Context, report domain.Report) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports[report.ID] = report
	return nil
}

// FindByID returns the report or domain.ErrReportNotFound.
func (r *Repository) FindByID(_ context.Context, id string) (domain.Report, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	report, ok := r.reports[id]
	if !ok {
		return domain.Report{}, domain.ErrReportNotFound
	}
	return report, nil
}
