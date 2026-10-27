package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	"github.com/jurwaCST/Jurwa/services/routing-rules/application"
	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

// Handler implements routingrulesv1.RoutingRulesServiceServer.
type Handler struct {
	routingrulesv1.UnimplementedRoutingRulesServiceServer
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetRule(ctx context.Context, req *routingrulesv1.GetRuleRequest) (*routingrulesv1.GetRuleResponse, error) {
	q := domain.Query{
		Category: toDomainCategory(req.GetCategory()),
		Building: req.GetLocation().GetBuilding(),
	}
	rule, err := h.svc.GetRule(ctx, q)
	if err != nil {
		return nil, toStatus(err)
	}
	return &routingrulesv1.GetRuleResponse{Rule: toProtoRule(rule)}, nil
}

func (h *Handler) ListRules(ctx context.Context, req *routingrulesv1.ListRulesRequest) (*routingrulesv1.ListRulesResponse, error) {
	rules, err := h.svc.ListRules(ctx, req.GetIncludeInactive())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &routingrulesv1.ListRulesResponse{}
	for _, r := range rules {
		resp.Rules = append(resp.Rules, toProtoRule(r))
	}
	return resp, nil
}

func (h *Handler) UpsertRule(ctx context.Context, req *routingrulesv1.UpsertRuleRequest) (*routingrulesv1.UpsertRuleResponse, error) {
	in := req.GetRule()
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "rule is required")
	}
	rule, err := h.svc.UpsertRule(ctx, domain.Rule{
		Category:       toDomainCategory(in.GetCategory()),
		Building:       in.GetBuilding(),
		TargetPosition: in.GetTargetPosition(),
		Priority:       int(in.GetPriority()),
		Active:         in.GetActive(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &routingrulesv1.UpsertRuleResponse{Rule: toProtoRule(rule)}, nil
}

// toDomainCategory maps the proto enum to the domain code. CATEGORY_UNSPECIFIED
// and unknown numbers become "", which the domain rejects.
func toDomainCategory(c commonv1.Category) domain.Category {
	if c == commonv1.Category_CATEGORY_UNSPECIFIED {
		return ""
	}
	return domain.Category(commonv1.Category_name[int32(c)])
}

func toProtoCategory(c domain.Category) commonv1.Category {
	return commonv1.Category(commonv1.Category_value[string(c)])
}

func toProtoRule(r domain.Rule) *routingrulesv1.Rule {
	out := &routingrulesv1.Rule{
		Id:             r.ID,
		Category:       toProtoCategory(r.Category),
		Building:       r.Building,
		TargetPosition: r.TargetPosition,
		Priority:       int32(r.Priority),
		Active:         r.Active,
	}
	if !r.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(r.UpdatedAt)
	}
	return out
}

// toStatus maps domain errors to gRPC status codes, as written in the contract.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCategory), errors.Is(err, domain.ErrInvalidRule):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrRuleNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		slog.Error("routing-rules request failed", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}
