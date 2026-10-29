package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

// LocationRepository reads the location list. The in-memory adapter is used in
// Sprint 2; a PostgreSQL adapter (docs/design/schema/location-taxonomy.sql) can replace it.
type LocationRepository interface {
	// ListByCategory returns every entry of the category, active or not.
	ListByCategory(ctx context.Context, category domain.Category) ([]domain.Entry, error)
}
