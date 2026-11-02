// Package memory is an in-memory NoteRepository. Data is lost on restart.
// It is used in Sprint 2 and in tests; MongoDB replaces it in Sprint 3.
package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
)

// Repository keeps notes grouped by report. It is safe for concurrent use.
type Repository struct {
	mu       sync.RWMutex
	byReport map[string][]domain.Note
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{byReport: make(map[string][]domain.Note)}
}

// Save adds a note to its report.
func (r *Repository) Save(_ context.Context, note domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byReport[note.ReportID] = append(r.byReport[note.ReportID], note)
	return nil
}

// ListByReport returns a copy of one report's notes, in save order.
func (r *Repository) ListByReport(_ context.Context, reportID string) ([]domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.byReport[reportID]), nil
}
