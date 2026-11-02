package domain

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func TestNewPolicy(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		hours    int32
		wantErr  error
	}{
		// Equivalence classes and boundary values for window_hours (must be > 0).
		{"negative window", Hostel, -1, ErrInvalidWindow},
		{"zero window (boundary)", Hostel, 0, ErrInvalidWindow},
		{"one hour (smallest valid)", Hostel, 1, nil},
		{"typical window", AcademicBlock, 48, nil},
		{"largest int32", Others, math.MaxInt32, nil},
		// Category must be one of the three.
		{"category unspecified", CategoryUnspecified, 24, ErrInvalidCategory},
		{"unknown category", Category(9), 24, ErrInvalidCategory},
		{"hostel", Hostel, 24, nil},
		{"academic block", AcademicBlock, 24, nil},
		{"others", Others, 24, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPolicy(tt.category, tt.hours, "admin-1", now)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.category, p.Category)
			assert.Equal(t, tt.hours, p.WindowHours)
			assert.Equal(t, "admin-1", p.UpdatedBy)
			assert.Equal(t, now, p.UpdatedAt)
		})
	}
}

func TestPolicy_Window(t *testing.T) {
	p, err := NewPolicy(Hostel, 48, "admin-1", now)
	require.NoError(t, err)
	assert.Equal(t, 48*time.Hour, p.Window())
}

func TestCategory_String(t *testing.T) {
	assert.Equal(t, "HOSTEL", Hostel.String())
	assert.Equal(t, "ACADEMIC_BLOCK", AcademicBlock.String())
	assert.Equal(t, "OTHERS", Others.String())
	assert.Equal(t, "CATEGORY_UNSPECIFIED", CategoryUnspecified.String())
	assert.Equal(t, "CATEGORY_UNSPECIFIED", Category(9).String())
}
