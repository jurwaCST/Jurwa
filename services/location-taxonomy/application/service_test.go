package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

// fakeRepo returns the entries of the asked category, or err when set.
type fakeRepo struct {
	entries []domain.Entry
	err     error
}

func (f fakeRepo) ListByCategory(_ context.Context, c domain.Category) ([]domain.Entry, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []domain.Entry
	for _, e := range f.entries {
		if e.Category == c {
			out = append(out, e)
		}
	}
	return out, nil
}

func seed() fakeRepo {
	return fakeRepo{entries: []domain.Entry{
		{ID: "h1", Category: domain.CategoryHostel, Location: domain.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}, Active: true},
		{ID: "h2", Category: domain.CategoryHostel, Location: domain.Location{Building: "Old Hostel"}, Active: false},
		{ID: "a1", Category: domain.CategoryAcademicBlock, Location: domain.Location{Building: "Academic Block 2", Area: "Lab 3"}, Active: true},
	}}
}

func TestListLocationsReturnsActiveEntriesOfCategory(t *testing.T) {
	svc := NewService(seed())

	got, err := svc.ListLocations(context.Background(), domain.CategoryHostel)

	require.NoError(t, err)
	require.Len(t, got, 1, "inactive and other-category entries are left out")
	assert.Equal(t, "h1", got[0].ID)
}

func TestListLocationsEmptyCategory(t *testing.T) {
	svc := NewService(seed())

	got, err := svc.ListLocations(context.Background(), domain.CategoryOthers)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestListLocationsRejectsBadCategory(t *testing.T) {
	svc := NewService(seed())

	_, err := svc.ListLocations(context.Background(), "LIBRARY")

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
		{name: "known location is valid", category: domain.CategoryHostel, location: domain.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}, wantValid: true},
		{name: "location from another category is invalid", category: domain.CategoryHostel, location: domain.Location{Building: "Academic Block 2", Area: "Lab 3"}, wantReason: domain.ReasonUnknownBuilding},
		{name: "unknown area is invalid", category: domain.CategoryAcademicBlock, location: domain.Location{Building: "Academic Block 2", Area: "Lab 9"}, wantReason: domain.ReasonUnknownWingOrArea},
		{name: "missing category is an error", category: "", location: domain.Location{Building: "Hostel B"}, wantErr: domain.ErrMissingCategory},
		{name: "missing building is an error", category: domain.CategoryHostel, location: domain.Location{}, wantErr: domain.ErrMissingBuilding},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(seed())

			valid, reason, err := svc.ValidateLocation(context.Background(), tt.category, tt.location)

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

func TestRepositoryErrorIsNotAnInvalidQuery(t *testing.T) {
	svc := NewService(fakeRepo{err: errors.New("database down")})

	_, err := svc.ListLocations(context.Background(), domain.CategoryHostel)
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidQuery)

	_, _, err = svc.ValidateLocation(context.Background(), domain.CategoryHostel, domain.Location{Building: "Hostel B"})
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidQuery)
}
