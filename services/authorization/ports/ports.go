package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

// PolicyRepository loads the permission table.
// adapters/out/memory holds the default table; adapters/out/postgres reads it
// from the database when DATABASE_URL is set.
type PolicyRepository interface {
	Load(ctx context.Context) (domain.Policy, error)
}
