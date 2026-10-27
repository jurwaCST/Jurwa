// Package grpc is the inbound adapter: it turns IdentityService requests into
// use case calls and domain errors into gRPC status codes. No business rules.
package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	"github.com/jurwaCST/Jurwa/services/identity/application"
	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

type Handler struct {
	identityv1.UnimplementedIdentityServiceServer
	app *application.Service
}

func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	u, err := h.app.Register(ctx, req.GetEmail(), req.GetFullName(), req.GetPassword())
	if err != nil {
		return nil, toStatus(err)
	}
	slog.Info("user registered", "user_id", u.ID)
	return &identityv1.RegisterResponse{User: toProtoUser(u)}, nil
}

func (h *Handler) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	res, err := h.app.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, toStatus(err)
	}
	slog.Info("user logged in", "user_id", res.User.ID)
	return &identityv1.LoginResponse{
		AccessToken: res.AccessToken,
		ExpiresAt:   timestamppb.New(res.ExpiresAt),
		User:        toProtoUser(res.User),
	}, nil
}

func (h *Handler) ValidateToken(ctx context.Context, req *identityv1.ValidateTokenRequest) (*identityv1.ValidateTokenResponse, error) {
	u, err := h.app.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, toStatus(err)
	}
	return &identityv1.ValidateTokenResponse{User: toProtoUser(u)}, nil
}

// toStatus maps domain errors to the codes promised in identity.proto.
// Unknown errors are logged and hidden behind INTERNAL.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		slog.Error("internal error", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}

var roleToProto = map[domain.Role]commonv1.Role{
	domain.RoleStudent:      commonv1.Role_STUDENT,
	domain.RoleCategoryHead: commonv1.Role_CATEGORY_HEAD,
	domain.RoleStaff:        commonv1.Role_STAFF,
	domain.RoleSystemAdmin:  commonv1.Role_SYSTEM_ADMIN,
	domain.RolePresident:    commonv1.Role_PRESIDENT,
}

// toProtoUser never copies the password hash.
func toProtoUser(u domain.User) *identityv1.User {
	roles := make([]commonv1.Role, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, roleToProto[r])
	}
	return &identityv1.User{
		Id:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Roles:     roles,
		Active:    u.Active,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}
