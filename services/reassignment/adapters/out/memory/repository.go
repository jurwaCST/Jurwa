// Package memory is an in-memory AssignmentRepository. It is replaced by
// PostgreSQL (docs/design/schema/reassignment.sql) later.
package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

// Repository keeps each report's assignments in order, safe for concurrent use.
type Repository struct {
	mu      sync.RWMutex
	history map[string][]domain.Assignment // report id -> oldest first; the last one is current
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{history: make(map[string][]domain.Assignment)}
}

func (r *Repository) Current(_ context.Context, reportID string) (domain.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h := r.history[reportID]
	if len(h) == 0 {
		return domain.Assignment{}, domain.ErrNotFound
	}
	return h[len(h)-1], nil
}

func (r *Repository) Replace(_ context.Context, previousID string, next domain.Assignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	h := r.history[next.ReportID]
	switch {
	case previousID == "" && len(h) > 0:
		return domain.ErrAlreadyAssigned
	case previousID != "" && (len(h) == 0 || h[len(h)-1].ID != previousID):
		return domain.ErrConflict
	}
	if len(h) > 0 {
		h[len(h)-1].Current = false
	}
	next.Current = true
	r.history[next.ReportID] = append(h, next)
	return nil
}

func (r *Repository) List(_ context.Context, reportID string) ([]domain.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.history[reportID]), nil
}
