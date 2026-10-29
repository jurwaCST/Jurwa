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
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	"github.com/jurwaCST/Jurwa/services/location-taxonomy/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/location-taxonomy/application"
	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

func newTestHandler() *Handler {
	return NewHandler(application.NewService(memory.NewRepository(memory.SeedEntries())))
}

// brokenApp fails every call, to check internal errors are hidden.
type brokenApp struct{}

func (brokenApp) ListLocations(context.Context, domain.Category) ([]domain.Entry, error) {
	return nil, errors.New("list locations: database password is wrong")
}

func (brokenApp) ValidateLocation(context.Context, domain.Category, domain.Location) (bool, string, error) {
	return false, "", errors.New("list locations: database password is wrong")
}

func TestListLocationsReturnsActiveSeedEntries(t *testing.T) {
	h := newTestHandler()

	resp, err := h.ListLocations(context.Background(), &locationtaxonomyv1.ListLocationsRequest{Category: commonv1.Category_OTHERS})

	require.NoError(t, err)
	got := resp.GetLocations()
	require.Len(t, got, 3, "the inactive Old Workshop is left out")
	for _, e := range got {
		assert.True(t, domain.IsUUID(e.GetId()))
		assert.Equal(t, commonv1.Category_OTHERS, e.GetCategory())
		assert.NotEqual(t, "Old Workshop", e.GetLocation().GetBuilding())
	}
	assert.Equal(t, "Library", got[0].GetLocation().GetBuilding())
}

func TestListLocationsKeepsCategory(t *testing.T) {
	for _, c := range []commonv1.Category{commonv1.Category_HOSTEL, commonv1.Category_ACADEMIC_BLOCK, commonv1.Category_OTHERS} {
		t.Run(c.String(), func(t *testing.T) {
			resp, err := newTestHandler().ListLocations(context.Background(), &locationtaxonomyv1.ListLocationsRequest{Category: c})

			require.NoError(t, err)
			require.NotEmpty(t, resp.GetLocations())
			for _, e := range resp.GetLocations() {
				assert.Equal(t, c, e.GetCategory())
			}
		})
	}
}

func TestListLocationsInvalidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category commonv1.Category
	}{
		{name: "unspecified category", category: commonv1.Category_CATEGORY_UNSPECIFIED},
		{name: "unknown enum number", category: commonv1.Category(99)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTestHandler().ListLocations(context.Background(), &locationtaxonomyv1.ListLocationsRequest{Category: tt.category})

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestValidateLocation(t *testing.T) {
	tests := []struct {
		name       string
		category   commonv1.Category
		location   *commonv1.Location
		wantValid  bool
		wantReason string
	}{
		{name: "known hostel location", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}, wantValid: true},
		{name: "known block location", category: commonv1.Category_ACADEMIC_BLOCK, location: &commonv1.Location{Building: "Academic Block 2", Area: "Lab 3"}, wantValid: true},
		{name: "unknown building", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "Hostel Z"}, wantReason: domain.ReasonUnknownBuilding},
		{name: "unknown area", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Roof"}, wantReason: domain.ReasonUnknownWingOrArea},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := newTestHandler().ValidateLocation(context.Background(), &locationtaxonomyv1.ValidateLocationRequest{Category: tt.category, Location: tt.location})

			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, resp.GetValid())
			assert.Equal(t, tt.wantReason, resp.GetReason())
		})
	}
}

func TestValidateLocationInvalidArgument(t *testing.T) {
	tests := []struct {
		name string
		req  *locationtaxonomyv1.ValidateLocationRequest
	}{
		{name: "missing category", req: &locationtaxonomyv1.ValidateLocationRequest{Location: &commonv1.Location{Building: "Hostel B"}}},
		{name: "missing location", req: &locationtaxonomyv1.ValidateLocationRequest{Category: commonv1.Category_HOSTEL}},
		{name: "blank building", req: &locationtaxonomyv1.ValidateLocationRequest{Category: commonv1.Category_HOSTEL, Location: &commonv1.Location{Building: "  "}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTestHandler().ValidateLocation(context.Background(), tt.req)

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestInternalErrorsAreHidden(t *testing.T) {
	h := NewHandler(brokenApp{})

	_, err := h.ListLocations(context.Background(), &locationtaxonomyv1.ListLocationsRequest{Category: commonv1.Category_HOSTEL})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")

	_, err = h.ValidateLocation(context.Background(), &locationtaxonomyv1.ValidateLocationRequest{Category: commonv1.Category_HOSTEL, Location: &commonv1.Location{Building: "Hostel B"}})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")
}

func TestCategoryToProtoUnknownIsUnspecified(t *testing.T) {
	assert.Equal(t, commonv1.Category_CATEGORY_UNSPECIFIED, categoryToProto("LIBRARY"))
}
