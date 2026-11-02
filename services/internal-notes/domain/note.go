package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxBodyChars is the longest note body, in characters (not bytes), after trimming spaces.
const MaxBodyChars = 1000

// Role is the caller's role. The values match the Role enum names in proto/common/v1.
type Role string

const (
	RoleStudent      Role = "STUDENT"
	RoleCategoryHead Role = "CATEGORY_HEAD"
	RoleStaff        Role = "STAFF"
	RoleSystemAdmin  Role = "SYSTEM_ADMIN"
	RolePresident    Role = "PRESIDENT"
)

// Who may use notes, as in docs/design/api-routes.md and the authorization
// policy. Notes are internal: students never see them (FR-21), and the College
// President's dashboard does not show them. System Admin may read but not add.
var (
	addRoles  = []Role{RoleStaff, RoleCategoryHead}
	listRoles = []Role{RoleStaff, RoleCategoryHead, RoleSystemAdmin}
)

// Note is one staff-only note on a report.
type Note struct {
	ID        string
	ReportID  string
	AuthorID  string
	Body      string
	CreatedAt time.Time
}

// NewNoteInput is what a staff user sends. Role comes from the caller's JWT.
type NewNoteInput struct {
	ReportID string
	AuthorID string
	Role     Role
	Body     string
}

// ErrInvalidInput is wrapped by every rule error, so callers can map all of
// them to one response (INVALID_ARGUMENT) with errors.Is.
var ErrInvalidInput = errors.New("invalid input")

var (
	ErrInvalidID       = fmt.Errorf("%w: note id must be a UUID", ErrInvalidInput)
	ErrInvalidReportID = fmt.Errorf("%w: report id must be a UUID", ErrInvalidInput)
	ErrInvalidAuthorID = fmt.Errorf("%w: author id must be a UUID", ErrInvalidInput)
	ErrBlankBody       = fmt.Errorf("%w: note body must not be blank", ErrInvalidInput)
	ErrBodyTooLong     = fmt.Errorf("%w: note body must be at most 1000 characters", ErrInvalidInput)
)

// ErrPermissionDenied is returned when the role may not add or read notes.
var ErrPermissionDenied = errors.New("permission denied")

var (
	errCannotAdd  = fmt.Errorf("%w: only staff and category heads may add internal notes", ErrPermissionDenied)
	errCannotList = fmt.Errorf("%w: only staff, category heads and system administrators may read internal notes", ErrPermissionDenied)
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s has the canonical UUID form.
func IsUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// CheckCanList returns ErrPermissionDenied unless the role may read notes.
func CheckCanList(role Role) error {
	if !slices.Contains(listRoles, role) {
		return errCannotList
	}
	return nil
}

// NewNote checks the role first, then the input, and returns the note.
// The role is checked first so a student learns nothing about the rules.
func NewNote(id string, in NewNoteInput, now time.Time) (Note, error) {
	if !slices.Contains(addRoles, in.Role) {
		return Note{}, errCannotAdd
	}
	body := strings.TrimSpace(in.Body)
	switch {
	case !IsUUID(id):
		return Note{}, ErrInvalidID
	case !IsUUID(in.ReportID):
		return Note{}, ErrInvalidReportID
	case !IsUUID(in.AuthorID):
		return Note{}, ErrInvalidAuthorID
	case body == "":
		return Note{}, ErrBlankBody
	case utf8.RuneCountInString(body) > MaxBodyChars:
		return Note{}, ErrBodyTooLong
	}
	return Note{ID: id, ReportID: in.ReportID, AuthorID: in.AuthorID, Body: body, CreatedAt: now}, nil
}

// SortOldestFirst orders notes by creation time. Notes with the same time keep their order.
func SortOldestFirst(notes []Note) {
	slices.SortStableFunc(notes, func(a, b Note) int { return a.CreatedAt.Compare(b.CreatedAt) })
}

// FilterByReport returns only the notes of one report, in the same order.
func FilterByReport(notes []Note, reportID string) []Note {
	out := make([]Note, 0, len(notes))
	for _, n := range notes {
		if n.ReportID == reportID {
			out = append(out, n)
		}
	}
	return out
}
