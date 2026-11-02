package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Errors. The gRPC adapter maps them to status codes.
var (
	ErrInvalidReportID  = errors.New("report_id must be a UUID")
	ErrInvalidStudentID = errors.New("student_user_id must be a UUID")
	ErrInvalidRating    = errors.New("rating must be 1 to 5")
	ErrCommentTooLong   = errors.New("comment must be at most 500 characters")
	ErrNotResolved      = errors.New("feedback is allowed only after the report is RESOLVED")
	ErrAlreadyExists    = errors.New("this report already has feedback")
	ErrNotFound         = errors.New("this report has no feedback")
	ErrUnavailable      = errors.New("status-lifecycle is unavailable")
)

const (
	MinRating       = 1
	MaxRating       = 5
	MaxCommentChars = 500 // from proto/feedback/v1/feedback.proto
)

// Feedback is a student's rating of how one report was resolved.
type Feedback struct {
	ID            string
	ReportID      string
	StudentUserID string
	Rating        int32
	Comment       string // "" when the student left no comment
	CreatedAt     time.Time
}

// ValidateReportID checks that id is a UUID.
func ValidateReportID(id string) error {
	if uuid.Validate(id) != nil {
		return ErrInvalidReportID
	}
	return nil
}

// NewFeedback checks the input and returns the feedback. Whether the report
// is resolved, and whether it already has feedback, is checked by the
// application layer, which has the store and the status checker.
func NewFeedback(id, reportID, studentID string, rating int32, comment string, now time.Time) (Feedback, error) {
	if err := ValidateReportID(reportID); err != nil {
		return Feedback{}, err
	}
	if uuid.Validate(studentID) != nil {
		return Feedback{}, ErrInvalidStudentID
	}
	if rating < MinRating || rating > MaxRating {
		return Feedback{}, ErrInvalidRating
	}
	c := strings.TrimSpace(comment)
	if utf8.RuneCountInString(c) > MaxCommentChars {
		return Feedback{}, ErrCommentTooLong
	}
	return Feedback{
		ID:            id,
		ReportID:      reportID,
		StudentUserID: studentID,
		Rating:        rating,
		Comment:       c,
		CreatedAt:     now,
	}, nil
}
