package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxDescriptionLength is the longest allowed description, counted in characters (not bytes).
const MaxDescriptionLength = 500

// Category is the report category. The values match the codes stored in the databases.
type Category string

const (
	CategoryHostel        Category = "HOSTEL"         // routed to DSA
	CategoryAcademicBlock Category = "ACADEMIC_BLOCK" // routed to DAA
	CategoryOthers        Category = "OTHERS"         // campus-wide, routed to SSO
)

// Status is the report status. This service only ever creates reports as OPEN;
// status-lifecycle owns every change after that.
type Status string

const StatusOpen Status = "OPEN"

// Location says where the problem is. Building is required; wing and area are optional.
type Location struct {
	Building string
	Wing     string
	Area     string
}

// Report is a submitted facility issue report.
type Report struct {
	ID          string
	ReporterID  string
	Category    Category
	Location    Location
	Description string
	PhotoID     *string // nil when no photo was attached
	Status      Status
	CreatedAt   time.Time
}

// NewReportInput is the report form as filled in by the student.
type NewReportInput struct {
	ReporterID  string
	Category    Category
	Location    Location
	Description string
	PhotoID     *string
}

// ErrInvalidReport is wrapped by every form rule error, so callers can map all of
// them to one response (INVALID_ARGUMENT) with errors.Is.
var ErrInvalidReport = errors.New("invalid report")

var (
	ErrMissingReporter    = fmt.Errorf("%w: reporter id is required", ErrInvalidReport)
	ErrInvalidReporter    = fmt.Errorf("%w: reporter id must be a UUID", ErrInvalidReport)
	ErrMissingCategory    = fmt.Errorf("%w: category is required", ErrInvalidReport)
	ErrInvalidCategory    = fmt.Errorf("%w: category must be HOSTEL, ACADEMIC_BLOCK or OTHERS", ErrInvalidReport)
	ErrMissingLocation    = fmt.Errorf("%w: location building is required", ErrInvalidReport)
	ErrMissingDescription = fmt.Errorf("%w: description is required", ErrInvalidReport)
	ErrDescriptionTooLong = fmt.Errorf("%w: description must be at most %d characters", ErrInvalidReport, MaxDescriptionLength)
	ErrInvalidPhoto       = fmt.Errorf("%w: photo id must not be blank when given", ErrInvalidReport)
)

// ErrReportNotFound is returned when no report has the requested id.
var ErrReportNotFound = errors.New("report not found")

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s has the canonical UUID form.
func IsUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// NewReport checks the form rules and returns a new report with status OPEN.
// The id and time come from the caller so the rules stay easy to test.
func NewReport(id string, in NewReportInput, now time.Time) (Report, error) {
	if err := validate(in); err != nil {
		return Report{}, err
	}
	return Report{
		ID:          id,
		ReporterID:  in.ReporterID,
		Category:    in.Category,
		Location:    in.Location,
		Description: in.Description,
		PhotoID:     in.PhotoID,
		Status:      StatusOpen,
		CreatedAt:   now,
	}, nil
}

func validate(in NewReportInput) error {
	if in.ReporterID == "" {
		return ErrMissingReporter
	}
	if !IsUUID(in.ReporterID) {
		return ErrInvalidReporter
	}
	switch in.Category {
	case "":
		return ErrMissingCategory
	case CategoryHostel, CategoryAcademicBlock, CategoryOthers:
	default:
		return ErrInvalidCategory
	}
	if strings.TrimSpace(in.Location.Building) == "" {
		return ErrMissingLocation
	}
	if strings.TrimSpace(in.Description) == "" {
		return ErrMissingDescription
	}
	if utf8.RuneCountInString(in.Description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	if in.PhotoID != nil && strings.TrimSpace(*in.PhotoID) == "" {
		return ErrInvalidPhoto
	}
	return nil
}

// RoutingKey is everything routing may look at: category and location, never the description.
type RoutingKey struct {
	Category Category
	Location Location
}

// RoutingKey returns the part of the report that routing uses.
func (r Report) RoutingKey() RoutingKey {
	return RoutingKey{Category: r.Category, Location: r.Location}
}
