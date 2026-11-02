package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/application"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

func newHandler() *Handler {
	return NewHandler(application.NewService(memory.NewRepository(domain.DefaultCategories())))
}

func TestListCategories(t *testing.T) {
	resp, err := newHandler().ListCategories(context.Background(), &categorytaxonomyv1.ListCategoriesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetCategories(), 3)
	assert.Equal(t, commonv1.Category_HOSTEL, resp.GetCategories()[0].GetCategory())
	assert.Equal(t, commonv1.Category_ACADEMIC_BLOCK, resp.GetCategories()[1].GetCategory())
	assert.Equal(t, commonv1.Category_OTHERS, resp.GetCategories()[2].GetCategory())
	assert.Equal(t, "DSA", resp.GetCategories()[0].GetHeadPosition())
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		category commonv1.Category
		wantName string
		wantCode codes.Code
	}{
		{name: "hostel", category: commonv1.Category_HOSTEL, wantName: "Hostel"},
		{name: "academic block", category: commonv1.Category_ACADEMIC_BLOCK, wantName: "Academic Block"},
		{name: "others", category: commonv1.Category_OTHERS, wantName: "Others"},
		{name: "unspecified", category: commonv1.Category_CATEGORY_UNSPECIFIED, wantCode: codes.InvalidArgument},
		{name: "value not in the list", category: commonv1.Category(99), wantCode: codes.NotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := newHandler().GetCategory(context.Background(), &categorytaxonomyv1.GetCategoryRequest{Category: tt.category})
			if tt.wantCode != codes.OK {
				assert.Equal(t, tt.wantCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, resp.GetCategory().GetName())
		})
	}
}

func TestUpsertCategory(t *testing.T) {
	ctx := context.Background()
	h := newHandler()

	resp, err := h.UpsertCategory(ctx, &categorytaxonomyv1.UpsertCategoryRequest{
		Category: commonv1.Category_HOSTEL, DisplayName: "Hostels", ActorRole: commonv1.Role_SYSTEM_ADMIN,
	})
	require.NoError(t, err)
	assert.Equal(t, "Hostels", resp.GetCategory().GetName())

	tests := []struct {
		name string
		req  *categorytaxonomyv1.UpsertCategoryRequest
		want codes.Code
	}{
		{name: "student", req: &categorytaxonomyv1.UpsertCategoryRequest{Category: commonv1.Category_HOSTEL, DisplayName: "x", ActorRole: commonv1.Role_STUDENT}, want: codes.PermissionDenied},
		{name: "empty name", req: &categorytaxonomyv1.UpsertCategoryRequest{Category: commonv1.Category_HOSTEL, ActorRole: commonv1.Role_SYSTEM_ADMIN}, want: codes.InvalidArgument},
		{name: "unspecified", req: &categorytaxonomyv1.UpsertCategoryRequest{DisplayName: "x", ActorRole: commonv1.Role_SYSTEM_ADMIN}, want: codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := h.UpsertCategory(ctx, tt.req)
			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}

func TestToStatus(t *testing.T) {
	assert.Equal(t, codes.FailedPrecondition, status.Code(toStatus(domain.ErrRequiredCategory)))
	assert.Equal(t, codes.Internal, status.Code(toStatus(errors.New("database down"))))
}
