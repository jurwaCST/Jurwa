package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	"github.com/jurwaCST/Jurwa/services/routing-rules/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/routing-rules/application"
	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

func newHandler() *Handler {
	return NewHandler(application.NewService(memory.NewRepository(domain.DefaultRules())))
}

func TestGetRule(t *testing.T) {
	tests := []struct {
		name     string
		category commonv1.Category
		want     string
		wantCode codes.Code
	}{
		{name: "hostel", category: commonv1.Category_HOSTEL, want: "DSA"},
		{name: "academic block", category: commonv1.Category_ACADEMIC_BLOCK, want: "DAA"},
		{name: "others", category: commonv1.Category_OTHERS, want: "SSO"},
		{name: "unspecified", category: commonv1.Category_CATEGORY_UNSPECIFIED, wantCode: codes.InvalidArgument},
		{name: "unknown number", category: commonv1.Category(99), wantCode: codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := newHandler().GetRule(context.Background(), &routingrulesv1.GetRuleRequest{
				Category: tt.category,
				Location: &commonv1.Location{Building: "Hostel B"},
			})
			if tt.wantCode != codes.OK {
				assert.Equal(t, tt.wantCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, resp.GetRule().GetTargetPosition())
			assert.Equal(t, tt.category, resp.GetRule().GetCategory())
			assert.NotEmpty(t, resp.GetRule().GetId())
		})
	}
}

func TestListAndUpsertRules(t *testing.T) {
	ctx := context.Background()
	h := newHandler()

	list, err := h.ListRules(ctx, &routingrulesv1.ListRulesRequest{})
	require.NoError(t, err)
	assert.Len(t, list.GetRules(), 3)

	up, err := h.UpsertRule(ctx, &routingrulesv1.UpsertRuleRequest{Rule: &routingrulesv1.Rule{
		Category: commonv1.Category_HOSTEL, Building: "Hostel B", TargetPosition: "SSO", Active: true,
	}})
	require.NoError(t, err)
	assert.Equal(t, int32(domain.DefaultPriority), up.GetRule().GetPriority())
	assert.NotNil(t, up.GetRule().GetUpdatedAt())

	got, err := h.GetRule(ctx, &routingrulesv1.GetRuleRequest{
		Category: commonv1.Category_HOSTEL, Location: &commonv1.Location{Building: "Hostel B"},
	})
	require.NoError(t, err)
	assert.Equal(t, "SSO", got.GetRule().GetTargetPosition())

	_, err = h.UpsertRule(ctx, &routingrulesv1.UpsertRuleRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.UpsertRule(ctx, &routingrulesv1.UpsertRuleRequest{Rule: &routingrulesv1.Rule{Category: commonv1.Category_OTHERS}})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestToStatus(t *testing.T) {
	assert.Equal(t, codes.NotFound, status.Code(toStatus(domain.ErrRuleNotFound)))
	assert.Equal(t, codes.Internal, status.Code(toStatus(errors.New("database down"))))
}
