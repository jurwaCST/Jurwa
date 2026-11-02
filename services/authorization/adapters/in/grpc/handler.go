package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authorizationv1 "github.com/jurwaCST/Jurwa/proto/gen/authorization/v1"
	"github.com/jurwaCST/Jurwa/services/authorization/application"
	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

// Handler implements AuthorizationService from proto/authorization/v1.
type Handler struct {
	authorizationv1.UnimplementedAuthorizationServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

// CheckPermission answers allowed or denied. A denial is a normal response
// (allowed=false with a reason), not a gRPC error.
func (h *Handler) CheckPermission(ctx context.Context, req *authorizationv1.CheckPermissionRequest) (*authorizationv1.CheckPermissionResponse, error) {
	// domain.Role uses the same numbers as commonv1.Role.
	d, err := h.app.CheckPermission(ctx, domain.Role(req.GetRole()), req.GetAction(), req.GetResource())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &authorizationv1.CheckPermissionResponse{Allowed: d.Allowed, Reason: d.Reason}, nil
}

// toStatusError maps domain errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, domain.ErrRoleRequired),
		errors.Is(err, domain.ErrActionRequired),
		errors.Is(err, domain.ErrResourceRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
