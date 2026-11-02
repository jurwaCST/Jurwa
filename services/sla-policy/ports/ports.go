package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
)

// PolicyRepository stores one SLA policy per category.
// In-memory for now; PostgreSQL (docs/design/schema/sla-policy.sql) later.
type PolicyRepository interface {
	// Get returns domain.ErrNotFound when no policy is set for the category.
	Get(ctx context.Context, category domain.Category) (domain.Policy, error)
	// Save creates the policy or replaces the existing one for its category.
	Save(ctx context.Context, p domain.Policy) error
}
