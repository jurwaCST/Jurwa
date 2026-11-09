package grpcclients

import (
	"context"

	"google.golang.org/grpc/metadata"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
)

// Reassignment implements ports.Reassignment.
type Reassignment struct {
	C reassignmentv1.ReassignmentServiceClient
}

func (r Reassignment) ReassignReport(ctx context.Context, role string, in *reassignmentv1.ReassignReportRequest) (*reassignmentv1.ReassignReportResponse, error) {
	return r.C.ReassignReport(metadata.AppendToOutgoingContext(ctx, RoleMetadataKey, role), in)
}

// Directory implements ports.Directory.
type Directory struct {
	C departmentdirectoryv1.DepartmentDirectoryServiceClient
}

func (d Directory) ListHeldPositions(ctx context.Context) (*departmentdirectoryv1.ListPositionsResponse, error) {
	return d.C.ListPositions(ctx, &departmentdirectoryv1.ListPositionsRequest{HeldOnly: true})
}
