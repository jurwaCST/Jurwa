package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jurwaCST/Jurwa/services/feedback/domain"
	"github.com/jurwaCST/Jurwa/services/feedback/ports"
)

// Service holds the feedback use cases.
type Service struct {
	repo   ports.FeedbackRepository
	status ports.StatusChecker
	newID  func() string
	now    func() time.Time
}

// NewService wires the use cases to the store and the status checker.
func NewService(repo ports.FeedbackRepository, status ports.StatusChecker) *Service {
	return &Service{
		repo:   repo,
		status: status,
		newID:  func() string { return uuid.NewString() },
		now:    func() time.Time { return time.Now().UTC() },
	}
}

// Submit stores the student's rating. Order of checks:
//  1. the input (no network call for bad input)
//  2. one feedback per report (local, cheap)
//  3. the report is RESOLVED (asks status-lifecycle)
//
// Whether the student is the reporter is checked by the api-gateway
// ("Student (own report)" in docs/design/api-routes.md).
func (s *Service) Submit(ctx context.Context, reportID, studentID string, rating int32, comment string) (domain.Feedback, error) {
	f, err := domain.NewFeedback(s.newID(), reportID, studentID, rating, comment, s.now())
	if err != nil {
		return domain.Feedback{}, err
	}
	switch _, err := s.repo.GetByReport(ctx, reportID); {
	case err == nil:
		return domain.Feedback{}, domain.ErrAlreadyExists
	case !errors.Is(err, domain.ErrNotFound):
		return domain.Feedback{}, fmt.Errorf("check existing feedback: %w", err)
	}
	resolved, err := s.status.IsResolved(ctx, reportID)
	if err != nil {
		return domain.Feedback{}, err
	}
	if !resolved {
		return domain.Feedback{}, domain.ErrNotResolved
	}
	// Create checks again under a lock, so two calls at the same time cannot
	// both succeed.
	if err := s.repo.Create(ctx, f); err != nil {
		return domain.Feedback{}, err
	}
	return f, nil
}

// Get returns the feedback for a report.
func (s *Service) Get(ctx context.Context, reportID string) (domain.Feedback, error) {
	if err := domain.ValidateReportID(reportID); err != nil {
		return domain.Feedback{}, err
	}
	return s.repo.GetByReport(ctx, reportID)
}
