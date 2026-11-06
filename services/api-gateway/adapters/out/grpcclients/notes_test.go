package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
)

// fakeNotesClient records the outgoing x-user-role metadata of each call.
type fakeNotesClient struct {
	role []string
	got  *internalnotesv1.AddNoteRequest
}

func (f *fakeNotesClient) AddNote(ctx context.Context, in *internalnotesv1.AddNoteRequest, _ ...grpc.CallOption) (*internalnotesv1.AddNoteResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	f.role, f.got = md.Get("x-user-role"), in
	return &internalnotesv1.AddNoteResponse{Note: &internalnotesv1.Note{Body: in.GetBody()}}, nil
}

func (f *fakeNotesClient) ListNotes(ctx context.Context, in *internalnotesv1.ListNotesRequest, _ ...grpc.CallOption) (*internalnotesv1.ListNotesResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	f.role = md.Get("x-user-role")
	return &internalnotesv1.ListNotesResponse{Notes: []*internalnotesv1.Note{{ReportId: in.GetReportId()}}}, nil
}

func TestNotesSendRoleAsMetadata(t *testing.T) {
	ctx := context.Background()
	fake := &fakeNotesClient{}
	n := Notes{C: fake}

	resp, err := n.AddNote(ctx, "STAFF", &internalnotesv1.AddNoteRequest{ReportId: "r1", AuthorId: "u1", Body: "hi"})
	require.NoError(t, err)
	assert.Equal(t, "hi", resp.GetNote().GetBody())
	assert.Equal(t, []string{"STAFF"}, fake.role)
	assert.Equal(t, "u1", fake.got.GetAuthorId())

	list, err := n.ListNotes(ctx, "SYSTEM_ADMIN", "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", list.GetNotes()[0].GetReportId())
	assert.Equal(t, []string{"SYSTEM_ADMIN"}, fake.role)
}
