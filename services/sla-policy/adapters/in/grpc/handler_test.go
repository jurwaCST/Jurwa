package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slapolicyv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-policy/v1"
	"github.com/jurwaCST/Jurwa/services/sla-policy/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/sla-policy/application"
	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
)

const adminID = "22222222-2222-2222-2222-222222222222"

func newHandler() *Handler {
	return NewHandler(application.NewService(memory.New(), time.Now))
}

func TestUpsertThenGet(t *testing.T) {
	h := newHandler()
	ctx := context.Background()

	up, err := h.UpsertPolicy(ctx, &slapolicyv1.UpsertPolicyRequest{
		Category: commonv1.Category_HOSTEL, WindowHours: 48, UpdatedBy: adminID,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(48), up.GetPolicy().GetWindowHours())
	assert.Equal(t, adminID, up.GetPolicy().GetUpdatedBy())

	got, err := h.GetPolicy(ctx, &slapolicyv1.GetPolicyRequest{Category: commonv1.Category_HOSTEL})
	require.NoError(t, err)
	assert.Equal(t, commonv1.Category_HOSTEL, got.GetPolicy().GetCategory())
	assert.Equal(t, int32(48), got.GetPolicy().GetWindowHours())
}

func TestErrorCodes(t *testing.T) {
	h := newHandler()
	ctx := context.Background()

	_, err := h.GetPolicy(ctx, &slapolicyv1.GetPolicyRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "category unspecified")

	_, err = h.GetPolicy(ctx, &slapolicyv1.GetPolicyRequest{Category: commonv1.Category_OTHERS})
	assert.Equal(t, codes.NotFound, status.Code(err), "no policy set")

	_, err = h.UpsertPolicy(ctx, &slapolicyv1.UpsertPolicyRequest{Category: commonv1.Category_HOSTEL, WindowHours: 0, UpdatedBy: adminID})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "zero window")

	_, err = h.UpsertPolicy(ctx, &slapolicyv1.UpsertPolicyRequest{Category: commonv1.Category_HOSTEL, WindowHours: 24})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "updated_by missing")
}

func TestToStatusError_Internal(t *testing.T) {
	assert.Equal(t, codes.Internal, status.Code(toStatusError(errors.New("db down"))))
}

func TestToProto_NoUpdatedBy(t *testing.T) {
	out := toProto(domain.Policy{Category: domain.Others, WindowHours: 12})
	assert.Nil(t, out.UpdatedBy)
}
