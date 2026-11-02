package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

// Errors returned by the use cases, on top of the domain errors.
var (
	// ErrInvalidArgument: a field is missing or badly formed.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrNoPolicy: sla-policy has no window for the category, so no deadline can be set.
	ErrNoPolicy = ports.ErrNoPolicy
	// ErrPolicyUnavailable: sla-policy could not be reached.
	ErrPolicyUnavailable = errors.New("sla-policy unavailable")
)

// ListDueTimers limits, from the .proto file.
const (
	defaultLimit = 100
	maxLimit     = 500
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service holds the sla-tracking use cases.
type Service struct {
	repo      ports.TimerRepository
	policy    ports.PolicyClient
	publisher ports.EventPublisher
	now       func() time.Time

	// lapseMu stops two checker runs from lapsing the same timer twice.
	// The Redis adapter will do this with an atomic ZREM instead.
	lapseMu sync.Mutex
}

// NewService wires the use cases. now is time.Now outside tests.
func NewService(repo ports.TimerRepository, policy ports.PolicyClient, publisher ports.EventPublisher, now func() time.Time) *Service {
	return &Service{repo: repo, policy: policy, publisher: publisher, now: now}
}

// StartTimer starts the SLA timer of a report. The deadline is now plus the
// window from sla-policy for the report's category.
func (s *Service) StartTimer(ctx context.Context, reportID string, category domain.Category) (domain.Timer, error) {
	if err := checkReportID(reportID); err != nil {
		return domain.Timer{}, err
	}
	if !category.Valid() {
		return domain.Timer{}, fmt.Errorf("%w: %d", domain.ErrInvalidCategory, int32(category))
	}
	window, err := s.policy.Window(ctx, category)
	switch {
	case errors.Is(err, ports.ErrNoPolicy):
		return domain.Timer{}, fmt.Errorf("%w: %s", ErrNoPolicy, category)
	case err != nil:
		return domain.Timer{}, fmt.Errorf("%w: %v", ErrPolicyUnavailable, err)
	}
	t, err := domain.NewTimer(reportID, category, window, s.now().UTC())
	if err != nil {
		return domain.Timer{}, err
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return domain.Timer{}, err
	}
	return t, nil
}

// StopTimer stops the timer when the report is resolved.
func (s *Service) StopTimer(ctx context.Context, reportID string) (domain.Timer, error) {
	t, err := s.GetTimer(ctx, reportID)
	if err != nil {
		return domain.Timer{}, err
	}
	t.Stop()
	if err := s.repo.Update(ctx, t); err != nil {
		return domain.Timer{}, err
	}
	return t, nil
}

// GetTimer returns the timer of a report.
func (s *Service) GetTimer(ctx context.Context, reportID string) (domain.Timer, error) {
	if err := checkReportID(reportID); err != nil {
		return domain.Timer{}, err
	}
	return s.repo.Get(ctx, reportID)
}

// ListDueTimers finds running timers whose deadline is at or before now,
// marks each one LAPSED and publishes sla.lapsed for it. It returns the timers
// it lapsed, so a timer is escalated only once. now is the zero time to use
// server time, and may not be later than server time. limit 0 means 100.
func (s *Service) ListDueTimers(ctx context.Context, now time.Time, limit int32) ([]domain.Timer, error) {
	serverNow := s.now().UTC()
	if now.IsZero() {
		now = serverNow
	}
	if now.After(serverNow) {
		return nil, fmt.Errorf("%w: now must not be in the future", ErrInvalidArgument)
	}
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 1 || limit > maxLimit {
		return nil, fmt.Errorf("%w: limit must be 1 to %d", ErrInvalidArgument, maxLimit)
	}

	s.lapseMu.Lock()
	defer s.lapseMu.Unlock()

	due, err := s.repo.ListDue(ctx, now, int(limit))
	if err != nil {
		return nil, err
	}
	lapsed := make([]domain.Timer, 0, len(due))
	for _, t := range due {
		if err := t.Lapse(now); err != nil {
			return lapsed, err
		}
		// Publish first: if saving fails, the next run publishes again
		// (at least once) instead of the event being lost.
		ev := ports.SLALapsed{ReportID: t.ReportID, Category: t.Category, Deadline: t.Deadline, LapsedAt: now}
		if err := s.publisher.PublishSLALapsed(ctx, ev); err != nil {
			return lapsed, fmt.Errorf("publish sla.lapsed for %s: %w", t.ReportID, err)
		}
		if err := s.repo.Update(ctx, t); err != nil {
			return lapsed, fmt.Errorf("save lapsed timer %s: %w", t.ReportID, err)
		}
		lapsed = append(lapsed, t)
	}
	return lapsed, nil
}

func checkReportID(id string) error {
	if !uuidPattern.MatchString(id) {
		return fmt.Errorf("%w: report_id must be a UUID", ErrInvalidArgument)
	}
	return nil
}
