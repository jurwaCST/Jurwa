package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
)

type fakeAnalyticsClient struct {
	analyticsv1.AnalyticsServiceClient
	summary   *analyticsv1.GetSummaryRequest
	breakdown *analyticsv1.GetCategoryBreakdownRequest
}

func (f *fakeAnalyticsClient) GetSummary(_ context.Context, in *analyticsv1.GetSummaryRequest, _ ...grpc.CallOption) (*analyticsv1.GetSummaryResponse, error) {
	f.summary = in
	return &analyticsv1.GetSummaryResponse{}, nil
}

func (f *fakeAnalyticsClient) GetCategoryBreakdown(_ context.Context, in *analyticsv1.GetCategoryBreakdownRequest, _ ...grpc.CallOption) (*analyticsv1.GetCategoryBreakdownResponse, error) {
	f.breakdown = in
	return &analyticsv1.GetCategoryBreakdownResponse{}, nil
}

func TestAnalytics(t *testing.T) {
	ctx := context.Background()
	fc := &fakeAnalyticsClient{}
	a := Analytics{C: fc}
	p := &analyticsv1.Period{}

	_, err := a.GetSummary(ctx, p)
	require.NoError(t, err)
	assert.Same(t, p, fc.summary.GetPeriod())

	_, err = a.GetCategoryBreakdown(ctx, nil)
	require.NoError(t, err)
	assert.Nil(t, fc.breakdown.GetPeriod())
}
