package application

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/services/analytics/domain"
	"github.com/jurwaCST/Jurwa/services/analytics/ports"
)

// Service holds the analytics use cases. It is read-only for users: the
// counts change only through Record, which the event consumer calls.
type Service struct {
	store ports.ProjectionStore
}

// NewService wires the use cases to the store.
func NewService(store ports.ProjectionStore) *Service {
	return &Service{store: store}
}

// Record applies one event (report.submitted or status.changed). It returns
// false for a repeat or an event it cannot use. The Redis Streams consumer
// calls it in Sprint 3.
func (s *Service) Record(ctx context.Context, e domain.Event) (bool, error) {
	applied, err := s.store.Apply(ctx, e)
	if err != nil {
		return false, fmt.Errorf("record event: %w", err)
	}
	return applied, nil
}

// Summary returns the counts for reports created in the period.
func (s *Service) Summary(ctx context.Context, period domain.Period) (domain.Summary, error) {
	reports, err := s.reports(ctx, period)
	if err != nil {
		return domain.Summary{}, err
	}
	return domain.Summarize(reports, period), nil
}

// Breakdown returns the counts per category for reports created in the period.
func (s *Service) Breakdown(ctx context.Context, period domain.Period) ([]domain.CategoryCounts, error) {
	reports, err := s.reports(ctx, period)
	if err != nil {
		return nil, err
	}
	return domain.Breakdown(reports, period), nil
}

func (s *Service) reports(ctx context.Context, period domain.Period) ([]domain.Report, error) {
	if err := period.Validate(); err != nil {
		return nil, err
	}
	reports, err := s.store.Reports(ctx)
	if err != nil {
		return nil, fmt.Errorf("load reports: %w", err)
	}
	return reports, nil
}
