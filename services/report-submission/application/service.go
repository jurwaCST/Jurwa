package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
	"github.com/jurwaCST/Jurwa/services/report-submission/ports"
)

// Service holds the report-submission use cases.
type Service struct {
	repo ports.ReportRepository
	ids  ports.IDGenerator
	pub  ports.EventPublisher
	now  func() time.Time
}

// NewService wires the use cases to their ports. now is usually time.Now.
func NewService(repo ports.ReportRepository, ids ports.IDGenerator, pub ports.EventPublisher, now func() time.Time) *Service {
	return &Service{repo: repo, ids: ids, pub: pub, now: now}
}

// SubmitReport validates the form, stores the report as OPEN, publishes
// report.submitted and returns the report. A form rule error wraps
// domain.ErrInvalidReport and nothing is stored or published.
func (s *Service) SubmitReport(ctx context.Context, in domain.NewReportInput) (domain.Report, error) {
	report, err := domain.NewReport(s.ids.NewID(), in, s.now().UTC())
	if err != nil {
		return domain.Report{}, err
	}
	if err := s.repo.Save(ctx, report); err != nil {
		return domain.Report{}, fmt.Errorf("save report: %w", err)
	}
	if err := s.pub.ReportSubmitted(ctx, report); err != nil {
		return domain.Report{}, fmt.Errorf("publish report.submitted: %w", err)
	}
	return report, nil
}

// GetReport returns the report with the given id, or domain.ErrReportNotFound.
func (s *Service) GetReport(ctx context.Context, id string) (domain.Report, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateDescription changes the description of a report, only for the student
// who sent it (SCRUM-163). The api-gateway checks first that nobody works on it.
func (s *Service) UpdateDescription(ctx context.Context, reportID, reporterID, description string) (domain.Report, error) {
	report, err := s.repo.FindByID(ctx, reportID)
	if err != nil {
		return domain.Report{}, err
	}
	if err := report.CheckReporter(reporterID); err != nil {
		return domain.Report{}, err
	}
	changed, err := report.WithDescription(description)
	if err != nil {
		return domain.Report{}, err
	}
	if err := s.repo.Save(ctx, changed); err != nil {
		return domain.Report{}, fmt.Errorf("save report: %w", err)
	}
	return changed, nil
}

// WithdrawReport deletes a report for the student who sent it and publishes
// report.withdrawn, naming who held it (SCRUM-163).
func (s *Service) WithdrawReport(ctx context.Context, reportID, reporterID, assigneeID string) error {
	report, err := s.repo.FindByID(ctx, reportID)
	if err != nil {
		return err
	}
	if err := report.CheckReporter(reporterID); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, reportID); err != nil {
		return fmt.Errorf("delete report: %w", err)
	}
	if err := s.pub.ReportWithdrawn(ctx, report, assigneeID); err != nil {
		return fmt.Errorf("publish report.withdrawn: %w", err)
	}
	return nil
}
