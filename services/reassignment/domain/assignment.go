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

// MaxReasonLen is the longest reason allowed, in characters.
const MaxReasonLen = 500

// Errors returned by the domain. Adapters map them to gRPC codes.
var (
	ErrInvalidReason    = errors.New("reason must not be blank and at most 500 characters")
	ErrSameAssignee     = errors.New("report is already assigned to this user")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotFound         = errors.New("report has no assignment")
	ErrAlreadyAssigned  = errors.New("report already has an assignment")
	ErrConflict         = errors.New("assignment changed meanwhile, try again")
)

// Role is the caller's role. The values match the Role enum names in proto/common/v1.
type Role string

const (
	RoleStudent      Role = "STUDENT"
	RoleCategoryHead Role = "CATEGORY_HEAD"
	RoleStaff        Role = "STAFF"
	RoleSystemAdmin  Role = "SYSTEM_ADMIN"
	RolePresident    Role = "PRESIDENT"
)

// listRoles may read the assignment history (docs/design/api-routes.md).
var listRoles = []Role{RoleCategoryHead, RoleStaff, RoleSystemAdmin}

// Kind uses the same numbers as jurwa.reassignment.v1.AssignmentKind.
type Kind int32

const (
	KindUnspecified Kind = 0
	Routed          Kind = 1 // first assignment by routing-engine
	Reassigned      Kind = 2 // a person moved it
	Escalated       Kind = 3 // the SLA lapsed (Saga step)
	Compensated     Kind = 4 // an escalation was undone (Saga compensation)
)

func (k Kind) String() string {
	switch k {
	case Routed:
		return "ROUTED"
	case Reassigned:
		return "REASSIGNED"
	case Escalated:
		return "ESCALATED"
	case Compensated:
		return "COMPENSATED"
	default:
		return "ASSIGNMENT_KIND_UNSPECIFIED"
	}
}

// Assignment is one row of a report's assignment history.
type Assignment struct {
	ID                   string
	ReportID             string
	AssigneeUserID       string
	AssigneePositionCode string // "" when assigned to a person, not a position
	Kind                 Kind
	AssignedBy           string // "" when the system acted
	Reason               string
	Current              bool
	AssignedAt           time.Time
}

// NewRouted is the first assignment, made by the system when routing-engine
// publishes report.routed.
func NewRouted(id, reportID, assigneeUserID, positionCode string, now time.Time) Assignment {
	return Assignment{
		ID:                   id,
		ReportID:             reportID,
		AssigneeUserID:       assigneeUserID,
		AssigneePositionCode: positionCode,
		Kind:                 Routed,
		Current:              true,
		AssignedAt:           now,
	}
}

// CheckCanReassign allows a Category Head, or a staff member who is the current assignee.
func CheckCanReassign(role Role, actorID string, current Assignment) error {
	switch {
	case role == RoleCategoryHead:
		return nil
	case role == RoleStaff && actorID == current.AssigneeUserID:
		return nil
	default:
		return fmt.Errorf("%w: only a Category Head or the current assignee can reassign", ErrPermissionDenied)
	}
}

// CheckCanList allows Category Head, Staff and System Administrator.
func CheckCanList(role Role) error {
	if !slices.Contains(listRoles, role) {
		return fmt.Errorf("%w: role %q cannot view assignments", ErrPermissionDenied, role)
	}
	return nil
}

// Reassign returns the new current assignment that replaces current.
func Reassign(id string, current Assignment, toUserID, reason, actorID string, now time.Time) (Assignment, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" || utf8.RuneCountInString(reason) > MaxReasonLen {
		return Assignment{}, ErrInvalidReason
	}
	if toUserID == current.AssigneeUserID {
		return Assignment{}, ErrSameAssignee
	}
	return Assignment{
		ID:             id,
		ReportID:       current.ReportID,
		AssigneeUserID: toUserID,
		Kind:           Reassigned,
		AssignedBy:     actorID,
		Reason:         reason,
		Current:        true,
		AssignedAt:     now,
	}, nil
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s looks like a UUID.
func IsUUID(s string) bool { return uuidPattern.MatchString(s) }
