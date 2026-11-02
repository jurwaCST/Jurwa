// Package memory is an in-memory, append-only EventStore. It is replaced by
// MongoDB (collection audit_events) later.
package memory

import (
	"context"
	"maps"
	"sort"
	"sync"

	"github.com/jurwaCST/Jurwa/services/audit-log/domain"
)

// Store keeps every entry in append order, safe for concurrent use.
type Store struct {
	mu     sync.RWMutex
	events []domain.Event
}

// New returns an empty store.
func New() *Store {
	return &Store{}
}

func (s *Store) Append(_ context.Context, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.Details = maps.Clone(e.Details)
	s.events = append(s.events, e)
	return nil
}

func (s *Store) ListByReport(_ context.Context, reportID string) ([]domain.Event, error) {
	s.mu.RLock()
	out := make([]domain.Event, 0)
	for _, e := range s.events {
		if e.ReportID == reportID {
			e.Details = maps.Clone(e.Details)
			out = append(out, e)
		}
	}
	s.mu.RUnlock()

	// Stable keeps append order for entries with the same time.
	sort.SliceStable(out, func(i, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out, nil
}
