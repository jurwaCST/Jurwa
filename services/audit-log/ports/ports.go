package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/audit-log/domain"
)

// EventStore is append-only on purpose: there is no update or delete.
// In-memory for now; MongoDB collection audit_events later
// (docs/design/database-schema.md).
type EventStore interface {
	Append(ctx context.Context, e domain.Event) error
	// ListByReport returns the report's entries, oldest first.
	ListByReport(ctx context.Context, reportID string) ([]domain.Event, error)
}
