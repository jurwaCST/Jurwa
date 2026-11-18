package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

func newTestService() *Service {
	return NewService(memory.NewRepository(memory.SampleEntries()))
}

// failingRepo returns an error, to check a storage failure is not reported as bad input.
type failingRepo struct{}

func (failingRepo) ListByCategory(context.Context, domain.Category) ([]domain.Entry, error) {
	return nil, errors.New("database down")
}

func TestListLocationsReturnsOnlyActiveRowsOfThatCategory(t *testing.T) {
	tests := []struct {
		name     string
		category domain.Category
		wantLen  int
	}{
		{name: "hostel", category: domain.CategoryHostel, wantLen: 525},
		{name: "academic block", category: domain.CategoryAcademicBlock, wantLen: 20},
		{name: "others leaves out the switched-off sample rows", category: domain.CategoryOthers, wantLen: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTestService().ListLocations(context.Background(), tt.category)

			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			for _, e := range got {
				assert.Equal(t, tt.category, e.Category)
				assert.True(t, e.Active)
			}
		})
	}
}

func TestListLocationsRejectsBadCategory(t *testing.T) {
	_, err := newTestService().ListLocations(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrMissingCategory)

	_, err = newTestService().ListLocations(context.Background(), "LIBRARY")
	assert.ErrorIs(t, err, domain.ErrInvalidCategory)
}

func TestValidateLocation(t *testing.T) {
	tests := []struct {
		name       string
		category   domain.Category
		location   domain.Location
		wantErr    error
		wantValid  bool
		wantReason string
	}{
		{name: "real hostel location", category: domain.CategoryHostel, location: domain.Location{Building: "HC", Wing: "Floor 1", Area: "Left wing bathroom"}, wantValid: true},
		{name: "real block location", category: domain.CategoryAcademicBlock, location: domain.Location{Building: "Academic Block", Area: "LH1"}, wantValid: true},
		{name: "campus-wide area", category: domain.CategoryOthers, location: domain.Location{Building: domain.CampusBuilding, Area: "Library"}, wantValid: true},
		{name: "switched-off sample area", category: domain.CategoryOthers, location: domain.Location{Building: domain.CampusBuilding, Area: "Wi-Fi"}, wantReason: domain.ReasonUnknownArea},
		{name: "switched-off sample building", category: domain.CategoryHostel, location: domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}, wantReason: domain.ReasonUnknownBuilding},
		{name: "building of another category", category: domain.CategoryHostel, location: domain.Location{Building: "Academic Block"}, wantReason: domain.ReasonUnknownBuilding},
		{name: "floor the building does not have", category: domain.CategoryHostel, location: domain.Location{Building: "HE", Wing: "Floor 3"}, wantReason: domain.ReasonUnknownWing},
		{name: "missing category", category: "", location: domain.Location{Building: "Hostel A"}, wantErr: domain.ErrMissingCategory},
		{name: "empty building", category: domain.CategoryHostel, location: domain.Location{}, wantErr: domain.ErrMissingBuilding},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason, err := newTestService().ValidateLocation(context.Background(), tt.category, tt.location)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantReason, reason)
		})
	}
}

func TestStorageErrorIsNotInvalidInput(t *testing.T) {
	svc := NewService(failingRepo{})

	_, err := svc.ListLocations(context.Background(), domain.CategoryHostel)
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidInput)

	_, _, err = svc.ValidateLocation(context.Background(), domain.CategoryHostel, domain.Location{Building: "Hostel A"})
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidInput)
}
