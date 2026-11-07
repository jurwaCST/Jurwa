// Package postgres is the PostgreSQL UserRepository (database identity).
// cmd/server uses it when DATABASE_URL is set, otherwise the in-memory adapter.
package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

// schemaSQL is a copy of docs/design/schema/identity.sql (a test checks they
// match). It is safe to run on every start: it only creates what is missing.
//
//go:embed schema.sql
var schemaSQL string

// uniqueViolation is the PostgreSQL error code for a duplicate key.
const uniqueViolation = "23505"

// Repository implements ports.UserRepository.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the users table if it is missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate users: %w", err)
	}
	return nil
}

// Create saves a new user. A used email (in any case) gives domain.ErrEmailTaken.
func (r *Repository) Create(ctx context.Context, u domain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, full_name, password_hash, roles, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
		u.ID, u.Email, u.FullName, u.PasswordHash, roleNames(u.Roles), u.Active, u.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "users_email_lower_uq" {
		return domain.ErrEmailTaken
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

const columns = `id::text, email, full_name, password_hash, roles, is_active, created_at`

// FindByEmail matches the email in any case, like the unique index.
func (r *Repository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.findOne(ctx, `SELECT `+columns+` FROM users WHERE lower(email) = lower($1)`, email)
}

// FindByID returns domain.ErrUserNotFound for an id that is not a UUID,
// instead of a database error.
func (r *Repository) FindByID(ctx context.Context, id string) (domain.User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.User{}, domain.ErrUserNotFound
	}
	return r.findOne(ctx, `SELECT `+columns+` FROM users WHERE id = $1`, id)
}

func (r *Repository) findOne(ctx context.Context, sql string, arg string) (domain.User, error) {
	var u domain.User
	var roles []string
	err := r.pool.QueryRow(ctx, sql, arg).Scan(&u.ID, &u.Email, &u.FullName, &u.PasswordHash, &roles, &u.Active, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	u.CreatedAt = u.CreatedAt.UTC()
	for _, r := range roles {
		u.Roles = append(u.Roles, domain.Role(r))
	}
	return u, nil
}

func roleNames(roles []domain.Role) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = string(r)
	}
	return out
}
