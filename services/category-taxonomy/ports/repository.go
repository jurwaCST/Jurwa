package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

// CategoryRepository stores the categories. Sprint 2 uses the in-memory adapter;
// the PostgreSQL adapter (docs/design/schema/category-taxonomy.sql) comes in Sprint 3.
type CategoryRepository interface {
	// List returns every category ordered by SortOrder.
	List(ctx context.Context) ([]domain.Info, error)
	// Save replaces an existing category. It returns domain.ErrNotFound for an
	// unknown code, because categories are never added at run time.
	Save(ctx context.Context, info domain.Info) error
}
