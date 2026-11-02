// Package memory is an in-memory ProjectionStore. Data is lost on restart.
// A database replaces it in Sprint 3.
package memory

import (
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// Store is safe for use by several calls at the same time.
type Store struct {
	mu sync.RWMutex
	p  *domain.Projection
}

// New returns an empty store.
func New() *Store {
	return &Store{p: domain.NewProjection()}
}

func (s *Store) Apply(_ context.Context, e domain.Event) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.p.Apply(e), nil
}

func (s *Store) Reports(_ context.Context) ([]domain.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.p.Reports(), nil
}
