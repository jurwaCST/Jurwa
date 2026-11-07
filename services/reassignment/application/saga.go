package application

// reassignment's part of the escalation Saga (SCRUM-89,
// docs/design/escalation-saga.md): on sla.lapsed the report goes back up to the
// Category Head; on escalation.compensated that is undone.

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

// SagaEvent is what the event consumers pass in from one event.
type SagaEvent struct {
	EventID       string // meta.event_id, used to ignore a repeated event
	CorrelationID string // meta.correlation_id, the same for one Saga run
	ReportID      string
	Reason        string // why a step failed (escalation.compensated only)
}

// sagaMemory remembers handled event ids and undone Saga runs. It is in memory:
// after a restart a repeated event is caught by the assignment checks instead.
type sagaMemory struct {
	// step makes EscalateOnLapse and CompensateEscalation run one at a time, so
	// an undo can never slip in between "is this run undone?" and the escalation.
	step   sync.Mutex
	mu     sync.Mutex
	seen   map[string]bool // event ids already handled
	undone map[string]bool // correlation ids whose escalation was undone
}

func (m *sagaMemory) handled(eventID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seen[eventID]
}

func (m *sagaMemory) markHandled(eventID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen == nil {
		m.seen = map[string]bool{}
	}
	m.seen[eventID] = true
}

func (m *sagaMemory) markUndone(correlationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.undone == nil {
		m.undone = map[string]bool{}
	}
	m.undone[correlationID] = true
}

func (m *sagaMemory) isUndone(correlationID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.undone[correlationID]
}

func checkSagaEvent(ev SagaEvent) error {
	if ev.EventID == "" {
		return fmt.Errorf("%w: event_id is required", ErrInvalidArgument)
	}
	return checkIDs("report_id", ev.ReportID)
}

// EscalateOnLapse hands the report back up to the Category Head after its SLA
// lapsed, and publishes report.reassigned with no actor (the system acted).
// It does nothing for a repeated event, an unknown report, or a Saga run whose
// escalation was already undone (the compensation can arrive first).
// A returned error makes the event bus try the event again.
func (s *Service) EscalateOnLapse(ctx context.Context, ev SagaEvent) error {
	if err := checkSagaEvent(ev); err != nil {
		return err
	}
	s.saga.step.Lock()
	defer s.saga.step.Unlock()
	if s.saga.handled(ev.EventID) || s.saga.isUndone(ev.CorrelationID) {
		return nil
	}
	history, err := s.repo.List(ctx, ev.ReportID)
	if err != nil {
		return err
	}
	head, err := domain.CategoryHead(history)
	if errors.Is(err, domain.ErrNotFound) {
		s.saga.markHandled(ev.EventID) // routing never assigned it: nothing to hand up
		return nil
	}
	current := history[len(history)-1]
	next := domain.NewEscalated(s.newID(), head, s.now().UTC())
	if err := s.repo.Replace(ctx, current.ID, next); err != nil {
		return err
	}
	s.saga.markHandled(ev.EventID)
	return s.publish(ctx, current, next, ev.CorrelationID)
}

// CompensateEscalation undoes the escalation of a Saga run when a later step
// failed: the report goes back to the person who held it before.
func (s *Service) CompensateEscalation(ctx context.Context, ev SagaEvent) error {
	if err := checkSagaEvent(ev); err != nil {
		return err
	}
	s.saga.step.Lock()
	defer s.saga.step.Unlock()
	if s.saga.handled(ev.EventID) {
		return nil
	}
	s.saga.markUndone(ev.CorrelationID) // a lapse of this run arriving later is skipped
	history, err := s.repo.List(ctx, ev.ReportID)
	if err != nil {
		return err
	}
	if len(history) < 2 || history[len(history)-1].Kind != domain.Escalated {
		s.saga.markHandled(ev.EventID) // nothing escalated (yet): nothing to undo
		return nil
	}
	current, previous := history[len(history)-1], history[len(history)-2]
	next, err := domain.NewCompensated(s.newID(), current, previous, ev.Reason, s.now().UTC())
	if err != nil {
		return err
	}
	if err := s.repo.Replace(ctx, current.ID, next); err != nil {
		return err
	}
	s.saga.markHandled(ev.EventID)
	return s.publish(ctx, current, next, ev.CorrelationID)
}

func (s *Service) publish(ctx context.Context, from, to domain.Assignment, correlationID string) error {
	ev := ports.ReportReassigned{
		AssignmentID: to.ID, ReportID: to.ReportID, FromUserID: from.AssigneeUserID, ToUserID: to.AssigneeUserID,
		Reason: to.Reason, AssignedAt: to.AssignedAt, CorrelationID: correlationID,
	}
	if err := s.pub.PublishReportReassigned(ctx, ev); err != nil {
		return fmt.Errorf("publish report.reassigned: %w", err)
	}
	return nil
}
