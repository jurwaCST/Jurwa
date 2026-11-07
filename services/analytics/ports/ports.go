package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// ProjectionStore keeps the reports built from events.
// adapters/out/memory, or adapters/out/postgres when DATABASE_URL is set.
type ProjectionStore interface {
	// Apply records one event. applied is false for a repeat or unusable event.
	Apply(ctx context.Context, e domain.Event) (applied bool, err error)
	Reports(ctx context.Context) ([]domain.Report, error)
}
