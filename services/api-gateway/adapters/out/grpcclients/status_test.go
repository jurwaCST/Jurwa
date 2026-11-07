package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

type fakeStatusClient struct {
	statuslifecyclev1.StatusLifecycleServiceClient
	changed *statuslifecyclev1.ChangeStatusRequest
	asked   *statuslifecyclev1.GetStatusRequest
}

func (f *fakeStatusClient) ChangeStatus(_ context.Context, in *statuslifecyclev1.ChangeStatusRequest, _ ...grpc.CallOption) (*statuslifecyclev1.ChangeStatusResponse, error) {
	f.changed = in
	return &statuslifecyclev1.ChangeStatusResponse{}, nil
}

func (f *fakeStatusClient) GetStatus(_ context.Context, in *statuslifecyclev1.GetStatusRequest, _ ...grpc.CallOption) (*statuslifecyclev1.GetStatusResponse, error) {
	f.asked = in
	return &statuslifecyclev1.GetStatusResponse{}, nil
}

func TestStatus(t *testing.T) {
	ctx := context.Background()
	fc := &fakeStatusClient{}
	s := Status{C: fc}

	in := &statuslifecyclev1.ChangeStatusRequest{ReportId: "r1", ToStatus: commonv1.Status_IN_PROGRESS, ActorId: "u1"}
	_, err := s.ChangeStatus(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, in, fc.changed)

	_, err = s.GetStatus(ctx, "r2")
	require.NoError(t, err)
	assert.Equal(t, "r2", fc.asked.GetReportId())
}
