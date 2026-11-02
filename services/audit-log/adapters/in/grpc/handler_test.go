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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	auditlogv1 "github.com/jurwaCST/Jurwa/proto/gen/audit-log/v1"
	"github.com/jurwaCST/Jurwa/services/audit-log/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/audit-log/application"
)

const (
	reportID = "11111111-1111-1111-1111-111111111111"
	actorID  = "22222222-2222-2222-2222-222222222222"
)

func newHandler() *Handler {
	return NewHandler(application.NewService(memory.New(), time.Now, uuid.NewString))
}

func as(role string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(RoleMetadataKey, role))
}

func TestAppendThenList(t *testing.T) {
	h := newHandler()
	at := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

	resp, err := h.Append(context.Background(), &auditlogv1.AppendRequest{Event: &auditlogv1.AuditEvent{
		Id:         "ignored",
		OccurredAt: timestamppb.New(at),
		Service:    "status-lifecycle",
		Action:     "STATUS_CHANGED",
		ReportId:   proto.String(reportID),
		ActorId:    proto.String(actorID),
		OnBehalfOf: proto.String("Cook, Hostel A kitchen"),
		Details:    map[string]string{"from": "OPEN", "to": "ASSIGNED"},
	}})
	require.NoError(t, err)
	assert.True(t, uuid.Validate(resp.GetId()) == nil, "the server sets the id")
	assert.NotEqual(t, "ignored", resp.GetId())

	_, err = h.Append(context.Background(), &auditlogv1.AppendRequest{Event: &auditlogv1.AuditEvent{
		Service: "sla-escalation-checker", Action: "ESCALATED", ReportId: proto.String(reportID),
	}})
	require.NoError(t, err)

	list, err := h.ListByReport(as("system_admin"), &auditlogv1.ListByReportRequest{ReportId: reportID})
	require.NoError(t, err)
	require.Len(t, list.GetEvents(), 2)
	first, second := list.GetEvents()[0], list.GetEvents()[1]
	assert.Equal(t, resp.GetId(), first.GetId(), "oldest first")
	assert.Equal(t, at, first.GetOccurredAt().AsTime())
	assert.Equal(t, "Cook, Hostel A kitchen", first.GetOnBehalfOf())
	assert.Equal(t, "ASSIGNED", first.GetDetails()["to"])
	assert.Nil(t, second.ActorId, "the system acted")
	assert.Nil(t, second.OnBehalfOf)
}

func TestErrorCodes(t *testing.T) {
	h := newHandler()
	ctx := context.Background()

	_, err := h.Append(ctx, &auditlogv1.AppendRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "event missing")

	_, err = h.Append(ctx, &auditlogv1.AppendRequest{Event: &auditlogv1.AuditEvent{Action: "ROUTED"}})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "service missing")

	_, err = h.Append(ctx, &auditlogv1.AppendRequest{Event: &auditlogv1.AuditEvent{Service: "x"}})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "action missing")

	_, err = h.ListByReport(ctx, &auditlogv1.ListByReportRequest{ReportId: reportID})
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "no role")

	_, err = h.ListByReport(as("CATEGORY_HEAD"), &auditlogv1.ListByReportRequest{ReportId: reportID})
	assert.Equal(t, codes.PermissionDenied, status.Code(err), "category head")

	_, err = h.ListByReport(as("SYSTEM_ADMIN"), &auditlogv1.ListByReportRequest{ReportId: "abc"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "bad report id")
}

func TestToStatusError_Internal(t *testing.T) {
	assert.Equal(t, codes.Internal, status.Code(toStatusError(errors.New("db down"))))
}
