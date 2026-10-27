package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/identity/application"
	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	tokens, err := domain.NewTokenIssuer([]byte("test-secret-at-least-32-bytes-long!!"), 15*time.Minute)
	require.NoError(t, err)
	return NewHandler(application.NewService(memory.NewUsers(), tokens))
}

func TestRegisterLoginValidateFlow(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	reg, err := h.Register(ctx, &identityv1.RegisterRequest{Email: "gayley@rub.edu.bt", FullName: "Gayley Choden", Password: "secret123"})
	require.NoError(t, err)
	assert.Equal(t, []commonv1.Role{commonv1.Role_STUDENT}, reg.GetUser().GetRoles())

	login, err := h.Login(ctx, &identityv1.LoginRequest{Email: "gayley@rub.edu.bt", Password: "secret123"})
	require.NoError(t, err)
	assert.NotEmpty(t, login.GetAccessToken())
	assert.True(t, login.GetExpiresAt().AsTime().After(time.Now()))

	val, err := h.ValidateToken(ctx, &identityv1.ValidateTokenRequest{AccessToken: login.GetAccessToken()})
	require.NoError(t, err)
	assert.Equal(t, reg.GetUser().GetId(), val.GetUser().GetId())
	assert.Equal(t, []commonv1.Role{commonv1.Role_STUDENT}, val.GetUser().GetRoles())
}

func TestErrorCodes(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()
	_, err := h.Register(ctx, &identityv1.RegisterRequest{Email: "gayley@rub.edu.bt", FullName: "Gayley Choden", Password: "secret123"})
	require.NoError(t, err)

	tests := []struct {
		name string
		call func() error
		want codes.Code
	}{
		{name: "invalid email", want: codes.InvalidArgument, call: func() error {
			_, err := h.Register(ctx, &identityv1.RegisterRequest{Email: "bad", FullName: "X", Password: "secret123"})
			return err
		}},
		{name: "short password", want: codes.InvalidArgument, call: func() error {
			_, err := h.Register(ctx, &identityv1.RegisterRequest{Email: "x@rub.edu.bt", FullName: "X", Password: "short"})
			return err
		}},
		{name: "duplicate email", want: codes.AlreadyExists, call: func() error {
			_, err := h.Register(ctx, &identityv1.RegisterRequest{Email: "gayley@rub.edu.bt", FullName: "X", Password: "secret123"})
			return err
		}},
		{name: "wrong password", want: codes.Unauthenticated, call: func() error {
			_, err := h.Login(ctx, &identityv1.LoginRequest{Email: "gayley@rub.edu.bt", Password: "wrong-pass"})
			return err
		}},
		{name: "invalid token", want: codes.Unauthenticated, call: func() error {
			_, err := h.ValidateToken(ctx, &identityv1.ValidateTokenRequest{AccessToken: "not.a.jwt"})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, status.Code(tt.call()))
		})
	}
}
