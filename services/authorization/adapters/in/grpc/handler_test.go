package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authorizationv1 "github.com/jurwaCST/Jurwa/proto/gen/authorization/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	"github.com/jurwaCST/Jurwa/services/authorization/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/authorization/application"
)

func newHandler() *Handler {
	return NewHandler(application.NewService(memory.New()))
}

func TestCheckPermissionAllowedAndDenied(t *testing.T) {
	h := newHandler()
	ctx := context.Background()

	res, err := h.CheckPermission(ctx, &authorizationv1.CheckPermissionRequest{
		Role: commonv1.Role_STUDENT, Action: "report.submit", Resource: "report",
	})
	require.NoError(t, err)
	assert.True(t, res.GetAllowed())
	assert.Empty(t, res.GetReason())

	res, err = h.CheckPermission(ctx, &authorizationv1.CheckPermissionRequest{
		Role: commonv1.Role_PRESIDENT, Action: "report.reassign", Resource: "report",
	})
	require.NoError(t, err, "a denial is a normal answer, not an error")
	assert.False(t, res.GetAllowed())
	assert.Equal(t, "PRESIDENT may not report.reassign", res.GetReason())
}

func TestCheckPermissionInvalidArgument(t *testing.T) {
	h := newHandler()
	ctx := context.Background()
	tests := []struct {
		name string
		req  *authorizationv1.CheckPermissionRequest
	}{
		{"role unspecified", &authorizationv1.CheckPermissionRequest{Action: "me.read", Resource: "me"}},
		{"action missing", &authorizationv1.CheckPermissionRequest{Role: commonv1.Role_STUDENT, Resource: "me"}},
		{"resource missing", &authorizationv1.CheckPermissionRequest{Role: commonv1.Role_STUDENT, Action: "me.read"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := h.CheckPermission(ctx, tt.req)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestToStatusErrorInternal(t *testing.T) {
	err := toStatusError(errors.New("store down"))
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "store down", "internal details are not sent to the client")
}
