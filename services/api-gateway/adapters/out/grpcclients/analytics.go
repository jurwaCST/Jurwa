package grpcclients

import (
	"context"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
)

// Analytics implements ports.Analytics.
type Analytics struct {
	C analyticsv1.AnalyticsServiceClient
}

func (a Analytics) GetSummary(ctx context.Context, period *analyticsv1.Period) (*analyticsv1.GetSummaryResponse, error) {
	return a.C.GetSummary(ctx, &analyticsv1.GetSummaryRequest{Period: period})
}

func (a Analytics) GetCategoryBreakdown(ctx context.Context, period *analyticsv1.Period) (*analyticsv1.GetCategoryBreakdownResponse, error) {
	return a.C.GetCategoryBreakdown(ctx, &analyticsv1.GetCategoryBreakdownRequest{Period: period})
}
