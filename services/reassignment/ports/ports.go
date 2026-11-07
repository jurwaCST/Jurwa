package ports

import (
	"context"
	"time"

	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

// AssignmentRepository stores the assignment history. In-memory for now;
// PostgreSQL (docs/design/schema/reassignment.sql) later.
type AssignmentRepository interface {
	// Current returns the current assignment. domain.ErrNotFound when there is none.
	Current(ctx context.Context, reportID string) (domain.Assignment, error)
	// Replace marks the assignment with id previousID as no longer current and
	// adds next as current, in one step. previousID "" means the report must have
	// no assignment yet (domain.ErrAlreadyAssigned otherwise). When the current
	// assignment is not previousID, nothing changes and it returns domain.ErrConflict.
	Replace(ctx context.Context, previousID string, next domain.Assignment) error
	// List returns all assignments of a report, oldest first.
	List(ctx context.Context, reportID string) ([]domain.Assignment, error)
}

// ReportReassigned is the payload of the report.reassigned event.
type ReportReassigned struct {
	AssignmentID string
	ReportID     string
	FromUserID   string
	ToUserID     string
	ActorID      string // "" when the system acted (escalation, compensation)
	Reason       string
	AssignedAt   time.Time
	// CorrelationID ties the event to one escalation Saga run. "" outside a Saga.
	CorrelationID string
}

// EventPublisher sends events to the event bus (Redis Streams later).
type EventPublisher interface {
	PublishReportReassigned(ctx context.Context, ev ReportReassigned) error
}
