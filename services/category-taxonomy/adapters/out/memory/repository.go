// Package memory is an in-memory CategoryRepository, seeded with the three
// categories. Data is lost when the service stops.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

// Repository implements ports.CategoryRepository. It is safe for concurrent use.
type Repository struct {
	mu   sync.RWMutex
	cats map[domain.Category]domain.Info
}

// NewRepository returns a repository holding the given categories.
func NewRepository(seed []domain.Info) *Repository {
	r := &Repository{cats: make(map[domain.Category]domain.Info)}
	for _, c := range seed {
		r.cats[c.Category] = c
	}
	return r
}

func (r *Repository) List(_ context.Context) ([]domain.Info, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Info, 0, len(r.cats))
	for _, c := range r.cats {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	return out, nil
}

func (r *Repository) Save(_ context.Context, info domain.Info) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cats[info.Category]; !ok {
		return fmt.Errorf("%w: %q", domain.ErrNotFound, info.Category)
	}
	r.cats[info.Category] = info
	return nil
}
