package statusclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

const reportID = "33333333-3333-3333-3333-333333333333"

// fakeStatusClient answers GetStatus with a fixed status or error. The
// embedded interface covers the methods feedback never calls.
type fakeStatusClient struct {
	statuslifecyclev1.StatusLifecycleServiceClient
	got    *statuslifecyclev1.GetStatusRequest
	status commonv1.Status
	err    error
}

func (f *fakeStatusClient) GetStatus(_ context.Context, in *statuslifecyclev1.GetStatusRequest, _ ...grpc.CallOption) (*statuslifecyclev1.GetStatusResponse, error) {
	f.got = in
	if f.err != nil {
		return nil, f.err
	}
	return &statuslifecyclev1.GetStatusResponse{Status: &statuslifecyclev1.ReportStatus{ReportId: in.GetReportId(), Status: f.status}}, nil
}

func TestIsResolved(t *testing.T) {
	tests := []struct {
		status commonv1.Status
		want   bool
	}{
		{commonv1.Status_OPEN, false},
		{commonv1.Status_ASSIGNED, false},
		{commonv1.Status_IN_PROGRESS, false},
		{commonv1.Status_RESOLVED, true},
		{commonv1.Status_ESCALATED, false},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			fake := &fakeStatusClient{status: tt.status}
			got, err := New(fake).IsResolved(context.Background(), reportID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, reportID, fake.got.GetReportId())
		})
	}
}

func TestIsResolvedErrors(t *testing.T) {
	t.Run("unknown report is not resolved", func(t *testing.T) {
		got, err := New(&fakeStatusClient{err: status.Error(codes.NotFound, "no status")}).IsResolved(context.Background(), reportID)
		require.NoError(t, err)
		assert.False(t, got)
	})
	for _, code := range []codes.Code{codes.Unavailable, codes.DeadlineExceeded} {
		t.Run(code.String()+" is unavailable", func(t *testing.T) {
			_, err := New(&fakeStatusClient{err: status.Error(code, "x")}).IsResolved(context.Background(), reportID)
			assert.ErrorIs(t, err, domain.ErrUnavailable)
		})
	}
	t.Run("other errors are passed on", func(t *testing.T) {
		_, err := New(&fakeStatusClient{err: status.Error(codes.Internal, "boom")}).IsResolved(context.Background(), reportID)
		require.Error(t, err)
		assert.NotErrorIs(t, err, domain.ErrUnavailable)
	})
}
