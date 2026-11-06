// Package postgres is the PostgreSQL CategoryRepository (database
// category_taxonomy). cmd/server uses it when DATABASE_URL is set, otherwise
// the in-memory adapter.
package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

// schemaSQL is a copy of docs/design/schema/category-taxonomy.sql (a test
// checks they match). It is safe to run on every start.
//
//go:embed schema.sql
var schemaSQL string

// Repository implements ports.CategoryRepository.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the table and seeds the three categories if they are missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate categories: %w", err)
	}
	return nil
}

// List returns every category ordered by sort order.
func (r *Repository) List(ctx context.Context) ([]domain.Info, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, name, description, sort_order, is_active, head_position
		FROM categories ORDER BY sort_order, code`)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	cats, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Info, error) {
		var c domain.Info
		var code string
		err := row.Scan(&code, &c.Name, &c.Description, &c.SortOrder, &c.Active, &c.HeadPosition)
		c.Category = domain.Category(code)
		return c, err
	})
	if err != nil {
		return nil, fmt.Errorf("read categories: %w", err)
	}
	return cats, nil
}

// Save replaces an existing category. Categories are never added at run time,
// so an unknown code is domain.ErrNotFound.
func (r *Repository) Save(ctx context.Context, info domain.Info) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE categories
		SET name = $2, description = $3, sort_order = $4, is_active = $5, head_position = $6, updated_at = now()
		WHERE code = $1`,
		string(info.Category), info.Name, info.Description, info.SortOrder, info.Active, info.HeadPosition)
	if err != nil {
		return fmt.Errorf("save category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %q", domain.ErrNotFound, info.Category)
	}
	return nil
}
