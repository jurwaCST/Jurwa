package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

// ErrInvalidArgument: a required field is missing or badly formed.
var ErrInvalidArgument = errors.New("invalid argument")

// Service holds the reassignment use cases.
type Service struct {
	repo  ports.AssignmentRepository
	pub   ports.EventPublisher
	now   func() time.Time
	newID func() string
}

// NewService wires the use cases. now is time.Now and newID is uuid.NewString outside tests.
func NewService(repo ports.AssignmentRepository, pub ports.EventPublisher, now func() time.Time, newID func() string) *Service {
	return &Service{repo: repo, pub: pub, now: now, newID: newID}
}

// RecordRouted saves the first assignment of a report. It is called for each
// report.routed event (and by the SEED_ASSIGNMENTS setting until the event bus exists).
func (s *Service) RecordRouted(ctx context.Context, reportID, assigneeUserID, positionCode string) (domain.Assignment, error) {
	if err := checkIDs("report_id", reportID, "assignee_user_id", assigneeUserID); err != nil {
		return domain.Assignment{}, err
	}
	a := domain.NewRouted(s.newID(), reportID, assigneeUserID, positionCode, s.now().UTC())
	if err := s.repo.Replace(ctx, "", a); err != nil {
		return domain.Assignment{}, err
	}
	return a, nil
}

// ReassignCommand is the input of ReassignReport. Role comes from the caller's JWT.
type ReassignCommand struct {
	ReportID string
	ToUserID string
	Reason   string
	ActorID  string
	Role     domain.Role
}

// ReassignReport hands a report to another user and publishes report.reassigned.
func (s *Service) ReassignReport(ctx context.Context, cmd ReassignCommand) (domain.Assignment, error) {
	if err := checkIDs("report_id", cmd.ReportID, "to_user_id", cmd.ToUserID, "actor_id", cmd.ActorID); err != nil {
		return domain.Assignment{}, err
	}
	current, err := s.repo.Current(ctx, cmd.ReportID)
	if err != nil {
		return domain.Assignment{}, err
	}
	if err := domain.CheckCanReassign(cmd.Role, cmd.ActorID, current); err != nil {
		return domain.Assignment{}, err
	}
	next, err := domain.Reassign(s.newID(), current, cmd.ToUserID, cmd.Reason, cmd.ActorID, s.now().UTC())
	if err != nil {
		return domain.Assignment{}, err
	}
	// Replace fails with ErrConflict if someone else reassigned it since Current.
	if err := s.repo.Replace(ctx, current.ID, next); err != nil {
		return domain.Assignment{}, err
	}
	ev := ports.ReportReassigned{
		AssignmentID: next.ID, ReportID: next.ReportID, FromUserID: current.AssigneeUserID,
		ToUserID: next.AssigneeUserID, ActorID: cmd.ActorID, Reason: next.Reason, AssignedAt: next.AssignedAt,
	}
	if err := s.pub.PublishReportReassigned(ctx, ev); err != nil {
		return domain.Assignment{}, fmt.Errorf("publish report.reassigned: %w", err)
	}
	return next, nil
}

// ListReassignments returns all assignments of a report, oldest first.
// An unknown report gives an empty list.
func (s *Service) ListReassignments(ctx context.Context, role domain.Role, reportID string) ([]domain.Assignment, error) {
	if err := domain.CheckCanList(role); err != nil {
		return nil, err
	}
	if err := checkIDs("report_id", reportID); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, reportID)
}

// checkIDs takes name, value pairs and checks them in order.
func checkIDs(pairs ...string) error {
	for i := 0; i+1 < len(pairs); i += 2 {
		if !domain.IsUUID(pairs[i+1]) {
			return fmt.Errorf("%w: %s must be a UUID", ErrInvalidArgument, pairs[i])
		}
	}
	return nil
}
