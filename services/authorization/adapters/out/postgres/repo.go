// Package postgres is the PostgreSQL PolicyRepository (database authorization_db).
// cmd/server uses it when DATABASE_URL is set, otherwise the in-memory adapter.
package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

// schemaSQL is a copy of docs/design/schema/authorization.sql (a test checks
// they match). It is safe to run on every start: it creates the tables only if
// they are missing and seeds the default permission table only once.
//
//go:embed schema.sql
var schemaSQL string

// Repository implements ports.PolicyRepository.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the tables and seeds the default permissions if missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate authorization: %w", err)
	}
	return nil
}

// Load reads the whole permission table. A role name the service does not
// know is an error, so a bad row is found at start instead of being ignored.
func (r *Repository) Load(ctx context.Context) (domain.Policy, error) {
	rows, err := r.pool.Query(ctx, `SELECT action, role_code FROM role_permissions ORDER BY action, role_code`)
	if err != nil {
		return domain.Policy{}, fmt.Errorf("load permissions: %w", err)
	}
	defer rows.Close()

	rules := map[string][]domain.Role{}
	for rows.Next() {
		var action, code string
		if err := rows.Scan(&action, &code); err != nil {
			return domain.Policy{}, fmt.Errorf("read permission: %w", err)
		}
		role, ok := domain.ParseRole(code)
		if !ok {
			return domain.Policy{}, fmt.Errorf("permission %s: unknown role %q", action, code)
		}
		rules[action] = append(rules[action], role)
	}
	if err := rows.Err(); err != nil {
		return domain.Policy{}, fmt.Errorf("load permissions: %w", err)
	}
	return domain.NewPolicy(rules), nil
}
