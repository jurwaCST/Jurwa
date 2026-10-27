package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

// Handler implements StatusLifecycleService from proto/status-lifecycle/v1.
type Handler struct {
	statuslifecyclev1.UnimplementedStatusLifecycleServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) ChangeStatus(ctx context.Context, req *statuslifecyclev1.ChangeStatusRequest) (*statuslifecyclev1.ChangeStatusResponse, error) {
	rs, err := h.app.ChangeStatus(ctx, application.ChangeStatusCommand{
		ReportID:        req.GetReportId(),
		To:              domain.Status(req.GetToStatus()),
		ActorID:         req.GetActorId(),
		ExpectedVersion: req.GetExpectedVersion(),
		Reason:          req.GetReason(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &statuslifecyclev1.ChangeStatusResponse{Status: toProto(rs)}, nil
}

func (h *Handler) GetStatus(ctx context.Context, req *statuslifecyclev1.GetStatusRequest) (*statuslifecyclev1.GetStatusResponse, error) {
	rs, err := h.app.GetStatus(ctx, req.GetReportId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &statuslifecyclev1.GetStatusResponse{Status: toProto(rs)}, nil
}

// toProto converts the domain entity to the protobuf message.
// domain.Status uses the same numbers as commonv1.Status.
func toProto(rs domain.ReportStatus) *statuslifecyclev1.ReportStatus {
	out := &statuslifecyclev1.ReportStatus{
		ReportId:       rs.ReportID,
		Status:         commonv1.Status(rs.Status),
		PreviousStatus: commonv1.Status(rs.Previous),
		Version:        rs.Version,
		UpdatedAt:      timestamppb.New(rs.UpdatedAt),
	}
	if rs.UpdatedBy != "" {
		updatedBy := rs.UpdatedBy
		out.UpdatedBy = &updatedBy
	}
	return out
}

// toStatusError maps domain errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidArgument), errors.Is(err, domain.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrVersionConflict):
		return status.Error(codes.Aborted, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
