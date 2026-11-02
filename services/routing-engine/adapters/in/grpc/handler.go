package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	routingenginev1 "github.com/jurwaCST/Jurwa/proto/gen/routing-engine/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/application"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// Handler implements routingenginev1.RoutingEngineServiceServer.
type Handler struct {
	routingenginev1.UnimplementedRoutingEngineServiceServer
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RouteReport(ctx context.Context, req *routingenginev1.RouteReportRequest) (*routingenginev1.RouteReportResponse, error) {
	d, err := h.svc.RouteReport(ctx, domain.Request{
		ReportID: req.GetReportId(),
		Category: toDomainCategory(req.GetCategory()),
		Building: req.GetLocation().GetBuilding(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &routingenginev1.RouteReportResponse{
		CategoryHeadId: d.CategoryHeadID,
		TargetPosition: d.TargetPosition,
		RuleId:         d.RuleID,
	}, nil
}

// toDomainCategory maps the proto enum to the domain code. CATEGORY_UNSPECIFIED
// and unknown numbers become codes the domain rejects.
func toDomainCategory(c commonv1.Category) domain.Category {
	if c == commonv1.Category_CATEGORY_UNSPECIFIED {
		return ""
	}
	return domain.Category(c.String())
}

// toStatus maps domain errors to gRPC status codes, as written in the contract.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCategory), errors.Is(err, domain.ErrInvalidLocation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrRuleNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrNoHolder):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	default:
		slog.Error("routing-engine request failed", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}
