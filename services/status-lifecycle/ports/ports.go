package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

// StatusRepository stores the current status of each report.
// The in-memory adapter is used in Sprint 1; PostgreSQL comes in Sprint 2.
type StatusRepository interface {
	// Get returns domain.ErrNotFound when the report is unknown.
	Get(ctx context.Context, reportID string) (domain.ReportStatus, error)
	// Create returns domain.ErrAlreadyExists when the report already has a status.
	Create(ctx context.Context, rs domain.ReportStatus) error
	// Update saves rs only if the stored version still equals expectedVersion,
	// otherwise it returns domain.ErrVersionConflict (optimistic locking).
	Update(ctx context.Context, rs domain.ReportStatus, expectedVersion int32) error
}

// StatusChanged is published after every successful status change.
type StatusChanged struct {
	ReportID string
	From     domain.Status
	To       domain.Status
	ActorID  string // "" when the system acted
	Version  int32
	Reason   string
	// CorrelationID ties the event to one escalation Saga run. "" outside a Saga.
	CorrelationID string
}

// EscalationCompensated is published when status-lifecycle cannot set
// ESCALATED, so the earlier Saga step (reassignment's escalation) is undone.
type EscalationCompensated struct {
	ReportID      string
	FailedStep    string // "status-lifecycle.set-escalated"
	UndoneStep    string // "reassignment.escalate"
	Reason        string
	CorrelationID string
}

// EventPublisher sends events to other services (Redis Streams).
type EventPublisher interface {
	PublishStatusChanged(ctx context.Context, ev StatusChanged) error
	PublishEscalationCompensated(ctx context.Context, ev EscalationCompensated) error
}
