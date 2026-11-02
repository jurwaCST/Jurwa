package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Category is the report category. The values match the codes stored in the databases.
type Category string

const (
	CategoryHostel        Category = "HOSTEL"
	CategoryAcademicBlock Category = "ACADEMIC_BLOCK"
	CategoryOthers        Category = "OTHERS"
)

// Status is a report status. "" in a timeline entry means "not a status change".
type Status string

const (
	StatusOpen       Status = "OPEN"
	StatusAssigned   Status = "ASSIGNED"
	StatusInProgress Status = "IN_PROGRESS"
	StatusResolved   Status = "RESOLVED"
	StatusEscalated  Status = "ESCALATED"
)

// Event types written in the timeline. They match the stream names in
// proto/events/v1/events.proto without the "events:" prefix.
const (
	EventReportSubmitted = "report.submitted"
	EventStatusChanged   = "status.changed"
)

// Location says where the problem is. Building is required; wing and area are "" when not used.
type Location struct {
	Building string
	Wing     string
	Area     string
}

// Summary is the read model of one report, used for "my reports" lists.
// AssigneeID is "" until the report is routed.
type Summary struct {
	ReportID   string
	ReporterID string
	Category   Category
	Location   Location
	Status     Status
	AssigneeID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Entry is one line of a report's timeline. ActorID is "" when the system acted.
type Entry struct {
	ID         string
	ReportID   string
	EventType  string
	FromStatus Status
	ToStatus   Status
	ActorID    string
	Summary    string
	OccurredAt time.Time
}

// EntryInput is what an event handler passes in to record one timeline entry.
type EntryInput struct {
	ReportID   string
	EventType  string
	FromStatus Status
	ToStatus   Status
	ActorID    string
	Summary    string
	OccurredAt time.Time
}

// Page sizes for ListReportsByReporter, from the proto contract.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ErrInvalidInput is wrapped by every rule error, so callers can map all of
// them to one response (INVALID_ARGUMENT) with errors.Is.
var ErrInvalidInput = errors.New("invalid input")

var (
	ErrInvalidReportID   = fmt.Errorf("%w: report id must be a UUID", ErrInvalidInput)
	ErrInvalidReporterID = fmt.Errorf("%w: reporter id must be a UUID", ErrInvalidInput)
	ErrInvalidEntryID    = fmt.Errorf("%w: entry id must be a UUID", ErrInvalidInput)
	ErrInvalidActorID    = fmt.Errorf("%w: actor id must be a UUID or empty", ErrInvalidInput)
	ErrInvalidCategory   = fmt.Errorf("%w: category must be HOSTEL, ACADEMIC_BLOCK or OTHERS", ErrInvalidInput)
	ErrInvalidStatus     = fmt.Errorf("%w: unknown status", ErrInvalidInput)
	ErrMissingBuilding   = fmt.Errorf("%w: location building is required", ErrInvalidInput)
	ErrMissingEventType  = fmt.Errorf("%w: event type is required", ErrInvalidInput)
	ErrMissingSummary    = fmt.Errorf("%w: summary is required", ErrInvalidInput)
	ErrMissingTime       = fmt.Errorf("%w: time is required", ErrInvalidInput)
	ErrInvalidPageSize   = fmt.Errorf("%w: page size must be 1 to 100", ErrInvalidInput)
	ErrInvalidPageToken  = fmt.Errorf("%w: page token is not valid", ErrInvalidInput)
)

var (
	// ErrReportNotFound is returned when no report has the requested id.
	ErrReportNotFound = errors.New("report not found")
	// ErrReportExists is returned when a report is recorded twice.
	ErrReportExists = errors.New("report already recorded")
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s has the canonical UUID form.
func IsUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// NewSummary checks the input and returns the summary of a new OPEN report.
func NewSummary(reportID, reporterID string, category Category, loc Location, createdAt time.Time) (Summary, error) {
	switch {
	case !IsUUID(reportID):
		return Summary{}, ErrInvalidReportID
	case !IsUUID(reporterID):
		return Summary{}, ErrInvalidReporterID
	case !validCategory(category):
		return Summary{}, ErrInvalidCategory
	case strings.TrimSpace(loc.Building) == "":
		return Summary{}, ErrMissingBuilding
	case createdAt.IsZero():
		return Summary{}, ErrMissingTime
	}
	return Summary{
		ReportID:   reportID,
		ReporterID: reporterID,
		Category:   category,
		Location:   loc,
		Status:     StatusOpen,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}, nil
}

// NewEntry checks the input and returns a timeline entry.
func NewEntry(id string, in EntryInput) (Entry, error) {
	switch {
	case !IsUUID(id):
		return Entry{}, ErrInvalidEntryID
	case !IsUUID(in.ReportID):
		return Entry{}, ErrInvalidReportID
	case strings.TrimSpace(in.EventType) == "":
		return Entry{}, ErrMissingEventType
	case strings.TrimSpace(in.Summary) == "":
		return Entry{}, ErrMissingSummary
	case in.OccurredAt.IsZero():
		return Entry{}, ErrMissingTime
	case in.ActorID != "" && !IsUUID(in.ActorID):
		return Entry{}, ErrInvalidActorID
	case !validStatusOrNone(in.FromStatus), !validStatusOrNone(in.ToStatus):
		return Entry{}, ErrInvalidStatus
	}
	return Entry{
		ID:         id,
		ReportID:   in.ReportID,
		EventType:  strings.TrimSpace(in.EventType),
		FromStatus: in.FromStatus,
		ToStatus:   in.ToStatus,
		ActorID:    in.ActorID,
		Summary:    strings.TrimSpace(in.Summary),
		OccurredAt: in.OccurredAt,
	}, nil
}

// Apply returns the summary after entry e. Only an entry newer than the
// summary changes it, so an event that arrives late never moves the status back.
func (s Summary) Apply(e Entry) Summary {
	if !e.OccurredAt.After(s.UpdatedAt) {
		return s
	}
	if e.ToStatus != "" {
		s.Status = e.ToStatus
	}
	s.UpdatedAt = e.OccurredAt
	return s
}

// SortTimeline orders entries oldest first. Entries with the same time keep
// the order in which they were recorded.
func SortTimeline(entries []Entry) {
	slices.SortStableFunc(entries, func(a, b Entry) int { return a.OccurredAt.Compare(b.OccurredAt) })
}

// SortNewestFirst orders summaries by creation time, newest first.
func SortNewestFirst(summaries []Summary) {
	slices.SortStableFunc(summaries, func(a, b Summary) int { return b.CreatedAt.Compare(a.CreatedAt) })
}

// FilterByReporter returns only the summaries of one reporter, in the same order.
func FilterByReporter(summaries []Summary, reporterID string) []Summary {
	out := make([]Summary, 0, len(summaries))
	for _, s := range summaries {
		if s.ReporterID == reporterID {
			out = append(out, s)
		}
	}
	return out
}

// Paginate returns one page of items. size 0 means DefaultPageSize. The token
// is the offset of the first item, as returned in next by the previous call;
// next is "" on the last page.
func Paginate(items []Summary, size int32, token string) (page []Summary, next string, err error) {
	if size == 0 {
		size = DefaultPageSize
	}
	if size < 1 || size > MaxPageSize {
		return nil, "", ErrInvalidPageSize
	}
	start := 0
	if token != "" {
		start, err = strconv.Atoi(token)
		if err != nil || start < 0 || start > len(items) {
			return nil, "", ErrInvalidPageToken
		}
	}
	end := min(start+int(size), len(items))
	if end < len(items) {
		next = strconv.Itoa(end)
	}
	return items[start:end], next, nil
}

func validCategory(c Category) bool {
	return c == CategoryHostel || c == CategoryAcademicBlock || c == CategoryOthers
}

func validStatusOrNone(s Status) bool {
	switch s {
	case "", StatusOpen, StatusAssigned, StatusInProgress, StatusResolved, StatusEscalated:
		return true
	default:
		return false
	}
}
