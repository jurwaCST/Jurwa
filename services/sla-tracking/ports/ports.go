package ports

// Mocks for interaction tests (gomock), next to the hand-written fakes.
// Regenerate with: go generate ./services/sla-tracking/ports/
//go:generate mockgen -source=ports.go -destination=mocks/ports_mock.go -package=mocks

import (
	"context"
	"errors"
	"time"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

// TimerRepository stores one timer per report. In-memory for now; Redis
// (keys sla:deadlines and sla:timer:<report_id>) later.
type TimerRepository interface {
	// Create saves a new timer. domain.ErrAlreadyExists when the report has one.
	Create(ctx context.Context, t domain.Timer) error
	// Get returns the timer. domain.ErrNotFound when there is none.
	Get(ctx context.Context, reportID string) (domain.Timer, error)
	// Update replaces an existing timer. domain.ErrNotFound when there is none.
	Update(ctx context.Context, t domain.Timer) error
	// ListDue returns up to limit running timers with deadline at or before now,
	// earliest deadline first.
	ListDue(ctx context.Context, now time.Time, limit int) ([]domain.Timer, error)
}

// ErrNoPolicy: sla-policy has no window for the category yet.
var ErrNoPolicy = errors.New("no SLA window set for this category")

// PolicyClient asks sla-policy for the SLA window of a category.
type PolicyClient interface {
	// Window returns the SLA window. ErrNoPolicy when none is set.
	Window(ctx context.Context, category domain.Category) (time.Duration, error)
}

// SLALapsed is the payload of the sla.lapsed event.
type SLALapsed struct {
	ReportID string
	Category domain.Category
	Deadline time.Time
	LapsedAt time.Time
}

// EventPublisher sends events to the event bus (Redis Streams later).
type EventPublisher interface {
	PublishSLALapsed(ctx context.Context, ev SLALapsed) error
}
