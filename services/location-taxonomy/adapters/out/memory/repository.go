// Package memory is an in-memory LocationRepository. The list is fixed when the
// repository is created, so reads need no lock.
package memory

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

// Repository keeps the location list in memory, grouped by category.
type Repository struct {
	byCategory map[domain.Category][]domain.Entry
}

// NewRepository stores a copy of entries, keeping their order.
func NewRepository(entries []domain.Entry) *Repository {
	byCategory := make(map[domain.Category][]domain.Entry)
	for _, e := range entries {
		byCategory[e.Category] = append(byCategory[e.Category], e)
	}
	return &Repository{byCategory: byCategory}
}

// ListByCategory returns a copy of every entry of the category, active or not.
func (r *Repository) ListByCategory(_ context.Context, category domain.Category) ([]domain.Entry, error) {
	stored := r.byCategory[category]
	out := make([]domain.Entry, len(stored))
	copy(out, stored)
	return out, nil
}
