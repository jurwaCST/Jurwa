package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SCRUM-63: exactly three categories, in a fixed order, each with its Category Head.
func TestDefaultCategories(t *testing.T) {
	cats := DefaultCategories()
	require.Len(t, cats, 3)

	want := []struct {
		category Category
		head     string
	}{
		{Hostel, "DSA"},
		{AcademicBlock, "DAA"},
		{Others, "SSO"},
	}
	for i, w := range want {
		assert.Equal(t, w.category, cats[i].Category)
		assert.Equal(t, w.head, cats[i].HeadPosition)
		assert.Equal(t, i+1, cats[i].SortOrder)
		assert.True(t, cats[i].Active)
		assert.NotEmpty(t, cats[i].Name)
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantName string
		wantHead string
		wantErr  error
	}{
		{name: "hostel", category: Hostel, wantName: "Hostel", wantHead: "DSA"},
		{name: "academic block", category: AcademicBlock, wantName: "Academic Block", wantHead: "DAA"},
		{name: "others", category: Others, wantName: "Others", wantHead: "SSO"},
		{name: "unspecified is invalid", category: "", wantErr: ErrInvalidCategory},
		{name: "value not in the list", category: "LIBRARY", wantErr: ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Find(DefaultCategories(), tt.category)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantHead, got.HeadPosition)
		})
	}
}

// SCRUM-64: only a System Administrator may rename a category, and the name must not be blank.
func TestRename(t *testing.T) {
	hostel := DefaultCategories()[0]

	tests := []struct {
		name     string
		newName  string
		newDesc  string
		role     Role
		wantName string
		wantDesc string
		wantErr  error
	}{
		{name: "admin renames", newName: "Hostels", role: SystemAdmin, wantName: "Hostels", wantDesc: hostel.Description},
		{name: "name is trimmed", newName: "  Hostel Rooms ", role: SystemAdmin, wantName: "Hostel Rooms", wantDesc: hostel.Description},
		{name: "admin changes help text", newName: "Hostel", newDesc: "Rooms and bathrooms", role: SystemAdmin, wantName: "Hostel", wantDesc: "Rooms and bathrooms"},
		{name: "empty name rejected", newName: "", role: SystemAdmin, wantErr: ErrInvalidName},
		{name: "blank name rejected", newName: "   ", role: SystemAdmin, wantErr: ErrInvalidName},
		{name: "name too long rejected", newName: string(make([]byte, MaxNameLength+1)), role: SystemAdmin, wantErr: ErrInvalidName},
		{name: "student is denied", newName: "Hostels", role: "STUDENT", wantErr: ErrPermissionDenied},
		{name: "category head is denied", newName: "Hostels", role: "CATEGORY_HEAD", wantErr: ErrPermissionDenied},
		{name: "missing role is denied", newName: "Hostels", role: "", wantErr: ErrPermissionDenied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Rename(hostel, tt.newName, tt.newDesc, tt.role)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantDesc, got.Description)
			assert.Equal(t, hostel.Category, got.Category, "category code never changes")
			assert.Equal(t, hostel.HeadPosition, got.HeadPosition, "renaming never changes routing")
			assert.True(t, got.Active)
		})
	}
}

// Routing depends on the three categories, so none of them can be removed.
func TestCheckRemovable(t *testing.T) {
	for _, c := range []Category{Hostel, AcademicBlock, Others} {
		assert.ErrorIs(t, CheckRemovable(c), ErrRequiredCategory, c)
	}
	assert.ErrorIs(t, CheckRemovable(""), ErrInvalidCategory)
	assert.ErrorIs(t, CheckRemovable("LIBRARY"), ErrNotFound)
}
