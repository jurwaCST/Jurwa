package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testID       = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	testReporter = "0b9e8d7c-6a5b-4c3d-9e2f-1a0b9c8d7e6f"
)

var testNow = time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)

func ptr(s string) *string { return &s }

// validInput returns a form that passes every rule. Each test case changes one field.
func validInput() NewReportInput {
	return NewReportInput{
		ReporterID:  testReporter,
		Category:    CategoryHostel,
		Location:    Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"},
		Description: "Tap in the bathroom is leaking.",
	}
}

func TestNewReportValidation(t *testing.T) {
	tests := []struct {
		name    string
		change  func(in *NewReportInput)
		wantErr error
	}{
		{name: "valid report is accepted", change: func(in *NewReportInput) {}},
		{name: "missing reporter is rejected", change: func(in *NewReportInput) { in.ReporterID = "" }, wantErr: ErrMissingReporter},
		{name: "reporter that is not a UUID is rejected", change: func(in *NewReportInput) { in.ReporterID = "student-1" }, wantErr: ErrInvalidReporter},
		{name: "missing category is rejected", change: func(in *NewReportInput) { in.Category = "" }, wantErr: ErrMissingCategory},
		{name: "invalid category is rejected", change: func(in *NewReportInput) { in.Category = "LIBRARY" }, wantErr: ErrInvalidCategory},
		{name: "academic block category is accepted", change: func(in *NewReportInput) { in.Category = CategoryAcademicBlock }},
		{name: "others category is accepted", change: func(in *NewReportInput) { in.Category = CategoryOthers }},
		{name: "missing location is rejected", change: func(in *NewReportInput) { in.Location = Location{} }, wantErr: ErrMissingLocation},
		{name: "whitespace building is rejected", change: func(in *NewReportInput) { in.Location.Building = "   " }, wantErr: ErrMissingLocation},
		{name: "building without wing and area is accepted", change: func(in *NewReportInput) { in.Location = Location{Building: "Academic Block 2"} }},
		{name: "empty description is rejected", change: func(in *NewReportInput) { in.Description = "" }, wantErr: ErrMissingDescription},
		{name: "whitespace description is rejected", change: func(in *NewReportInput) { in.Description = " \t\n " }, wantErr: ErrMissingDescription},
		{name: "500 character description is accepted", change: func(in *NewReportInput) { in.Description = strings.Repeat("a", 500) }},
		{name: "501 character description is rejected", change: func(in *NewReportInput) { in.Description = strings.Repeat("a", 501) }, wantErr: ErrDescriptionTooLong},
		{name: "500 multi-byte characters are accepted", change: func(in *NewReportInput) { in.Description = strings.Repeat("ཀ", 500) }},
		{name: "no photo is accepted", change: func(in *NewReportInput) { in.PhotoID = nil }},
		{name: "blank photo id is rejected", change: func(in *NewReportInput) { in.PhotoID = ptr(" ") }, wantErr: ErrInvalidPhoto},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.change(&in)

			report, err := NewReport(testID, in, testNow)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrInvalidReport, "every rule error is an invalid-report error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, StatusOpen, report.Status)
		})
	}
}

func TestNewReportFillsFields(t *testing.T) {
	in := validInput()
	in.PhotoID = ptr("9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d")

	report, err := NewReport(testID, in, testNow)

	require.NoError(t, err)
	assert.Equal(t, testID, report.ID)
	assert.Equal(t, testReporter, report.ReporterID)
	assert.Equal(t, CategoryHostel, report.Category)
	assert.Equal(t, in.Location, report.Location)
	assert.Equal(t, in.Description, report.Description)
	require.NotNil(t, report.PhotoID, "photo id is kept")
	assert.Equal(t, *in.PhotoID, *report.PhotoID)
	assert.Equal(t, StatusOpen, report.Status)
	assert.Equal(t, testNow, report.CreatedAt)
}

func TestRoutingKeyIgnoresDescription(t *testing.T) {
	a := validInput()
	b := validInput()
	b.Description = "Completely different words, even mentioning the library and the DAA."

	ra, err := NewReport(testID, a, testNow)
	require.NoError(t, err)
	rb, err := NewReport(testID, b, testNow)
	require.NoError(t, err)

	assert.Equal(t, ra.RoutingKey(), rb.RoutingKey(), "description must never change routing")
	assert.Equal(t, RoutingKey{Category: CategoryHostel, Location: a.Location}, ra.RoutingKey())
}
