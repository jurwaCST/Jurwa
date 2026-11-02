package grpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	auditlogv1 "github.com/jurwaCST/Jurwa/proto/gen/audit-log/v1"
	"github.com/jurwaCST/Jurwa/services/audit-log/application"
	"github.com/jurwaCST/Jurwa/services/audit-log/domain"
)

// RoleMetadataKey is the gRPC metadata key that carries the caller's role.
// api-gateway sets it from the JWT (issue I8, same as internal-notes).
// Append is called by other services and needs no role; ListByReport is
// for the System Administrator only.
const RoleMetadataKey = "x-user-role"

// Handler implements AuditLogService from proto/audit-log/v1.
type Handler struct {
	auditlogv1.UnimplementedAuditLogServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) Append(ctx context.Context, req *auditlogv1.AppendRequest) (*auditlogv1.AppendResponse, error) {
	ev := req.GetEvent()
	if ev == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}
	in := domain.Input{
		Service:    ev.GetService(),
		Action:     ev.GetAction(),
		ReportID:   ev.GetReportId(),
		ActorID:    ev.GetActorId(),
		OnBehalfOf: ev.GetOnBehalfOf(),
		Details:    ev.GetDetails(),
	}
	if ev.GetOccurredAt() != nil {
		in.OccurredAt = ev.GetOccurredAt().AsTime()
	}
	id, err := h.app.Append(ctx, in)
	if err != nil {
		return nil, toStatusError(err)
	}
	return &auditlogv1.AppendResponse{Id: id}, nil
}

func (h *Handler) ListByReport(ctx context.Context, req *auditlogv1.ListByReportRequest) (*auditlogv1.ListByReportResponse, error) {
	list, err := h.app.ListByReport(ctx, roleFromContext(ctx), req.GetReportId())
	if err != nil {
		return nil, toStatusError(err)
	}
	out := make([]*auditlogv1.AuditEvent, 0, len(list))
	for _, e := range list {
		out = append(out, toProto(e))
	}
	return &auditlogv1.ListByReportResponse{Events: out}, nil
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

// toProto converts a stored entry; empty optional fields are left out.
func toProto(e domain.Event) *auditlogv1.AuditEvent {
	out := &auditlogv1.AuditEvent{
		Id:         e.ID,
		OccurredAt: timestamppb.New(e.OccurredAt),
		Service:    e.Service,
		Action:     string(e.Action),
		Details:    e.Details,
	}
	if e.ReportID != "" {
		out.ReportId = &e.ReportID
	}
	if e.ActorID != "" {
		out.ActorId = &e.ActorID
	}
	if e.OnBehalfOf != "" {
		out.OnBehalfOf = &e.OnBehalfOf
	}
	return out
}

// toStatusError maps errors to gRPC codes.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidArgument),
		errors.Is(err, domain.ErrInvalidEvent):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
