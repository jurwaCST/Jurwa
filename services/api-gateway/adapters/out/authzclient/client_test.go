package authzclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authorizationv1 "github.com/jurwaCST/Jurwa/proto/gen/authorization/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
)

// fakeAuthzClient allows only STUDENT on report.submit and records each request.
type fakeAuthzClient struct {
	got   *authorizationv1.CheckPermissionRequest
	calls int
	err   error
}

func (f *fakeAuthzClient) CheckPermission(_ context.Context, in *authorizationv1.CheckPermissionRequest, _ ...grpc.CallOption) (*authorizationv1.CheckPermissionResponse, error) {
	f.got, f.calls = in, f.calls+1
	if f.err != nil {
		return nil, f.err
	}
	if in.GetRole() == commonv1.Role_STUDENT && in.GetAction() == "report.submit" {
		return &authorizationv1.CheckPermissionResponse{Allowed: true}, nil
	}
	return &authorizationv1.CheckPermissionResponse{Allowed: false, Reason: "not allowed"}, nil
}

func TestCheckPermission(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAuthzClient{}
	c := New(fake)

	ok, err := c.CheckPermission(ctx, "STUDENT", "report.submit", "/reports")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, commonv1.Role_STUDENT, fake.got.GetRole(), "role name becomes the proto enum")
	assert.Equal(t, "report.submit", fake.got.GetAction())
	assert.Equal(t, "/reports", fake.got.GetResource())

	ok, err = c.CheckPermission(ctx, "CATEGORY_HEAD", "report.submit", "/reports")
	require.NoError(t, err, "a denial is an answer, not an error")
	assert.False(t, ok)
}

func TestUnknownRoleIsDeniedWithoutACall(t *testing.T) {
	fake := &fakeAuthzClient{}
	for _, role := range []string{"", "ROLE_UNSPECIFIED", "ROOT"} {
		ok, err := New(fake).CheckPermission(context.Background(), role, "report.read", "/reports/r1")
		require.NoError(t, err)
		assert.False(t, ok, role)
	}
	assert.Equal(t, 0, fake.calls)
}

// When authorization cannot answer, the error is passed on so the gateway
// refuses the request (fail closed) instead of letting it through.
func TestErrorsArePassedOn(t *testing.T) {
	fake := &fakeAuthzClient{err: status.Error(codes.Unavailable, "down")}
	ok, err := New(fake).CheckPermission(context.Background(), "STUDENT", "report.submit", "/reports")
	assert.False(t, ok)
	assert.Equal(t, codes.Unavailable, status.Code(err))
}
