package application

// status-lifecycle's part of the escalation Saga (SCRUM-89,
// docs/design/escalation-saga.md): on sla.lapsed the report becomes ESCALATED.
// If that is not possible, escalation.compensated asks reassignment to undo
// its escalation.

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

// Saga step names written in escalation.compensated.
const (
	StepSetEscalated = "status-lifecycle.set-escalated"
	StepEscalate     = "reassignment.escalate"
)

// SagaEvent is what the event consumer passes in from one sla.lapsed event.
type SagaEvent struct {
	EventID       string
	CorrelationID string
	ReportID      string
}

// seenEvents remembers handled event ids, so a repeated event changes nothing.
type seenEvents struct {
	mu  sync.Mutex
	ids map[string]bool
}

func (s *seenEvents) has(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ids[id]
}

func (s *seenEvents) add(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ids == nil {
		s.ids = map[string]bool{}
	}
	s.ids[id] = true
}

// EscalateOnLapse sets the report to ESCALATED as the system. A report that is
// already ESCALATED (for example by the sla-escalation-checker function) counts
// as done. When the flow does not allow it (RESOLVED) or the report is unknown,
// it publishes escalation.compensated. A version conflict returns the error, so
// the event bus tries again.
func (s *Service) EscalateOnLapse(ctx context.Context, ev SagaEvent) error {
	if ev.EventID == "" {
		return fmt.Errorf("%w: event_id is required", ErrInvalidArgument)
	}
	if s.seen.has(ev.EventID) {
		return nil
	}
	_, err := s.ChangeStatus(ctx, ChangeStatusCommand{
		ReportID: ev.ReportID, To: domain.Escalated, Reason: "SLA lapsed", CorrelationID: ev.CorrelationID,
	})
	switch {
	case err == nil:
	case errors.Is(err, domain.ErrVersionConflict):
		return err
	case s.alreadyEscalated(ctx, ev.ReportID):
	case errors.Is(err, domain.ErrInvalidTransition), errors.Is(err, domain.ErrNotFound):
		if perr := s.pub.PublishEscalationCompensated(ctx, ports.EscalationCompensated{
			ReportID: ev.ReportID, FailedStep: StepSetEscalated, UndoneStep: StepEscalate,
			Reason: err.Error(), CorrelationID: ev.CorrelationID,
		}); perr != nil {
			return fmt.Errorf("publish escalation.compensated: %w", perr)
		}
	default:
		return err
	}
	s.seen.add(ev.EventID)
	return nil
}

func (s *Service) alreadyEscalated(ctx context.Context, reportID string) bool {
	rs, err := s.repo.Get(ctx, reportID)
	return err == nil && rs.Status == domain.Escalated
}
