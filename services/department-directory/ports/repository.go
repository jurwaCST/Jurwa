package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
)

// DirectoryRepository stores departments and positions. Sprint 2 uses the
// in-memory adapter; PostgreSQL (docs/design/schema/department-directory.sql) comes in Sprint 3.
type DirectoryRepository interface {
	// ListDepartments returns the departments ordered by code.
	ListDepartments(ctx context.Context) ([]domain.Department, error)
	// ListPositions returns every position ordered by code.
	ListPositions(ctx context.Context) ([]domain.Position, error)
}
