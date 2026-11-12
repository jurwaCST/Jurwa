package reportclient

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

const (
	reportID = "33333333-3333-4333-8333-333333333333"
	student  = "22222222-2222-4222-8222-222222222222"
)

// fakeClient answers GetReport with a fixed report or error. The embedded
// interface covers the methods notification never calls.
type fakeClient struct {
	reportsubmissionv1.ReportSubmissionServiceClient
	got *reportsubmissionv1.GetReportRequest
	err error
}

func (f *fakeClient) GetReport(_ context.Context, in *reportsubmissionv1.GetReportRequest, _ ...grpc.CallOption) (*reportsubmissionv1.GetReportResponse, error) {
	f.got = in
	if f.err != nil {
		return nil, f.err
	}
	return &reportsubmissionv1.GetReportResponse{Report: &commonv1.Report{
		Id: in.GetReportId(), ReporterId: student,
		Location: &commonv1.Location{Building: "HC", Wing: "Floor 1", Area: "Left wing bathroom"},
	}}, nil
}

func TestGetReportReadsTheStudentAndPlace(t *testing.T) {
	fake := &fakeClient{}
	got, err := New(fake).GetReport(context.Background(), reportID)
	require.NoError(t, err)
	assert.Equal(t, reportID, fake.got.GetReportId())
	assert.Equal(t, ports.Report{ID: reportID, ReporterID: student, Building: "HC", Wing: "Floor 1", Area: "Left wing bathroom"}, got)
}

func TestGetReportErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"not found", status.Error(codes.NotFound, "report not found"), ports.ErrReportNotFound},
		{"bad id", status.Error(codes.InvalidArgument, "report_id must be a UUID"), ports.ErrReportNotFound},
		{"down", status.Error(codes.Unavailable, "connection refused"), ports.ErrReportsUnavailable},
		{"slow", status.Error(codes.DeadlineExceeded, "deadline"), ports.ErrReportsUnavailable},
		{"other", errors.New("boom"), ports.ErrReportsUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&fakeClient{err: tt.err}).GetReport(context.Background(), reportID)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}
