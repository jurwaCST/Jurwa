// Package postgres is the PostgreSQL FeedbackRepository (database feedback).
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

	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

// schemaSQL is a copy of docs/design/schema/feedback.sql (a test checks they
// match). It is safe to run on every start: it only creates what is missing.
//
//go:embed schema.sql
var schemaSQL string

// uniqueViolation is the PostgreSQL error code for a duplicate key.
const uniqueViolation = "23505"

// Repository implements ports.FeedbackRepository.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the feedback table if it is missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate feedback: %w", err)
	}
	return nil
}

// Create saves the feedback. The unique constraint on report_id makes two
// calls at the same time safe: only one can win, the other gets
// domain.ErrAlreadyExists.
func (r *Repository) Create(ctx context.Context, f domain.Feedback) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO feedback (id, report_id, student_user_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		f.ID, f.ReportID, f.StudentUserID, f.Rating, f.Comment, f.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "feedback_report_uq" {
		return domain.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create feedback: %w", err)
	}
	return nil
}

// GetByReport returns domain.ErrNotFound for a report id that is not a UUID,
// instead of a database error.
func (r *Repository) GetByReport(ctx context.Context, reportID string) (domain.Feedback, error) {
	if uuid.Validate(reportID) != nil {
		return domain.Feedback{}, domain.ErrNotFound
	}
	var f domain.Feedback
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, report_id::text, student_user_id::text, rating, comment, created_at
		FROM feedback WHERE report_id = $1`, reportID,
	).Scan(&f.ID, &f.ReportID, &f.StudentUserID, &f.Rating, &f.Comment, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Feedback{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Feedback{}, fmt.Errorf("get feedback: %w", err)
	}
	f.CreatedAt = f.CreatedAt.UTC()
	return f, nil
}
