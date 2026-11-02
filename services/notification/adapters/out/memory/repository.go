// Package memory is an in-memory NotificationRepository. Data is lost on
// restart. A database replaces it in Sprint 3.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

type entry struct {
	n   domain.Notification
	seq int // save order, breaks ties when two have the same CreatedAt
}

// Repository is safe for use by several gRPC calls at the same time.
type Repository struct {
	mu   sync.RWMutex
	data map[string]entry
	next int
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{data: make(map[string]entry)}
}

// Save adds the notification, or replaces the one with the same id.
func (r *Repository) Save(_ context.Context, n domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.data[n.ID]; ok {
		r.data[n.ID] = entry{n: n, seq: old.seq}
		return nil
	}
	r.next++
	r.data[n.ID] = entry{n: n, seq: r.next}
	return nil
}

func (r *Repository) ListByRecipient(_ context.Context, userID string, unreadOnly bool, offset, limit int) ([]domain.Notification, bool, error) {
	r.mu.RLock()
	var found []entry
	for _, e := range r.data {
		if e.n.RecipientUserID != userID {
			continue
		}
		if unreadOnly && e.n.ReadAt != nil {
			continue
		}
		found = append(found, e)
	}
	r.mu.RUnlock()

	// Newest first. Same time: the one saved later comes first.
	sort.Slice(found, func(i, j int) bool {
		if !found[i].n.CreatedAt.Equal(found[j].n.CreatedAt) {
			return found[i].n.CreatedAt.After(found[j].n.CreatedAt)
		}
		return found[i].seq > found[j].seq
	})

	if offset >= len(found) {
		return nil, false, nil
	}
	end := min(offset+limit, len(found))
	out := make([]domain.Notification, 0, end-offset)
	for _, e := range found[offset:end] {
		out = append(out, e.n)
	}
	return out, end < len(found), nil
}
