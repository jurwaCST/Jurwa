package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

// ErrInvalidArgument: a required field is missing or badly formed.
var ErrInvalidArgument = errors.New("invalid argument")

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service holds the status-lifecycle use cases.
type Service struct {
	repo ports.StatusRepository
	pub  ports.EventPublisher
	now  func() time.Time
}

// NewService wires the use cases to their ports. now is time.Now outside tests.
func NewService(repo ports.StatusRepository, pub ports.EventPublisher, now func() time.Time) *Service {
	return &Service{repo: repo, pub: pub, now: now}
}

// ChangeStatusCommand is the input of ChangeStatus.
type ChangeStatusCommand struct {
	ReportID        string
	To              domain.Status
	ActorID         string // "" when the system acts
	ExpectedVersion int32  // 0 means do not check
	Reason          string
}

// Register gives a newly submitted report the status OPEN.
// Sprint 2: called when a report.submitted event arrives.
func (s *Service) Register(ctx context.Context, reportID string) (domain.ReportStatus, error) {
	if err := checkID("report_id", reportID); err != nil {
		return domain.ReportStatus{}, err
	}
	rs := domain.NewReportStatus(reportID, s.now().UTC())
	if err := s.repo.Create(ctx, rs); err != nil {
		return domain.ReportStatus{}, err
	}
	return rs, nil
}

// GetStatus returns the current status of a report.
func (s *Service) GetStatus(ctx context.Context, reportID string) (domain.ReportStatus, error) {
	if err := checkID("report_id", reportID); err != nil {
		return domain.ReportStatus{}, err
	}
	return s.repo.Get(ctx, reportID)
}

// ChangeStatus checks the status flow, saves the change and publishes status.changed.
func (s *Service) ChangeStatus(ctx context.Context, cmd ChangeStatusCommand) (domain.ReportStatus, error) {
	if err := checkID("report_id", cmd.ReportID); err != nil {
		return domain.ReportStatus{}, err
	}
	if cmd.ActorID != "" {
		if err := checkID("actor_id", cmd.ActorID); err != nil {
			return domain.ReportStatus{}, err
		}
	}

	rs, err := s.repo.Get(ctx, cmd.ReportID)
	if err != nil {
		return domain.ReportStatus{}, err
	}
	if cmd.ExpectedVersion != 0 && cmd.ExpectedVersion != rs.Version {
		return domain.ReportStatus{}, domain.ErrVersionConflict
	}

	from, oldVersion := rs.Status, rs.Version
	if err := rs.ChangeTo(cmd.To, cmd.ActorID, s.now().UTC()); err != nil {
		return domain.ReportStatus{}, err
	}
	if err := s.repo.Update(ctx, rs, oldVersion); err != nil {
		return domain.ReportStatus{}, err
	}

	ev := ports.StatusChanged{
		ReportID: rs.ReportID, From: from, To: rs.Status,
		ActorID: cmd.ActorID, Version: rs.Version, Reason: cmd.Reason,
	}
	if err := s.pub.PublishStatusChanged(ctx, ev); err != nil {
		return domain.ReportStatus{}, fmt.Errorf("publish status.changed: %w", err)
	}
	return rs, nil
}

func checkID(field, value string) error {
	if !uuidPattern.MatchString(value) {
		return fmt.Errorf("%w: %s must be a UUID", ErrInvalidArgument, field)
	}
	return nil
}
