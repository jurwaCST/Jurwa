package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/reassignment/application"
	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

const (
	reportID = "11111111-1111-1111-1111-111111111111"
	headID   = "22222222-2222-2222-2222-222222222222"
	staffID  = "33333333-3333-3333-3333-333333333333"
)

type nopPublisher struct{}

func (nopPublisher) PublishReportReassigned(context.Context, ports.ReportReassigned) error {
	return nil
}

func newHandler(t *testing.T) *Handler {
	app := application.NewService(memory.New(), nopPublisher{}, time.Now, uuid.NewString)
	_, err := app.RecordRouted(context.Background(), reportID, headID, "DSA")
	require.NoError(t, err)
	return NewHandler(app)
}

func as(role string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(RoleMetadataKey, role))
}

func TestReassignAndList(t *testing.T) {
	h := newHandler(t)

	resp, err := h.ReassignReport(as(" category_head "), &reassignmentv1.ReassignReportRequest{
		ReportId: reportID, ToUserId: staffID, Reason: "Plumbing", ActorId: headID,
	})
	require.NoError(t, err)
	a := resp.GetAssignment()
	assert.Equal(t, reassignmentv1.AssignmentKind_REASSIGNED, a.GetKind())
	assert.Equal(t, staffID, a.GetAssigneeUserId())
	assert.Equal(t, headID, a.GetAssignedBy())
	assert.True(t, a.GetCurrent())

	list, err := h.ListReassignments(as("SYSTEM_ADMIN"), &reassignmentv1.ListReassignmentsRequest{ReportId: reportID})
	require.NoError(t, err)
	require.Len(t, list.GetAssignments(), 2)
	first := list.GetAssignments()[0]
	assert.Equal(t, reassignmentv1.AssignmentKind_ROUTED, first.GetKind())
	assert.Equal(t, "DSA", first.GetAssigneePositionCode())
	assert.Nil(t, first.AssignedBy, "routed by the system")
	assert.False(t, first.GetCurrent())
}

func TestErrorCodes(t *testing.T) {
	h := newHandler(t)
	valid := func() *reassignmentv1.ReassignReportRequest {
		return &reassignmentv1.ReassignReportRequest{ReportId: reportID, ToUserId: staffID, Reason: "Plumbing", ActorId: headID}
	}

	_, err := h.ReassignReport(context.Background(), valid())
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "no role in metadata")

	_, err = h.ReassignReport(as("STUDENT"), valid())
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "student")

	req := valid()
	req.ActorId = staffID // staffID does not hold the report
	_, err = h.ReassignReport(as("STAFF"), req)
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "staff who is not the assignee")

	req = valid()
	req.Reason = ""
	_, err = h.ReassignReport(as("CATEGORY_HEAD"), req)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "blank reason")

	req = valid()
	req.ToUserId = headID
	_, err = h.ReassignReport(as("CATEGORY_HEAD"), req)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "same assignee")

	req = valid()
	req.ReportId = staffID
	_, err = h.ReassignReport(as("CATEGORY_HEAD"), req)
	assert.Equal(t, codes.NotFound, status.Code(err), "report has no assignment")

	_, err = h.ListReassignments(as("STUDENT"), &reassignmentv1.ListReassignmentsRequest{ReportId: reportID})
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "student cannot list")

	_, err = h.ListReassignments(as("STAFF"), &reassignmentv1.ListReassignmentsRequest{ReportId: "abc"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "bad report id")
}

func TestToStatusError(t *testing.T) {
	assert.Equal(t, codes.Aborted, status.Code(toStatusError(domain.ErrConflict)))
	assert.Equal(t, codes.AlreadyExists, status.Code(toStatusError(domain.ErrAlreadyAssigned)))
	assert.Equal(t, codes.Internal, status.Code(toStatusError(errors.New("db down"))))
}
