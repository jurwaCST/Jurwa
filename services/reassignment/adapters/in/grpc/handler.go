package grpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
	"github.com/jurwaCST/Jurwa/services/reassignment/application"
	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

// RoleMetadataKey is the gRPC metadata key that carries the caller's role, for
// example "STAFF". api-gateway sets it from the JWT (issue I8, same as
// internal-notes). A call without it has no role and is denied.
const RoleMetadataKey = "x-user-role"

// Handler implements ReassignmentService from proto/reassignment/v1.
type Handler struct {
	reassignmentv1.UnimplementedReassignmentServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) ReassignReport(ctx context.Context, req *reassignmentv1.ReassignReportRequest) (*reassignmentv1.ReassignReportResponse, error) {
	a, err := h.app.ReassignReport(ctx, application.ReassignCommand{
		ReportID: req.GetReportId(),
		ToUserID: req.GetToUserId(),
		Reason:   req.GetReason(),
		ActorID:  req.GetActorId(),
		Role:     roleFromContext(ctx),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &reassignmentv1.ReassignReportResponse{Assignment: toProto(a)}, nil
}

func (h *Handler) ListReassignments(ctx context.Context, req *reassignmentv1.ListReassignmentsRequest) (*reassignmentv1.ListReassignmentsResponse, error) {
	list, err := h.app.ListReassignments(ctx, roleFromContext(ctx), req.GetReportId())
	if err != nil {
		return nil, toStatusError(err)
	}
	out := make([]*reassignmentv1.Assignment, 0, len(list))
	for _, a := range list {
		out = append(out, toProto(a))
	}
	return &reassignmentv1.ListReassignmentsResponse{Assignments: out}, nil
}

// roleFromContext reads the caller's role from the incoming metadata, or "" when absent.
func roleFromContext(ctx context.Context) domain.Role {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get(RoleMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return domain.Role(strings.ToUpper(strings.TrimSpace(values[0])))
}

// toProto converts a domain assignment. Kind uses the same numbers as the proto enum.
func toProto(a domain.Assignment) *reassignmentv1.Assignment {
	out := &reassignmentv1.Assignment{
		Id:                   a.ID,
		ReportId:             a.ReportID,
		AssigneeUserId:       a.AssigneeUserID,
		AssigneePositionCode: a.AssigneePositionCode,
		Kind:                 reassignmentv1.AssignmentKind(a.Kind),
		Reason:               a.Reason,
		Current:              a.Current,
		AssignedAt:           timestamppb.New(a.AssignedAt),
	}
	if a.AssignedBy != "" {
		by := a.AssignedBy
		out.AssignedBy = &by
	}
	return out
}

// toStatusError maps errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidArgument),
		errors.Is(err, domain.ErrInvalidReason),
		errors.Is(err, domain.ErrSameAssignee):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, domain.ErrAlreadyAssigned):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
