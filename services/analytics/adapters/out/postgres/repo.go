// Package postgres is the PostgreSQL ProjectionStore (database analytics).
// cmd/server uses it when DATABASE_URL is set, otherwise the in-memory adapter.
package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// schemaSQL is a copy of docs/design/schema/analytics.sql (a test checks they
// match). It is safe to run on every start: it only creates what is missing.
//
//go:embed schema.sql
var schemaSQL string

// Repository implements ports.ProjectionStore. The Go type is called
// Repository like the other PostgreSQL adapters.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the tables if they are missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate analytics: %w", err)
	}
	return nil
}

// Apply records one event in one transaction: it marks the event id as seen,
// locks the report row, and asks domain.Next what changes. If the event is a
// repeat or cannot be used, the transaction is rolled back, so nothing is
// stored and the event id is not marked as seen.
func (r *Repository) Apply(ctx context.Context, e domain.Event) (bool, error) {
	if e.ID == "" || e.ReportID == "" {
		return false, nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("apply event: %w", err)
	}
	defer tx.Rollback(ctx) // does nothing after Commit

	tag, err := tx.Exec(ctx, `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`, e.ID)
	if err != nil {
		return false, fmt.Errorf("apply event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil // seen before
	}

	current, exists, err := find(ctx, tx, e.ReportID)
	if err != nil {
		return false, err
	}
	next, ok := domain.Next(current, exists, e)
	if !ok {
		return false, nil // rolled back by the deferred Rollback
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO report_facts (report_id, category_code, status, was_escalated, created_at, resolved_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (report_id) DO UPDATE SET
			status        = EXCLUDED.status,
			was_escalated = EXCLUDED.was_escalated,
			resolved_at   = EXCLUDED.resolved_at,
			updated_at    = now()`,
		next.ID, next.Category.String(), next.Status.String(), next.WasEscalated, next.CreatedAt, nullTime(next.ResolvedAt))
	if err != nil {
		return false, fmt.Errorf("save report: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("apply event: %w", err)
	}
	return true, nil
}

const columns = `report_id, category_code, status, was_escalated, created_at, resolved_at`

// find reads one report and locks its row until the transaction ends, so two
// events for the same report are applied one after the other.
func find(ctx context.Context, tx pgx.Tx, reportID string) (domain.Report, bool, error) {
	rows, err := tx.Query(ctx, `SELECT `+columns+` FROM report_facts WHERE report_id = $1 FOR UPDATE`, reportID)
	if err != nil {
		return domain.Report{}, false, fmt.Errorf("find report: %w", err)
	}
	rep, err := pgx.CollectExactlyOneRow(rows, scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Report{}, false, nil
	}
	if err != nil {
		return domain.Report{}, false, fmt.Errorf("find report: %w", err)
	}
	return rep, true, nil
}

// Reports returns every report, sorted by id byte by byte like the in-memory store.
func (r *Repository) Reports(ctx context.Context) ([]domain.Report, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+columns+` FROM report_facts ORDER BY report_id COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	reports, err := pgx.CollectRows(rows, scan)
	if err != nil {
		return nil, fmt.Errorf("read reports: %w", err)
	}
	return reports, nil
}

func scan(row pgx.CollectableRow) (domain.Report, error) {
	var rep domain.Report
	var category, status string
	var resolvedAt *time.Time
	if err := row.Scan(&rep.ID, &category, &status, &rep.WasEscalated, &rep.CreatedAt, &resolvedAt); err != nil {
		return rep, err
	}
	rep.Category = categories[category]
	rep.Status = statuses[status]
	rep.CreatedAt = rep.CreatedAt.UTC()
	if resolvedAt != nil {
		rep.ResolvedAt = resolvedAt.UTC()
	}
	return rep, nil
}

var categories = map[string]domain.Category{}
var statuses = map[string]domain.Status{}

func init() {
	for _, c := range domain.Categories {
		categories[c.String()] = c
	}
	for _, s := range []domain.Status{domain.Open, domain.Assigned, domain.InProgress, domain.Resolved, domain.Escalated} {
		statuses[s.String()] = s
	}
}

// nullTime is nil for the zero time, so the column stays NULL.
func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
