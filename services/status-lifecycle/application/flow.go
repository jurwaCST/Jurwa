package application

// The start of a report's life, read from the event bus: report.submitted gives
// it OPEN, report.routed moves it to ASSIGNED. The two streams are read in
// parallel, so report.routed can arrive first.

import (
	"context"
	"errors"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

// OnReportSubmitted gives a new report the status OPEN. A report that is
// already known (a repeated event, or report.routed was read first) counts as done.
func (s *Service) OnReportSubmitted(ctx context.Context, reportID string) error {
	_, err := s.Register(ctx, reportID)
	if errors.Is(err, domain.ErrAlreadyExists) {
		return nil
	}
	return err
}

// OnReportRouted sets ASSIGNED as the system once routing has chosen the
// Category Head. An unknown report is registered first. A report that is
// already past OPEN (a repeated event) counts as done. A version conflict is
// returned, so the event bus tries again.
func (s *Service) OnReportRouted(ctx context.Context, reportID, targetPosition, correlationID string) error {
	if err := s.OnReportSubmitted(ctx, reportID); err != nil {
		return err
	}
	rs, err := s.repo.Get(ctx, reportID)
	if err != nil {
		return err
	}
	if rs.Status != domain.Open {
		return nil
	}
	_, err = s.ChangeStatus(ctx, ChangeStatusCommand{
		ReportID: reportID, To: domain.Assigned, Reason: "routed to " + targetPosition, CorrelationID: correlationID,
	})
	if errors.Is(err, domain.ErrInvalidTransition) {
		return nil
	}
	return err
}
