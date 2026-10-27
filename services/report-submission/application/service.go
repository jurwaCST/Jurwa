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
	now  func() time.Time
}

// NewService wires the use cases to their ports. now is usually time.Now.
func NewService(repo ports.ReportRepository, ids ports.IDGenerator, now func() time.Time) *Service {
	return &Service{repo: repo, ids: ids, now: now}
}

// SubmitReport validates the form, stores the report as OPEN and returns it.
// A form rule error wraps domain.ErrInvalidReport and nothing is stored.
// TODO(Sprint 2): publish events:report.submitted on Redis Streams.
func (s *Service) SubmitReport(ctx context.Context, in domain.NewReportInput) (domain.Report, error) {
	report, err := domain.NewReport(s.ids.NewID(), in, s.now().UTC())
	if err != nil {
		return domain.Report{}, err
	}
	if err := s.repo.Save(ctx, report); err != nil {
		return domain.Report{}, fmt.Errorf("save report: %w", err)
	}
	return report, nil
}

// GetReport returns the report with the given id, or domain.ErrReportNotFound.
func (s *Service) GetReport(ctx context.Context, id string) (domain.Report, error) {
	return s.repo.FindByID(ctx, id)
}
