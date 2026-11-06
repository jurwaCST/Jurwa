// Package postgres is the PostgreSQL RuleRepository (database routing_rules).
// cmd/server uses it when DATABASE_URL is set, otherwise the in-memory adapter.
package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

// schemaSQL is a copy of docs/design/schema/routing-rules.sql (a test checks
// they match). It is safe to run on every start: it creates the table only if
// it is missing and seeds the three default rules only once.
//
//go:embed schema.sql
var schemaSQL string

// Repository implements ports.RuleRepository.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the table and seeds the default rules if they are missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate routing_rules: %w", err)
	}
	return nil
}

// COLLATE "C" sorts byte by byte, the same order as the in-memory adapter.
const columns = `id::text, category_code, building, target_position, priority, is_active, updated_at`
const order = ` ORDER BY category_code COLLATE "C", building COLLATE "C"`

func (r *Repository) FindByCategory(ctx context.Context, category domain.Category) ([]domain.Rule, error) {
	return r.query(ctx, `SELECT `+columns+` FROM routing_rules WHERE category_code = $1`+order, string(category))
}

func (r *Repository) List(ctx context.Context, includeInactive bool) ([]domain.Rule, error) {
	return r.query(ctx, `SELECT `+columns+` FROM routing_rules WHERE is_active OR $1`+order, includeInactive)
}

// Upsert creates or replaces the rule for (category, building) and returns it
// with its id. Replacing keeps the existing id.
func (r *Repository) Upsert(ctx context.Context, rule domain.Rule) (domain.Rule, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO routing_rules (category_code, building, target_position, priority, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, now()))
		ON CONFLICT (category_code, building) DO UPDATE SET
			target_position = EXCLUDED.target_position,
			priority        = EXCLUDED.priority,
			is_active       = EXCLUDED.is_active,
			updated_at      = EXCLUDED.updated_at
		RETURNING id::text, updated_at`,
		string(rule.Category), rule.Building, rule.TargetPosition, rule.Priority, rule.Active, nullTime(rule),
	).Scan(&rule.ID, &rule.UpdatedAt)
	if err != nil {
		return domain.Rule{}, fmt.Errorf("upsert rule: %w", err)
	}
	return rule, nil
}

func (r *Repository) query(ctx context.Context, sql string, arg any) ([]domain.Rule, error) {
	rows, err := r.pool.Query(ctx, sql, arg)
	if err != nil {
		return nil, fmt.Errorf("query rules: %w", err)
	}
	rules, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Rule, error) {
		var rule domain.Rule
		var category string
		err := row.Scan(&rule.ID, &category, &rule.Building, &rule.TargetPosition, &rule.Priority, &rule.Active, &rule.UpdatedAt)
		rule.Category = domain.Category(category)
		return rule, err
	})
	if err != nil {
		return nil, fmt.Errorf("read rules: %w", err)
	}
	return rules, nil
}

// nullTime is nil when the rule has no update time, so the database uses now().
func nullTime(rule domain.Rule) any {
	if rule.UpdatedAt.IsZero() {
		return nil
	}
	return rule.UpdatedAt
}
