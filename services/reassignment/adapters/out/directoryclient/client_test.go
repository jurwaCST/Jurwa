package directoryclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
)

// SCRUM-158: reassignment asks department-directory who holds the DSA or DAA
// position when a late report goes up to them.

type fakeDirectoryClient struct {
	departmentdirectoryv1.DepartmentDirectoryServiceClient
	got  string
	resp *departmentdirectoryv1.GetStaffResponse
	err  error
}

func (f *fakeDirectoryClient) GetStaff(_ context.Context, in *departmentdirectoryv1.GetStaffRequest, _ ...grpc.CallOption) (*departmentdirectoryv1.GetStaffResponse, error) {
	f.got = in.GetPositionCode()
	return f.resp, f.err
}

func TestGetHolder(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDirectoryClient{resp: &departmentdirectoryv1.GetStaffResponse{
		Position: &departmentdirectoryv1.Position{Code: "DSA", HolderUserId: "user-dsa", Active: true},
	}}
	got, err := New(fake).GetHolder(ctx, "DSA")
	require.NoError(t, err)
	assert.Equal(t, "user-dsa", got)
	assert.Equal(t, "DSA", fake.got)

	inactive := &fakeDirectoryClient{resp: &departmentdirectoryv1.GetStaffResponse{
		Position: &departmentdirectoryv1.Position{Code: "DSA", HolderUserId: "user-dsa", Active: false},
	}}
	got, err = New(inactive).GetHolder(ctx, "DSA")
	require.NoError(t, err)
	assert.Empty(t, got, "an inactive position has no holder")
}

// A position the directory does not know has no holder, so the report goes back
// to its routed head. A directory that is down is an error, so the event is
// tried again later.
func TestGetHolderErrors(t *testing.T) {
	for _, code := range []codes.Code{codes.NotFound, codes.InvalidArgument} {
		got, err := New(&fakeDirectoryClient{err: status.Error(code, "x")}).GetHolder(context.Background(), "DSA")
		require.NoError(t, err, code.String())
		assert.Empty(t, got)
	}
	for _, code := range []codes.Code{codes.Unavailable, codes.DeadlineExceeded, codes.Internal} {
		_, err := New(&fakeDirectoryClient{err: status.Error(code, "x")}).GetHolder(context.Background(), "DSA")
		assert.Error(t, err, code.String())
	}
}
