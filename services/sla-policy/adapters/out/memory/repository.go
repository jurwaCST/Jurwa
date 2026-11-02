// Package memory is an in-memory PolicyRepository. Data is lost on restart.
package memory

import (
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
)

// Repository is safe for use by several gRPC calls at the same time.
type Repository struct {
	mu   sync.RWMutex
	data map[domain.Category]domain.Policy
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{data: make(map[domain.Category]domain.Policy)}
}

func (r *Repository) Get(_ context.Context, category domain.Category) (domain.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[category]
	if !ok {
		return domain.Policy{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *Repository) Save(_ context.Context, p domain.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.Category] = p
	return nil
}
