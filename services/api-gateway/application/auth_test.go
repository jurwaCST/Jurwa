package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
	"github.com/jurwaCST/Jurwa/services/api-gateway/ports"
)

type fakeIdentity struct{ ports.Identity }

func (fakeIdentity) ValidateToken(_ context.Context, token string) (*identityv1.User, error) {
	if token == "good" {
		return &identityv1.User{Id: "u1", Email: "u1@rub.edu.bt", Roles: []commonv1.Role{commonv1.Role_STUDENT, commonv1.Role_STAFF}}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "invalid token")
}

// fakeAuthz allows exactly one (role, action) pair, or fails with err.
type fakeAuthz struct {
	role, action string
	err          error
}

func (f fakeAuthz) CheckPermission(_ context.Context, role, action, _ string) (bool, error) {
	return role == f.role && action == f.action, f.err
}

func TestAuthenticate(t *testing.T) {
	a := NewAuth(fakeIdentity{}, fakeAuthz{})
	p, err := a.Authenticate(context.Background(), "Bearer good")
	require.NoError(t, err)
	assert.Equal(t, domain.Principal{UserID: "u1", Email: "u1@rub.edu.bt", Roles: []string{"STUDENT", "STAFF"}}, p)

	_, err = a.Authenticate(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrMissingToken)
	_, err = a.Authenticate(context.Background(), "Bearer bad")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthorize(t *testing.T) {
	ctx := context.Background()
	p := domain.Principal{UserID: "u1", Roles: []string{"STUDENT", "STAFF"}}

	assert.NoError(t, NewAuth(fakeIdentity{}, fakeAuthz{role: "STAFF", action: "report.read"}).Authorize(ctx, p, "report.read", "report"), "any role is enough")
	assert.ErrorIs(t, NewAuth(fakeIdentity{}, fakeAuthz{role: "SYSTEM_ADMIN", action: "report.read"}).Authorize(ctx, p, "report.read", "report"), domain.ErrForbidden)
	assert.ErrorIs(t, NewAuth(fakeIdentity{}, fakeAuthz{}).Authorize(ctx, domain.Principal{}, "report.read", "report"), domain.ErrForbidden, "no roles")

	errDown := errors.New("authorization down")
	assert.ErrorIs(t, NewAuth(fakeIdentity{}, fakeAuthz{err: errDown}).Authorize(ctx, p, "report.read", "report"), errDown)
}
