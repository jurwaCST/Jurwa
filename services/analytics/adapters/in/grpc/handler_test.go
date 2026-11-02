package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/analytics/application"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func newHandlerWithData(t *testing.T) *Handler {
	t.Helper()
	app := application.NewService(memory.New())
	ctx := context.Background()
	for _, e := range []domain.Event{
		{ID: "e1", Type: domain.ReportSubmitted, ReportID: "r1", Category: domain.Hostel, OccurredAt: t0},
		{ID: "e2", Type: domain.ReportSubmitted, ReportID: "r2", Category: domain.Others, OccurredAt: t0},
		{ID: "e3", Type: domain.StatusChanged, ReportID: "r2", ToStatus: domain.Escalated, OccurredAt: t0.Add(48 * time.Hour)},
		{ID: "e4", Type: domain.StatusChanged, ReportID: "r2", ToStatus: domain.Resolved, OccurredAt: t0.Add(50 * time.Hour)},
	} {
		_, err := app.Record(ctx, e)
		require.NoError(t, err)
	}
	return NewHandler(app)
}

func TestGetSummary(t *testing.T) {
	res, err := newHandlerWithData(t).GetSummary(context.Background(), &analyticsv1.GetSummaryRequest{})
	require.NoError(t, err)
	c := res.GetCounts()
	assert.Equal(t, int64(2), c.GetTotal())
	assert.Equal(t, int64(1), c.GetOpen())
	assert.Equal(t, int64(1), c.GetResolved())
	assert.Equal(t, int64(1), c.GetEscalated())
	assert.Equal(t, 50.0, res.GetAverageResolutionHours())
}

func TestGetSummaryWithPeriod(t *testing.T) {
	h := newHandlerWithData(t)
	from := timestamppb.New(t0.Add(time.Hour))
	res, err := h.GetSummary(context.Background(), &analyticsv1.GetSummaryRequest{
		Period: &analyticsv1.Period{From: from},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), res.GetCounts().GetTotal(), "both reports were created before from")

	to := timestamppb.New(t0.Add(time.Hour))
	res, err = h.GetSummary(context.Background(), &analyticsv1.GetSummaryRequest{
		Period: &analyticsv1.Period{To: to},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.GetCounts().GetTotal())
}

func TestGetCategoryBreakdown(t *testing.T) {
	res, err := newHandlerWithData(t).GetCategoryBreakdown(context.Background(), &analyticsv1.GetCategoryBreakdownRequest{})
	require.NoError(t, err)
	cats := res.GetCategories()
	require.Len(t, cats, 3)
	assert.Equal(t, commonv1.Category_HOSTEL, cats[0].GetCategory())
	assert.Equal(t, int64(1), cats[0].GetCounts().GetOpen())
	assert.Equal(t, commonv1.Category_ACADEMIC_BLOCK, cats[1].GetCategory())
	assert.Equal(t, int64(0), cats[1].GetCounts().GetTotal())
	assert.Equal(t, commonv1.Category_OTHERS, cats[2].GetCategory())
	assert.Equal(t, int64(1), cats[2].GetCounts().GetEscalated())
}

func TestBadPeriodIsInvalidArgument(t *testing.T) {
	h := newHandlerWithData(t)
	bad := &analyticsv1.Period{From: timestamppb.New(t0.Add(time.Hour)), To: timestamppb.New(t0)}

	_, err := h.GetSummary(context.Background(), &analyticsv1.GetSummaryRequest{Period: bad})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = h.GetCategoryBreakdown(context.Background(), &analyticsv1.GetCategoryBreakdownRequest{Period: bad})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestToStatusErrorInternal(t *testing.T) {
	err := toStatusError(errors.New("db down"))
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "db down")
}
