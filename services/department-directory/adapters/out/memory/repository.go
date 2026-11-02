// Package memory is an in-memory DirectoryRepository. Data is lost when the service stops.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
)

// Repository implements ports.DirectoryRepository. It is safe for concurrent use.
type Repository struct {
	mu        sync.RWMutex
	depts     []domain.Department
	positions []domain.Position
}

// NewRepository returns a repository holding copies of the given data.
func NewRepository(depts []domain.Department, positions []domain.Position) *Repository {
	r := &Repository{
		depts:     append([]domain.Department(nil), depts...),
		positions: append([]domain.Position(nil), positions...),
	}
	sort.Slice(r.depts, func(i, j int) bool { return r.depts[i].Code < r.depts[j].Code })
	sort.Slice(r.positions, func(i, j int) bool { return r.positions[i].Code < r.positions[j].Code })
	return r
}

func (r *Repository) ListDepartments(_ context.Context) ([]domain.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Department(nil), r.depts...), nil
}

func (r *Repository) ListPositions(_ context.Context) ([]domain.Position, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Position(nil), r.positions...), nil
}
