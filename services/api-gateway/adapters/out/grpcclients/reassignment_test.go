package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
)

// fakeReassignmentClient records the request and the outgoing x-user-role.
type fakeReassignmentClient struct {
	reassignmentv1.ReassignmentServiceClient
	role []string
	got  *reassignmentv1.ReassignReportRequest
}

func (f *fakeReassignmentClient) ReassignReport(ctx context.Context, in *reassignmentv1.ReassignReportRequest, _ ...grpc.CallOption) (*reassignmentv1.ReassignReportResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	f.role, f.got = md.Get("x-user-role"), in
	return &reassignmentv1.ReassignReportResponse{Assignment: &reassignmentv1.Assignment{AssigneeUserId: in.GetToUserId()}}, nil
}

// reassignment reads the role from metadata: its request has no role field.
func TestReassignmentSendsRoleAsMetadata(t *testing.T) {
	fake := &fakeReassignmentClient{}
	resp, err := Reassignment{C: fake}.ReassignReport(context.Background(), "CATEGORY_HEAD",
		&reassignmentv1.ReassignReportRequest{ReportId: "r1", ToUserId: "u2", Reason: "Plumbing", ActorId: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "u2", resp.GetAssignment().GetAssigneeUserId())
	assert.Equal(t, []string{"CATEGORY_HEAD"}, fake.role)
	assert.Equal(t, "u1", fake.got.GetActorId())
}

type fakeDirectoryClient struct {
	departmentdirectoryv1.DepartmentDirectoryServiceClient
	got *departmentdirectoryv1.ListPositionsRequest
}

func (f *fakeDirectoryClient) ListPositions(_ context.Context, in *departmentdirectoryv1.ListPositionsRequest, _ ...grpc.CallOption) (*departmentdirectoryv1.ListPositionsResponse, error) {
	f.got = in
	return &departmentdirectoryv1.ListPositionsResponse{Positions: []*departmentdirectoryv1.Position{{Code: "DSA"}}}, nil
}

func TestDirectoryListsHeldPositions(t *testing.T) {
	fake := &fakeDirectoryClient{}
	resp, err := Directory{C: fake}.ListHeldPositions(context.Background())
	require.NoError(t, err)
	assert.True(t, fake.got.GetHeldOnly(), "only positions someone holds can be picked")
	assert.Equal(t, "DSA", resp.GetPositions()[0].GetCode())
}
