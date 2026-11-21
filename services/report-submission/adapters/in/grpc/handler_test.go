package grpc

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/uuid"
	"github.com/jurwaCST/Jurwa/services/report-submission/application"
)

const reporter = "0b9e8d7c-6a5b-4c3d-9e2f-1a0b9c8d7e6f"

func newTestHandler() *Handler {
	app := application.NewService(memory.NewRepository(), uuid.Generator{}, logpublisher.New(slog.New(slog.DiscardHandler)), time.Now)
	return NewHandler(app)
}

func validRequest() *reportsubmissionv1.SubmitReportRequest {
	return &reportsubmissionv1.SubmitReportRequest{
		ReporterId:  reporter,
		Category:    commonv1.Category_HOSTEL,
		Location:    &commonv1.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"},
		Description: "Tap in the bathroom is leaking.",
	}
}

func TestSubmitReportReturnsOpenReport(t *testing.T) {
	h := newTestHandler()
	req := validRequest()
	req.PhotoId = proto.String("9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d")

	resp, err := h.SubmitReport(context.Background(), req)

	require.NoError(t, err)
	r := resp.GetReport()
	assert.Len(t, r.GetId(), 36)
	assert.Equal(t, commonv1.Status_OPEN, r.GetStatus())
	assert.Equal(t, commonv1.Category_HOSTEL, r.GetCategory())
	assert.Equal(t, "Hostel B", r.GetLocation().GetBuilding())
	assert.Equal(t, req.GetPhotoId(), r.GetPhotoId())
	assert.NotNil(t, r.GetCreatedAt())
}

func TestSubmitReportInvalidArgument(t *testing.T) {
	tests := []struct {
		name   string
		change func(r *reportsubmissionv1.SubmitReportRequest)
	}{
		{name: "empty description", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.Description = "" }},
		{name: "501 character description", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.Description = strings.Repeat("a", 501) }},
		{name: "unspecified category", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.Category = commonv1.Category_CATEGORY_UNSPECIFIED }},
		{name: "unknown category number", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.Category = commonv1.Category(42) }},
		{name: "no location", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.Location = nil }},
		{name: "no reporter", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.ReporterId = "" }},
		{name: "blank photo id", change: func(r *reportsubmissionv1.SubmitReportRequest) { r.PhotoId = proto.String("") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRequest()
			tt.change(req)

			_, err := newTestHandler().SubmitReport(context.Background(), req)

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestGetReport(t *testing.T) {
	h := newTestHandler()
	created, err := h.SubmitReport(context.Background(), validRequest())
	require.NoError(t, err)

	got, err := h.GetReport(context.Background(), &reportsubmissionv1.GetReportRequest{ReportId: created.GetReport().GetId()})
	require.NoError(t, err)
	assert.True(t, proto.Equal(created.GetReport(), got.GetReport()))

	_, err = h.GetReport(context.Background(), &reportsubmissionv1.GetReportRequest{ReportId: "11111111-1111-4111-8111-111111111111"})
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = h.GetReport(context.Background(), &reportsubmissionv1.GetReportRequest{ReportId: "not-a-uuid"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCategoryRoundTrip(t *testing.T) {
	for _, c := range []commonv1.Category{commonv1.Category_HOSTEL, commonv1.Category_ACADEMIC_BLOCK, commonv1.Category_OTHERS} {
		assert.Equal(t, c, categoryToProto(categoryFromProto(c)))
	}
	assert.Equal(t, commonv1.Category_CATEGORY_UNSPECIFIED, categoryToProto("UNKNOWN_42"))
	assert.Equal(t, commonv1.Status_STATUS_UNSPECIFIED, statusToProto("CLOSED"))
}

func TestToStatusHidesInternalErrors(t *testing.T) {
	err := toStatus(assert.AnError)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), assert.AnError.Error())
}

// SCRUM-163: UpdateDescription and WithdrawReport, with their error codes.
func TestUpdateAndWithdraw(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	sent, err := h.SubmitReport(ctx, validRequest())
	require.NoError(t, err)
	id := sent.GetReport().GetId()

	up, err := h.UpdateDescription(ctx, &reportsubmissionv1.UpdateDescriptionRequest{ReportId: id, ReporterId: reporter, Description: "Now also the light"})
	require.NoError(t, err)
	assert.Equal(t, "Now also the light", up.GetReport().GetDescription())

	_, err = h.UpdateDescription(ctx, &reportsubmissionv1.UpdateDescriptionRequest{ReportId: id, ReporterId: "11111111-1111-4111-8111-111111111111", Description: "x"})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
	_, err = h.UpdateDescription(ctx, &reportsubmissionv1.UpdateDescriptionRequest{ReportId: id, ReporterId: reporter, Description: ""})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.WithdrawReport(ctx, &reportsubmissionv1.WithdrawReportRequest{ReportId: id, ReporterId: reporter, CurrentAssigneeId: "user-sso"})
	require.NoError(t, err)
	_, err = h.GetReport(ctx, &reportsubmissionv1.GetReportRequest{ReportId: id})
	assert.Equal(t, codes.NotFound, status.Code(err))
	_, err = h.WithdrawReport(ctx, &reportsubmissionv1.WithdrawReportRequest{ReportId: id, ReporterId: reporter})
	assert.Equal(t, codes.NotFound, status.Code(err))
}
