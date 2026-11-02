// Package clients adapts the sla-tracking and status-lifecycle gRPC clients
// to the checker's interfaces.
package clients

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jurwaCST/Jurwa/functions/sla-escalation-checker/checker"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slatrackingv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-tracking/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

// Default addresses for a local run.
const (
	DefaultSlaTrackingAddr     = "localhost:9012"
	DefaultStatusLifecycleAddr = "localhost:9011"
)

// SlaTracking implements checker.DueTimers.
type SlaTracking struct {
	c slatrackingv1.SlaTrackingServiceClient
}

// NewSlaTracking wraps the generated client.
func NewSlaTracking(c slatrackingv1.SlaTrackingServiceClient) *SlaTracking {
	return &SlaTracking{c: c}
}

// LapseDue calls ListDueTimers with server time. sla-tracking marks the
// returned timers LAPSED and publishes sla.lapsed for each.
func (s *SlaTracking) LapseDue(ctx context.Context, limit int) ([]checker.Lapsed, error) {
	resp, err := s.c.ListDueTimers(ctx, &slatrackingv1.ListDueTimersRequest{Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]checker.Lapsed, 0, len(resp.GetTimers()))
	for _, t := range resp.GetTimers() {
		out = append(out, checker.Lapsed{ReportID: t.GetReportId(), Deadline: t.GetDeadline().AsTime()})
	}
	return out, nil
}

// StatusLifecycle implements checker.StatusChanger.
type StatusLifecycle struct {
	c statuslifecyclev1.StatusLifecycleServiceClient
}

// NewStatusLifecycle wraps the generated client.
func NewStatusLifecycle(c statuslifecyclev1.StatusLifecycleServiceClient) *StatusLifecycle {
	return &StatusLifecycle{c: c}
}

// Escalate sets the report to ESCALATED with an empty actor_id (the system).
// FAILED_PRECONDITION means the flow does not allow it, for example the
// report is already RESOLVED or ESCALATED.
func (s *StatusLifecycle) Escalate(ctx context.Context, reportID, reason string) error {
	_, err := s.c.ChangeStatus(ctx, &statuslifecyclev1.ChangeStatusRequest{
		ReportId: reportID,
		ToStatus: commonv1.Status_ESCALATED,
		Reason:   reason,
	})
	if status.Code(err) == codes.FailedPrecondition {
		return fmt.Errorf("%w: %s", checker.ErrNotEscalatable, status.Convert(err).Message())
	}
	return err
}
