package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slatrackingv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-tracking/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/application"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

// Handler implements SlaTrackingService from proto/sla-tracking/v1.
type Handler struct {
	slatrackingv1.UnimplementedSlaTrackingServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) StartTimer(ctx context.Context, req *slatrackingv1.StartTimerRequest) (*slatrackingv1.StartTimerResponse, error) {
	t, err := h.app.StartTimer(ctx, req.GetReportId(), domain.Category(req.GetCategory()))
	if err != nil {
		return nil, toStatusError(err)
	}
	return &slatrackingv1.StartTimerResponse{Timer: toProto(t)}, nil
}

func (h *Handler) StopTimer(ctx context.Context, req *slatrackingv1.StopTimerRequest) (*slatrackingv1.StopTimerResponse, error) {
	t, err := h.app.StopTimer(ctx, req.GetReportId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &slatrackingv1.StopTimerResponse{Timer: toProto(t)}, nil
}

func (h *Handler) GetTimer(ctx context.Context, req *slatrackingv1.GetTimerRequest) (*slatrackingv1.GetTimerResponse, error) {
	t, err := h.app.GetTimer(ctx, req.GetReportId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &slatrackingv1.GetTimerResponse{Timer: toProto(t)}, nil
}

func (h *Handler) ListDueTimers(ctx context.Context, req *slatrackingv1.ListDueTimersRequest) (*slatrackingv1.ListDueTimersResponse, error) {
	var now time.Time
	if req.GetNow() != nil {
		now = req.GetNow().AsTime()
	}
	timers, err := h.app.ListDueTimers(ctx, now, req.GetLimit())
	if err != nil {
		return nil, toStatusError(err)
	}
	out := make([]*slatrackingv1.Timer, 0, len(timers))
	for _, t := range timers {
		out = append(out, toProto(t))
	}
	return &slatrackingv1.ListDueTimersResponse{Timers: out}, nil
}

// toProto converts a domain timer. Category and State use the same numbers as the proto enums.
func toProto(t domain.Timer) *slatrackingv1.Timer {
	return &slatrackingv1.Timer{
		ReportId:  t.ReportID,
		Category:  commonv1.Category(t.Category),
		StartedAt: timestamppb.New(t.StartedAt),
		Deadline:  timestamppb.New(t.Deadline),
		State:     slatrackingv1.TimerState(t.State),
	}
}

// toStatusError maps errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidArgument),
		errors.Is(err, domain.ErrInvalidCategory):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, application.ErrNoPolicy),
		errors.Is(err, domain.ErrInvalidWindow):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, application.ErrPolicyUnavailable):
		return status.Error(codes.Unavailable, "sla-policy unavailable")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
