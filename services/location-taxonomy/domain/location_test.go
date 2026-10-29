package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testID = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"

// sampleEntries is a small hierarchy: one hostel with two wings, one block with
// areas but no wings, a flat campus-wide list, and one switched-off row.
func sampleEntries() []Entry {
	return []Entry{
		{ID: "h1", Category: CategoryHostel, Location: Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}, Active: true},
		{ID: "h2", Category: CategoryHostel, Location: Location{Building: "Hostel A", Wing: "Wing 2", Area: "Common Room"}, Active: true},
		{ID: "b1", Category: CategoryAcademicBlock, Location: Location{Building: "Academic Block 1", Area: "Lab 3"}, Active: true},
		{ID: "o1", Category: CategoryOthers, Location: Location{Building: CampusBuilding, Area: "Wi-Fi"}, Active: true},
		{ID: "o2", Category: CategoryOthers, Location: Location{Building: CampusBuilding, Area: "Old Workshop"}, Active: false},
	}
}

func TestNewEntry(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		category Category
		location Location
		wantErr  error
	}{
		{name: "hostel with wing and area", id: testID, category: CategoryHostel, location: Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}},
		{name: "block with area and no wing", id: testID, category: CategoryAcademicBlock, location: Location{Building: "Academic Block 1", Area: "Lab 3"}},
		{name: "campus-wide area", id: testID, category: CategoryOthers, location: Location{Building: CampusBuilding, Area: "Canteen"}},
		{name: "id that is not a UUID", id: "h1", category: CategoryHostel, location: Location{Building: "Hostel A"}, wantErr: ErrInvalidID},
		{name: "missing category", id: testID, category: "", location: Location{Building: "Hostel A"}, wantErr: ErrMissingCategory},
		{name: "unknown category", id: testID, category: "LIBRARY", location: Location{Building: "Hostel A"}, wantErr: ErrInvalidCategory},
		{name: "missing building", id: testID, category: CategoryHostel, location: Location{Wing: "Wing 1"}, wantErr: ErrMissingBuilding},
		{name: "campus-wide entry with a wing", id: testID, category: CategoryOthers, location: Location{Building: CampusBuilding, Wing: "East", Area: "Wi-Fi"}, wantErr: ErrWingNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := NewEntry(tt.id, tt.category, tt.location)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrInvalidInput, "every rule error is an invalid-input error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.id, e.ID)
			assert.Equal(t, tt.category, e.Category)
			assert.Equal(t, tt.location, e.Location)
			assert.True(t, e.Active, "a new entry is active")
		})
	}
}

func TestValidateCategory(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  error
	}{
		{name: "hostel", category: CategoryHostel},
		{name: "academic block", category: CategoryAcademicBlock},
		{name: "others", category: CategoryOthers},
		{name: "missing", category: "", wantErr: ErrMissingCategory},
		{name: "unknown", category: "LIBRARY", wantErr: ErrInvalidCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategory(tt.category)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrInvalidInput)
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
		{name: "building, wing and area", location: Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}, wantValid: true},
		{name: "building only", location: Location{Building: "Hostel A"}, wantValid: true},
		{name: "building and wing", location: Location{Building: "Hostel A", Wing: "Wing 2"}, wantValid: true},
		{name: "building and area without wing", location: Location{Building: "Academic Block 1", Area: "Lab 3"}, wantValid: true},
		{name: "campus-wide area", location: Location{Building: CampusBuilding, Area: "Wi-Fi"}, wantValid: true},
		{name: "case and spaces are ignored", location: Location{Building: " hostel a ", Wing: "WING 1", Area: "bathroom 1 "}, wantValid: true},
		{name: "unknown building", location: Location{Building: "Hostel Z"}, wantReason: ReasonUnknownBuilding},
		{name: "wing of another building", location: Location{Building: "Academic Block 1", Wing: "Wing 1"}, wantReason: ReasonUnknownWing},
		{name: "area of another wing", location: Location{Building: "Hostel A", Wing: "Wing 2", Area: "Bathroom 1"}, wantReason: ReasonUnknownArea},
		{name: "area that needs its wing", location: Location{Building: "Hostel A", Area: "Bathroom 1"}, wantReason: ReasonUnknownArea},
		{name: "inactive row does not match", location: Location{Building: CampusBuilding, Area: "Old Workshop"}, wantReason: ReasonUnknownArea},
		{name: "empty building is an error", location: Location{Wing: "Wing 1"}, wantErr: ErrMissingBuilding},
		{name: "whitespace building is an error", location: Location{Building: "   "}, wantErr: ErrMissingBuilding},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason, err := CheckLocation(sampleEntries(), tt.location)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrInvalidInput)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantReason, reason)
		})
	}
}

func TestCheckLocationWithNoEntries(t *testing.T) {
	valid, reason, err := CheckLocation(nil, Location{Building: "Hostel A"})

	require.NoError(t, err)
	assert.False(t, valid)
	assert.Equal(t, ReasonUnknownBuilding, reason)
}

func TestActiveOnly(t *testing.T) {
	got := ActiveOnly(sampleEntries())

	require.Len(t, got, 4, "the switched-off row is left out")
	for _, e := range got {
		assert.True(t, e.Active, "entry %s is active", e.ID)
	}
	assert.Equal(t, "h1", got[0].ID, "order is kept")
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID(testID))
	assert.False(t, IsUUID("hostel-a"))
	assert.False(t, IsUUID(""))
}
