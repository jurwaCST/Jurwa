package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

const (
	feedbackID = "11111111-1111-1111-1111-111111111111"
	reportID   = "33333333-3333-3333-3333-333333333333"
	studentID  = "22222222-2222-2222-2222-222222222222"
)

func TestNewFeedbackValid(t *testing.T) {
	f, err := NewFeedback(feedbackID, reportID, studentID, 4, "Fixed quickly, thank you.", now)
	require.NoError(t, err)
	assert.Equal(t, feedbackID, f.ID)
	assert.Equal(t, reportID, f.ReportID)
	assert.Equal(t, studentID, f.StudentUserID)
	assert.Equal(t, int32(4), f.Rating)
	assert.Equal(t, "Fixed quickly, thank you.", f.Comment)
	assert.Equal(t, now, f.CreatedAt)
}

func TestRatingBoundaries(t *testing.T) {
	tests := []struct {
		rating  int32
		wantErr error
	}{
		{-1, ErrInvalidRating},
		{0, ErrInvalidRating}, // below the lowest
		{1, nil},              // lowest
		{2, nil},
		{3, nil},
		{4, nil},
		{5, nil},              // highest
		{6, ErrInvalidRating}, // above the highest
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("rating %d", tt.rating), func(t *testing.T) {
			_, err := NewFeedback(feedbackID, reportID, studentID, tt.rating, "", now)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCommentBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		wantErr error
	}{
		{"empty comment is allowed", "", nil},
		{"blank comment becomes empty", "   ", nil},
		{"499 characters", strings.Repeat("a", 499), nil},
		{"500 characters (limit)", strings.Repeat("a", 500), nil},
		{"501 characters", strings.Repeat("a", 501), ErrCommentTooLong},
		{"500 Dzongkha letters", strings.Repeat("ཀ", 500), nil},
		{"501 Dzongkha letters", strings.Repeat("ཀ", 501), ErrCommentTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFeedback(feedbackID, reportID, studentID, 5, tt.comment, now)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCommentIsTrimmed(t *testing.T) {
	f, err := NewFeedback(feedbackID, reportID, studentID, 5, "  good  ", now)
	require.NoError(t, err)
	assert.Equal(t, "good", f.Comment)
}

func TestIDsMustBeUUIDs(t *testing.T) {
	_, err := NewFeedback(feedbackID, "", studentID, 5, "", now)
	assert.ErrorIs(t, err, ErrInvalidReportID)
	_, err = NewFeedback(feedbackID, "report-1", studentID, 5, "", now)
	assert.ErrorIs(t, err, ErrInvalidReportID)
	_, err = NewFeedback(feedbackID, reportID, "", 5, "", now)
	assert.ErrorIs(t, err, ErrInvalidStudentID)
	_, err = NewFeedback(feedbackID, reportID, "gayley", 5, "", now)
	assert.ErrorIs(t, err, ErrInvalidStudentID)
}

func TestValidateReportID(t *testing.T) {
	assert.NoError(t, ValidateReportID(reportID))
	assert.ErrorIs(t, ValidateReportID("x"), ErrInvalidReportID)
}
