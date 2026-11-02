package grpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
)

// RoleMetadataKey is the gRPC metadata key that carries the caller's role, for
// example "STAFF". api-gateway sets it from the JWT; the proto messages have no
// role field. A call without it is treated as having no role and is denied.
const RoleMetadataKey = "x-user-role"

// UseCases is what the handler needs from the application layer.
type UseCases interface {
	AddNote(ctx context.Context, in domain.NewNoteInput) (domain.Note, error)
	ListNotes(ctx context.Context, role domain.Role, reportID string) ([]domain.Note, error)
}

// Handler implements internalnotesv1.InternalNotesServiceServer.
type Handler struct {
	internalnotesv1.UnimplementedInternalNotesServiceServer
	app UseCases
}

// NewHandler returns a gRPC handler backed by the use cases.
func NewHandler(app UseCases) *Handler {
	return &Handler{app: app}
}

// AddNote stores one staff-only note. A student or missing role returns
// PERMISSION_DENIED, a blank or too long body INVALID_ARGUMENT.
func (h *Handler) AddNote(ctx context.Context, req *internalnotesv1.AddNoteRequest) (*internalnotesv1.AddNoteResponse, error) {
	note, err := h.app.AddNote(ctx, domain.NewNoteInput{
		ReportID: req.GetReportId(),
		AuthorID: req.GetAuthorId(),
		Role:     roleFromContext(ctx),
		Body:     req.GetBody(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &internalnotesv1.AddNoteResponse{Note: noteToProto(note)}, nil
}

// ListNotes returns the notes of one report, oldest first. A student or missing
// role returns PERMISSION_DENIED.
func (h *Handler) ListNotes(ctx context.Context, req *internalnotesv1.ListNotesRequest) (*internalnotesv1.ListNotesResponse, error) {
	notes, err := h.app.ListNotes(ctx, roleFromContext(ctx), req.GetReportId())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &internalnotesv1.ListNotesResponse{Notes: make([]*internalnotesv1.Note, 0, len(notes))}
	for _, n := range notes {
		resp.Notes = append(resp.Notes, noteToProto(n))
	}
	return resp, nil
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

// toStatus maps domain errors to gRPC status codes and hides internal details.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func noteToProto(n domain.Note) *internalnotesv1.Note {
	return &internalnotesv1.Note{
		Id:        n.ID,
		ReportId:  n.ReportID,
		AuthorId:  n.AuthorID,
		Body:      n.Body,
		CreatedAt: timestamppb.New(n.CreatedAt),
	}
}
