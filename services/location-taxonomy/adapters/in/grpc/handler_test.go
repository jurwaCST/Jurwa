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
	return NewHandler(application.NewService(memory.NewRepository(memory.SampleEntries())))
}

// brokenApp fails every call, to check internal errors are hidden.
type brokenApp struct{}

func (brokenApp) ListLocations(context.Context, domain.Category) ([]domain.Entry, error) {
	return nil, errors.New("list locations: database password is wrong")
}

func (brokenApp) ValidateLocation(context.Context, domain.Category, domain.Location) (bool, string, error) {
	return false, "", errors.New("list locations: database password is wrong")
}

func TestListLocationsKeepsCategory(t *testing.T) {
	for _, c := range []commonv1.Category{commonv1.Category_HOSTEL, commonv1.Category_ACADEMIC_BLOCK, commonv1.Category_OTHERS} {
		t.Run(c.String(), func(t *testing.T) {
			resp, err := newTestHandler().ListLocations(context.Background(), &locationtaxonomyv1.ListLocationsRequest{Category: c})

			require.NoError(t, err)
			require.NotEmpty(t, resp.GetLocations())
			for _, e := range resp.GetLocations() {
				assert.True(t, domain.IsUUID(e.GetId()))
				assert.Equal(t, c, e.GetCategory())
				assert.NotEmpty(t, e.GetLocation().GetBuilding())
			}
		})
	}
}

func TestListLocationsOthersIsFlat(t *testing.T) {
	resp, err := newTestHandler().ListLocations(context.Background(), &locationtaxonomyv1.ListLocationsRequest{Category: commonv1.Category_OTHERS})

	require.NoError(t, err)
	require.Len(t, resp.GetLocations(), 7, "the switched-off sample rows (Wi-Fi, Estate, Canteen, Old Workshop) are left out")
	for _, e := range resp.GetLocations() {
		assert.Equal(t, domain.CampusBuilding, e.GetLocation().GetBuilding())
		assert.Empty(t, e.GetLocation().GetWing(), "campus-wide locations have no wing")
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
		{name: "real building, floor and room", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "HC", Wing: "Floor 1", Area: "Room 101"}, wantValid: true},
		{name: "campus-wide area", category: commonv1.Category_OTHERS, location: &commonv1.Location{Building: "Campus", Area: "Library"}, wantValid: true},
		{name: "unknown building", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "Hostel Z"}, wantReason: domain.ReasonUnknownBuilding},
		{name: "floor the building does not have", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "HE", Wing: "Floor 3"}, wantReason: domain.ReasonUnknownWing},
		{name: "unknown area", category: commonv1.Category_HOSTEL, location: &commonv1.Location{Building: "HC", Wing: "Floor 1", Area: "Roof"}, wantReason: domain.ReasonUnknownArea},
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
		{name: "missing category", req: &locationtaxonomyv1.ValidateLocationRequest{Location: &commonv1.Location{Building: "Hostel A"}}},
		{name: "missing location", req: &locationtaxonomyv1.ValidateLocationRequest{Category: commonv1.Category_HOSTEL}},
		{name: "empty building", req: &locationtaxonomyv1.ValidateLocationRequest{Category: commonv1.Category_HOSTEL, Location: &commonv1.Location{Building: "  "}}},
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

	_, err = h.ValidateLocation(context.Background(), &locationtaxonomyv1.ValidateLocationRequest{Category: commonv1.Category_HOSTEL, Location: &commonv1.Location{Building: "Hostel A"}})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")
}

func TestCategoryToProtoUnknownIsUnspecified(t *testing.T) {
	assert.Equal(t, commonv1.Category_CATEGORY_UNSPECIFIED, categoryToProto("LIBRARY"))
}
