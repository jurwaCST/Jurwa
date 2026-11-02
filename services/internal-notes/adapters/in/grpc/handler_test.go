package grpc

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	"github.com/jurwaCST/Jurwa/services/internal-notes/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/internal-notes/application"
	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
)

const (
	noteID   = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	reportA  = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	authorID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
)

var t0 = time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

func newTestHandler() *Handler {
	return NewHandler(application.NewService(memory.NewRepository(), func() string { return noteID }, func() time.Time { return t0 }))
}

// as returns a context that carries the caller's role, as api-gateway sends it.
func as(role string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(RoleMetadataKey, role))
}

func addReq(body string) *internalnotesv1.AddNoteRequest {
	return &internalnotesv1.AddNoteRequest{ReportId: reportA, AuthorId: authorID, Body: body}
}

func TestAddThenList(t *testing.T) {
	h := newTestHandler()

	added, err := h.AddNote(as("CATEGORY_HEAD"), addReq("Forwarded to the maintenance councillor."))
	require.NoError(t, err)
	assert.Equal(t, noteID, added.GetNote().GetId())
	assert.Equal(t, reportA, added.GetNote().GetReportId())
	assert.Equal(t, authorID, added.GetNote().GetAuthorId())
	assert.Equal(t, "Forwarded to the maintenance councillor.", added.GetNote().GetBody())
	assert.Equal(t, t0, added.GetNote().GetCreatedAt().AsTime())

	list, err := h.ListNotes(as("staff"), &internalnotesv1.ListNotesRequest{ReportId: reportA})
	require.NoError(t, err)
	require.Len(t, list.GetNotes(), 1)
	assert.Equal(t, added.GetNote().GetBody(), list.GetNotes()[0].GetBody())
}

func TestPermissionDenied(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "student", ctx: as("STUDENT")},
		{name: "president", ctx: as("PRESIDENT")},
		{name: "no role metadata", ctx: context.Background()},
		{name: "empty role", ctx: as("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()

			_, err := h.AddNote(tt.ctx, addReq("Any update?"))
			assert.Equal(t, codes.PermissionDenied, status.Code(err))

			_, err = h.ListNotes(tt.ctx, &internalnotesv1.ListNotesRequest{ReportId: reportA})
			assert.Equal(t, codes.PermissionDenied, status.Code(err))
		})
	}
}

func TestSystemAdminReadsButCannotAdd(t *testing.T) {
	h := newTestHandler()
	_, err := h.AddNote(as("STAFF"), addReq("Plumber booked for Monday."))
	require.NoError(t, err)

	_, err = h.AddNote(as("SYSTEM_ADMIN"), addReq("Admin note"))
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "only Category Head and Staff add notes")

	list, err := h.ListNotes(as("SYSTEM_ADMIN"), &internalnotesv1.ListNotesRequest{ReportId: reportA})
	require.NoError(t, err)
	assert.Len(t, list.GetNotes(), 1)
}

func TestInvalidArgument(t *testing.T) {
	h := newTestHandler()

	_, err := h.AddNote(as("STAFF"), addReq("   "))
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "blank body")

	_, err = h.AddNote(as("STAFF"), addReq(strings.Repeat("x", domain.MaxBodyChars+1)))
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "1001 characters")

	_, err = h.ListNotes(as("STAFF"), &internalnotesv1.ListNotesRequest{ReportId: "report-1"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "malformed report id")
}

// brokenApp fails every call, to check internal errors are hidden.
type brokenApp struct{}

func (brokenApp) AddNote(context.Context, domain.NewNoteInput) (domain.Note, error) {
	return domain.Note{}, errors.New("save note: mongodb password is wrong")
}

func (brokenApp) ListNotes(context.Context, domain.Role, string) ([]domain.Note, error) {
	return nil, errors.New("list notes: mongodb password is wrong")
}

func TestInternalErrorsAreHidden(t *testing.T) {
	h := NewHandler(brokenApp{})

	_, err := h.AddNote(as("STAFF"), addReq("Hello"))
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")

	_, err = h.ListNotes(as("STAFF"), &internalnotesv1.ListNotesRequest{ReportId: reportA})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")
}
