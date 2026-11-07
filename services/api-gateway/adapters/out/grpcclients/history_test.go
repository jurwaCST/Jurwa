package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
)

type fakeHistoryClient struct {
	list     *reporthistoryv1.ListReportsByReporterRequest
	assigned *reporthistoryv1.ListReportsByAssigneeRequest
	tl       *reporthistoryv1.GetReportTimelineRequest
}

func (f *fakeHistoryClient) ListReportsByReporter(_ context.Context, in *reporthistoryv1.ListReportsByReporterRequest, _ ...grpc.CallOption) (*reporthistoryv1.ListReportsByReporterResponse, error) {
	f.list = in
	return &reporthistoryv1.ListReportsByReporterResponse{NextPageToken: "n"}, nil
}

func (f *fakeHistoryClient) ListReportsByAssignee(_ context.Context, in *reporthistoryv1.ListReportsByAssigneeRequest, _ ...grpc.CallOption) (*reporthistoryv1.ListReportsByAssigneeResponse, error) {
	f.assigned = in
	return &reporthistoryv1.ListReportsByAssigneeResponse{NextPageToken: "a"}, nil
}

func (f *fakeHistoryClient) GetReportTimeline(_ context.Context, in *reporthistoryv1.GetReportTimelineRequest, _ ...grpc.CallOption) (*reporthistoryv1.GetReportTimelineResponse, error) {
	f.tl = in
	return &reporthistoryv1.GetReportTimelineResponse{}, nil
}

func TestHistory(t *testing.T) {
	ctx := context.Background()
	fake := &fakeHistoryClient{}
	h := History{C: fake}

	resp, err := h.ListReportsByReporter(ctx, "u1", 5, "tok")
	require.NoError(t, err)
	assert.Equal(t, "n", resp.GetNextPageToken())
	assert.Equal(t, "u1", fake.list.GetReporterId())
	assert.Equal(t, int32(5), fake.list.GetPageSize())
	assert.Equal(t, "tok", fake.list.GetPageToken())

	_, err = h.GetReportTimeline(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", fake.tl.GetReportId())
}

func TestHistoryListReportsByAssignee(t *testing.T) {
	fc := &fakeHistoryClient{}
	resp, err := History{C: fc}.ListReportsByAssignee(context.Background(), "u-head", 7, "tok")
	require.NoError(t, err)
	assert.Equal(t, "a", resp.GetNextPageToken())
	assert.Equal(t, "u-head", fc.assigned.GetAssigneeId())
	assert.Equal(t, int32(7), fc.assigned.GetPageSize())
	assert.Equal(t, "tok", fc.assigned.GetPageToken())
}
