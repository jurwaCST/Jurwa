package grpcclients

import (
	"context"

	"google.golang.org/grpc/metadata"

	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
)

// RoleMetadataKey carries the caller's role to services whose messages have
// no role field (internal-notes).
const RoleMetadataKey = "x-user-role"

// Notes implements ports.Notes.
type Notes struct {
	C internalnotesv1.InternalNotesServiceClient
}

func (n Notes) AddNote(ctx context.Context, role string, in *internalnotesv1.AddNoteRequest) (*internalnotesv1.AddNoteResponse, error) {
	return n.C.AddNote(metadata.AppendToOutgoingContext(ctx, RoleMetadataKey, role), in)
}

func (n Notes) ListNotes(ctx context.Context, role, reportID string) (*internalnotesv1.ListNotesResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, RoleMetadataKey, role)
	return n.C.ListNotes(ctx, &internalnotesv1.ListNotesRequest{ReportId: reportID})
}
