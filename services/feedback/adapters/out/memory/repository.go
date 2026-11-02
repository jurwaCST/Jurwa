// Package memory is an in-memory FeedbackRepository. Data is lost on
// restart. A database replaces it in Sprint 3.
package memory

import (
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

// Repository is safe for use by several gRPC calls at the same time.
type Repository struct {
	mu       sync.RWMutex
	byReport map[string]domain.Feedback
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{byReport: make(map[string]domain.Feedback)}
}

// Create checks and inserts under one lock, so two calls at the same time
// cannot both add feedback for the same report.
func (r *Repository) Create(_ context.Context, f domain.Feedback) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byReport[f.ReportID]; exists {
		return domain.ErrAlreadyExists
	}
	r.byReport[f.ReportID] = f
	return nil
}

func (r *Repository) GetByReport(_ context.Context, reportID string) (domain.Feedback, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.byReport[reportID]
	if !ok {
		return domain.Feedback{}, domain.ErrNotFound
	}
	return f, nil
}
