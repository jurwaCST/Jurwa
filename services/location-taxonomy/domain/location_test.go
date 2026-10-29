package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostelEntries is a small list for one category. The last entry is switched off.
func hostelEntries() []Entry {
	return []Entry{
		{ID: "e1", Category: CategoryHostel, Location: Location{Building: "Hostel B"}, Active: true},
		{ID: "e2", Category: CategoryHostel, Location: Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}, Active: true},
		{ID: "e3", Category: CategoryHostel, Location: Location{Building: "Hostel C", Wing: "Ground Floor", Area: "Common Room"}, Active: true},
		{ID: "e4", Category: CategoryHostel, Location: Location{Building: "Old Hostel", Area: "Laundry"}, Active: false},
	}
}

func TestValidateCategory(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  error
	}{
		{name: "hostel is accepted", category: CategoryHostel},
		{name: "academic block is accepted", category: CategoryAcademicBlock},
		{name: "others is accepted", category: CategoryOthers},
		{name: "missing category is rejected", category: "", wantErr: ErrMissingCategory},
		{name: "unknown category is rejected", category: "LIBRARY", wantErr: ErrInvalidCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategory(tt.category)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrInvalidQuery, "every rule error is an invalid-query error")
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCheckLocation(t *testing.T) {
	tests := []struct {
		name       string
		location   Location
		wantErr    error
		wantValid  bool
		wantReason string
	}{
		{name: "building only entry matches", location: Location{Building: "Hostel B"}, wantValid: true},
		{name: "full entry matches", location: Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}, wantValid: true},
		{name: "case and spaces are ignored", location: Location{Building: " hostel b ", Wing: "WING 2", Area: "bathroom 2  "}, wantValid: true},
		{name: "missing building is an error", location: Location{Wing: "Wing 2"}, wantErr: ErrMissingBuilding},
		{name: "whitespace building is an error", location: Location{Building: "   "}, wantErr: ErrMissingBuilding},
		{name: "unknown building is invalid", location: Location{Building: "Library"}, wantReason: ReasonUnknownBuilding},
		{name: "known building with unknown area is invalid", location: Location{Building: "Hostel B", Wing: "Wing 2", Area: "Roof"}, wantReason: ReasonUnknownWingOrArea},
		{name: "partial match is invalid", location: Location{Building: "Hostel C", Wing: "Ground Floor"}, wantReason: ReasonUnknownWingOrArea},
		{name: "inactive entry does not match", location: Location{Building: "Old Hostel", Area: "Laundry"}, wantReason: ReasonUnknownBuilding},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason, err := CheckLocation(hostelEntries(), tt.location)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrInvalidQuery)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantReason, reason)
		})
	}
}

func TestCheckLocationWithNoEntries(t *testing.T) {
	valid, reason, err := CheckLocation(nil, Location{Building: "Hostel B"})

	require.NoError(t, err)
	assert.False(t, valid)
	assert.Equal(t, ReasonUnknownBuilding, reason)
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID("6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"))
	assert.False(t, IsUUID("hostel-b"))
	assert.False(t, IsUUID(""))
}

func TestActiveOnly(t *testing.T) {
	got := ActiveOnly(hostelEntries())

	require.Len(t, got, 3)
	for _, e := range got {
		assert.True(t, e.Active, "entry %s should be active", e.ID)
	}
}
