package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	"github.com/jurwaCST/Jurwa/services/analytics/application"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// Handler implements AnalyticsService from proto/analytics/v1.
type Handler struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) GetSummary(ctx context.Context, req *analyticsv1.GetSummaryRequest) (*analyticsv1.GetSummaryResponse, error) {
	s, err := h.app.Summary(ctx, toPeriod(req.GetPeriod()))
	if err != nil {
		return nil, toStatusError(err)
	}
	return &analyticsv1.GetSummaryResponse{
		Counts:                 toProtoCounts(s.Counts),
		AverageResolutionHours: s.AverageResolutionHours,
	}, nil
}

func (h *Handler) GetCategoryBreakdown(ctx context.Context, req *analyticsv1.GetCategoryBreakdownRequest) (*analyticsv1.GetCategoryBreakdownResponse, error) {
	rows, err := h.app.Breakdown(ctx, toPeriod(req.GetPeriod()))
	if err != nil {
		return nil, toStatusError(err)
	}
	out := make([]*analyticsv1.CategoryCounts, len(rows))
	for i, r := range rows {
		// domain.Category uses the same numbers as commonv1.Category.
		out[i] = &analyticsv1.CategoryCounts{Category: commonv1.Category(r.Category), Counts: toProtoCounts(r.Counts)}
	}
	return &analyticsv1.GetCategoryBreakdownResponse{Categories: out}, nil
}

// toPeriod converts the optional proto period. An unset end means no limit.
func toPeriod(p *analyticsv1.Period) domain.Period {
	var out domain.Period
	if p.GetFrom() != nil {
		from := p.GetFrom().AsTime()
		out.From = &from
	}
	if p.GetTo() != nil {
		to := p.GetTo().AsTime()
		out.To = &to
	}
	return out
}

func toProtoCounts(c domain.Counts) *analyticsv1.StatusCounts {
	return &analyticsv1.StatusCounts{
		Total:      c.Total,
		Open:       c.Open,
		Assigned:   c.Assigned,
		InProgress: c.InProgress,
		Resolved:   c.Resolved,
		Escalated:  c.Escalated,
	}
}

func toStatusError(err error) error {
	if errors.Is(err, domain.ErrInvalidPeriod) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, "internal error")
}
