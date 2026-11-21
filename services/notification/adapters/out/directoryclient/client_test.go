package directoryclient

// SCRUM-162: notification names the person a report reached by their post,
// from department-directory, for example "Block Councillor (HC)".

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
)

type fakeDirectory struct {
	departmentdirectoryv1.DepartmentDirectoryServiceClient
	heldOnly bool
	err      error
}

func (f *fakeDirectory) ListPositions(_ context.Context, in *departmentdirectoryv1.ListPositionsRequest, _ ...grpc.CallOption) (*departmentdirectoryv1.ListPositionsResponse, error) {
	f.heldOnly = in.GetHeldOnly()
	if f.err != nil {
		return nil, f.err
	}
	return &departmentdirectoryv1.ListPositionsResponse{Positions: []*departmentdirectoryv1.Position{
		{Code: "SSO", Title: "Student Service Officer", HolderUserId: "u-sso", Active: true},
		{Code: "BLOCK_COUNCILLOR_HC", Title: "Block Councillor (HC)", HolderUserId: "u-hc", Active: true},
	}}, nil
}

func TestTitleOf(t *testing.T) {
	fake := &fakeDirectory{}
	c := New(fake)
	got, err := c.TitleOf(context.Background(), "u-hc")
	require.NoError(t, err)
	assert.Equal(t, "Block Councillor (HC)", got)
	assert.True(t, fake.heldOnly, "only positions someone holds")

	got, err = c.TitleOf(context.Background(), "u-nobody")
	require.NoError(t, err)
	assert.Empty(t, got, "a user who holds no position")
}

func TestTitleOfDirectoryDown(t *testing.T) {
	_, err := New(&fakeDirectory{err: errors.New("unavailable")}).TitleOf(context.Background(), "u-hc")
	assert.Error(t, err)
}
