package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

// RuleRepository stores routing rules. Sprint 1 uses the in-memory adapter;
// the PostgreSQL adapter (docs/design/schema/routing-rules.sql) comes in Sprint 2.
type RuleRepository interface {
	// FindByCategory returns every rule of one category, active or not.
	FindByCategory(ctx context.Context, category domain.Category) ([]domain.Rule, error)
	// List returns all rules, ordered by category then building.
	List(ctx context.Context, includeInactive bool) ([]domain.Rule, error)
	// Upsert creates or replaces the rule for (category, building) and returns it with its id.
	Upsert(ctx context.Context, rule domain.Rule) (domain.Rule, error)
}
