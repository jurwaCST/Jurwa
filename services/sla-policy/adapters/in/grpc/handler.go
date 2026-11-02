package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slapolicyv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-policy/v1"
	"github.com/jurwaCST/Jurwa/services/sla-policy/application"
	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
)

// Handler implements SlaPolicyService from proto/sla-policy/v1.
type Handler struct {
	slapolicyv1.UnimplementedSlaPolicyServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) GetPolicy(ctx context.Context, req *slapolicyv1.GetPolicyRequest) (*slapolicyv1.GetPolicyResponse, error) {
	p, err := h.app.GetPolicy(ctx, domain.Category(req.GetCategory()))
	if err != nil {
		return nil, toStatusError(err)
	}
	return &slapolicyv1.GetPolicyResponse{Policy: toProto(p)}, nil
}

func (h *Handler) UpsertPolicy(ctx context.Context, req *slapolicyv1.UpsertPolicyRequest) (*slapolicyv1.UpsertPolicyResponse, error) {
	p, err := h.app.UpsertPolicy(ctx, domain.Category(req.GetCategory()), req.GetWindowHours(), req.GetUpdatedBy())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &slapolicyv1.UpsertPolicyResponse{Policy: toProto(p)}, nil
}

// toProto converts the domain policy. domain.Category uses the same numbers as commonv1.Category.
func toProto(p domain.Policy) *slapolicyv1.Policy {
	out := &slapolicyv1.Policy{
		Category:    commonv1.Category(p.Category),
		WindowHours: p.WindowHours,
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
	if p.UpdatedBy != "" {
		updatedBy := p.UpdatedBy
		out.UpdatedBy = &updatedBy
	}
	return out
}

// toStatusError maps domain errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidArgument),
		errors.Is(err, domain.ErrInvalidCategory),
		errors.Is(err, domain.ErrInvalidWindow):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
