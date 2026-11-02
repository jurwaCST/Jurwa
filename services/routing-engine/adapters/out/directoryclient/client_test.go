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
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

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

func TestGetHolderErrors(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want error
	}{
		{name: "unknown position", code: codes.NotFound, want: domain.ErrNoHolder},
		{name: "empty position", code: codes.InvalidArgument, want: domain.ErrNoHolder},
		{name: "down", code: codes.Unavailable, want: domain.ErrUnavailable},
		{name: "too slow", code: codes.DeadlineExceeded, want: domain.ErrUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&fakeDirectoryClient{err: status.Error(tt.code, "x")}).GetHolder(context.Background(), "DSA")
			assert.ErrorIs(t, err, tt.want)
		})
	}
	_, err := New(&fakeDirectoryClient{err: status.Error(codes.Internal, "boom")}).GetHolder(context.Background(), "DSA")
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrUnavailable)
}
