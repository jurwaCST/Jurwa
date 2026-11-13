// Package postgres is the PostgreSQL NotificationRepository (database
// notification). cmd/server uses it when DATABASE_URL is set, otherwise the
// in-memory adapter.
package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

// schemaSQL is a copy of docs/design/schema/notification.sql (a test checks
// they match). It is safe to run on every start: it only creates what is missing.
//
//go:embed schema.sql
var schemaSQL string

// Repository implements ports.NotificationRepository.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the notifications table if it is missing.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate notifications: %w", err)
	}
	return nil
}

// Save adds the notification, or replaces the one with the same id. A
// replaced row keeps its saved_seq, so its place in the list does not change.
func (r *Repository) Save(ctx context.Context, n domain.Notification) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (id, recipient_user_id, channel, event_type, report_id, subject, body,
		                           delivery_status, created_at, sent_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			recipient_user_id = EXCLUDED.recipient_user_id,
			channel           = EXCLUDED.channel,
			event_type        = EXCLUDED.event_type,
			report_id         = EXCLUDED.report_id,
			subject           = EXCLUDED.subject,
			body              = EXCLUDED.body,
			delivery_status   = EXCLUDED.delivery_status,
			created_at        = EXCLUDED.created_at,
			sent_at           = EXCLUDED.sent_at,
			read_at           = EXCLUDED.read_at`,
		n.ID, n.RecipientUserID, n.Channel.String(), n.EventType, nullString(n.ReportID), n.Subject, n.Body,
		n.Status.String(), n.CreatedAt, n.SentAt, n.ReadAt)
	if err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return nil
}

// Create adds n only when its id is new. The primary key on id makes the
// database refuse the second row, and DO NOTHING leaves the first one as it
// is, so a message made from an event delivered twice is stored once.
func (r *Repository) Create(ctx context.Context, n domain.Notification) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (id, recipient_user_id, channel, event_type, report_id, subject, body,
		                           delivery_status, created_at, sent_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO NOTHING`,
		n.ID, n.RecipientUserID, n.Channel.String(), n.EventType, nullString(n.ReportID), n.Subject, n.Body,
		n.Status.String(), n.CreatedAt, n.SentAt, n.ReadAt)
	if err != nil {
		return false, fmt.Errorf("create notification: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// Exists says whether a notification with this id is stored.
func (r *Repository) Exists(ctx context.Context, id string) (bool, error) {
	var has bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM notifications WHERE id = $1)`, id).Scan(&has); err != nil {
		return false, fmt.Errorf("find notification: %w", err)
	}
	return has, nil
}

// ListByRecipient returns one user's SENT notifications, newest first. It
// reads one row more than limit to know whether there are more.
func (r *Repository) ListByRecipient(ctx context.Context, userID string, unreadOnly bool, offset, limit int) ([]domain.Notification, bool, error) {
	if uuid.Validate(userID) != nil {
		return nil, false, nil // no such user can have rows
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, recipient_user_id::text, channel, event_type, COALESCE(report_id::text, ''),
		       subject, body, delivery_status, created_at, sent_at, read_at
		FROM notifications
		WHERE recipient_user_id = $1 AND delivery_status = 'SENT' AND (NOT $2 OR read_at IS NULL)
		ORDER BY created_at DESC, saved_seq DESC
		OFFSET $3 LIMIT $4`,
		userID, unreadOnly, offset, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("list notifications: %w", err)
	}
	list, err := pgx.CollectRows(rows, scan)
	if err != nil {
		return nil, false, fmt.Errorf("read notifications: %w", err)
	}
	if len(list) > limit {
		return list[:limit], true, nil
	}
	if len(list) == 0 {
		return nil, false, nil
	}
	return list, false, nil
}

func scan(row pgx.CollectableRow) (domain.Notification, error) {
	var n domain.Notification
	var channel, status string
	var sentAt, readAt *time.Time
	err := row.Scan(&n.ID, &n.RecipientUserID, &channel, &n.EventType, &n.ReportID,
		&n.Subject, &n.Body, &status, &n.CreatedAt, &sentAt, &readAt)
	if err != nil {
		return n, err
	}
	n.Channel = channels[channel]
	n.Status = statuses[status]
	n.CreatedAt = n.CreatedAt.UTC()
	n.SentAt = utc(sentAt)
	n.ReadAt = utc(readAt)
	return n, nil
}

var channels = map[string]domain.Channel{
	domain.Email.String(): domain.Email,
	domain.InApp.String(): domain.InApp,
}

var statuses = map[string]domain.DeliveryStatus{
	domain.Pending.String(): domain.Pending,
	domain.Sent.String():    domain.Sent,
	domain.Failed.String():  domain.Failed,
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func utc(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}
