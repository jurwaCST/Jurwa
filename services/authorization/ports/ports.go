package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

// PolicyRepository loads the permission table.
// In-memory for now (the default table); a database in Sprint 3.
type PolicyRepository interface {
	Load(ctx context.Context) (domain.Policy, error)
}
