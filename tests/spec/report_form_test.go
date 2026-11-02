package spec

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	report "github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

const reporter = "6f1c2a9e-1b7d-4c55-9a43-2f1e0b9d7c11"

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

// validForm is a report that passes every rule; each case changes one field.
func validForm() report.NewReportInput {
	return report.NewReportInput{
		ReporterID:  reporter,
		Category:    report.CategoryHostel,
		Location:    report.Location{Building: "Hostel B"},
		Description: "Tap leaking",
	}
}

func check(t *testing.T, in report.NewReportInput, wantErr error) {
	t.Helper()
	got, err := report.NewReport("r1", in, now)
	if wantErr != nil {
		assert.ErrorIs(t, err, wantErr)
		assert.ErrorIs(t, err, report.ErrInvalidReport, "every form error maps to INVALID_ARGUMENT")
		return
	}
	require.NoError(t, err)
	assert.Equal(t, report.StatusOpen, got.Status, "a new report is always OPEN")
}

// Description: required, at most 500 characters. Boundaries 0, 1, 499, 500, 501.
func TestDescriptionBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		wantErr error
	}{
		{name: "0 characters (empty)", desc: "", wantErr: report.ErrMissingDescription},
		{name: "only spaces", desc: "     ", wantErr: report.ErrMissingDescription},
		{name: "1 character (lower bound)", desc: "x"},
		{name: "499 characters", desc: strings.Repeat("a", 499)},
		{name: "500 characters (upper bound)", desc: strings.Repeat("a", 500)},
		{name: "501 characters (just over)", desc: strings.Repeat("a", 501), wantErr: report.ErrDescriptionTooLong},
		{name: "500 Dzongkha characters (counted as characters, not bytes)", desc: strings.Repeat("ཀ", 500)},
		{name: "501 Dzongkha characters", desc: strings.Repeat("ཀ", 501), wantErr: report.ErrDescriptionTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validForm()
			in.Description = tt.desc
			check(t, in, tt.wantErr)
		})
	}
}

// Category: required, one of three. Classes: missing, each valid value, unknown value.
func TestCategoryClasses(t *testing.T) {
	tests := []struct {
		name     string
		category report.Category
		wantErr  error
	}{
		{name: "missing (unspecified)", category: "", wantErr: report.ErrMissingCategory},
		{name: "HOSTEL", category: report.CategoryHostel},
		{name: "ACADEMIC_BLOCK", category: report.CategoryAcademicBlock},
		{name: "OTHERS", category: report.CategoryOthers},
		{name: "unknown value", category: "LIBRARY", wantErr: report.ErrInvalidCategory},
		{name: "wrong case", category: "hostel", wantErr: report.ErrInvalidCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validForm()
			in.Category = tt.category
			check(t, in, tt.wantErr)
		})
	}
}

// Location: building required; wing and area optional.
func TestLocationClasses(t *testing.T) {
	tests := []struct {
		name     string
		location report.Location
		wantErr  error
	}{
		{name: "missing", location: report.Location{}, wantErr: report.ErrMissingLocation},
		{name: "building only spaces", location: report.Location{Building: "   "}, wantErr: report.ErrMissingLocation},
		{name: "wing and area without building", location: report.Location{Wing: "East", Area: "Bathroom 2"}, wantErr: report.ErrMissingLocation},
		{name: "building only", location: report.Location{Building: "Hostel B"}},
		{name: "building, wing and area", location: report.Location{Building: "Hostel B", Wing: "East", Area: "Bathroom 2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validForm()
			in.Location = tt.location
			check(t, in, tt.wantErr)
		})
	}
}

// Photo: optional. report-submission checks only the photo id; size and type
// (JPG or PNG, at most 5 MB) are checked by photo-media on upload.
func TestPhotoClasses(t *testing.T) {
	ptr := func(s string) *string { return &s }
	tests := []struct {
		name    string
		photoID *string
		wantErr error
	}{
		{name: "no photo", photoID: nil},
		{name: "photo id given", photoID: ptr("3f9d6c1e-0a2b-4c3d-8e9f-1a2b3c4d5e6f")},
		{name: "empty photo id", photoID: ptr(""), wantErr: report.ErrInvalidPhoto},
		{name: "blank photo id", photoID: ptr("  "), wantErr: report.ErrInvalidPhoto},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validForm()
			in.PhotoID = tt.photoID
			check(t, in, tt.wantErr)
		})
	}
}

// Reporter: set by the gateway from the JWT, must be a UUID.
func TestReporterClasses(t *testing.T) {
	tests := []struct {
		name     string
		reporter string
		wantErr  error
	}{
		{name: "missing", reporter: "", wantErr: report.ErrMissingReporter},
		{name: "not a UUID", reporter: "student-1", wantErr: report.ErrInvalidReporter},
		{name: "UUID", reporter: reporter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validForm()
			in.ReporterID = tt.reporter
			check(t, in, tt.wantErr)
		})
	}
}
