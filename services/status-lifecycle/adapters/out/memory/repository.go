// Package memory is an in-memory StatusRepository. Data is lost on restart.
// It is used for the Sprint 1 thin slice and in tests.
package memory

import (
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

// Repository is safe for use by several gRPC calls at the same time.
type Repository struct {
	mu   sync.Mutex
	data map[string]domain.ReportStatus
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{data: make(map[string]domain.ReportStatus)}
}

func (r *Repository) Get(_ context.Context, reportID string) (domain.ReportStatus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rs, ok := r.data[reportID]
	if !ok {
		return domain.ReportStatus{}, domain.ErrNotFound
	}
	return rs, nil
}

func (r *Repository) Create(_ context.Context, rs domain.ReportStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[rs.ReportID]; ok {
		return domain.ErrAlreadyExists
	}
	r.data[rs.ReportID] = rs
	return nil
}

func (r *Repository) Update(_ context.Context, rs domain.ReportStatus, expectedVersion int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.data[rs.ReportID]
	if !ok {
		return domain.ErrNotFound
	}
	if stored.Version != expectedVersion {
		return domain.ErrVersionConflict
	}
	r.data[rs.ReportID] = rs
	return nil
}
