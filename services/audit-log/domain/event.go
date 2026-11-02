package domain

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// Limits that keep one entry small. Values are counted in characters.
const (
	MaxServiceLen = 64
	MaxDetails    = 20
	MaxKeyLen     = 64
	MaxValueLen   = 500
)

// Errors returned by the domain. Adapters map them to gRPC codes.
var (
	ErrInvalidEvent     = errors.New("invalid audit event")
	ErrPermissionDenied = errors.New("permission denied")
)

// Action is what happened. The list comes from proto/audit-log/v1.
type Action string

const (
	ReportSubmitted Action = "REPORT_SUBMITTED"
	Routed          Action = "ROUTED"
	StatusChanged   Action = "STATUS_CHANGED"
	Reassigned      Action = "REASSIGNED"
	Escalated       Action = "ESCALATED"
	Compensated     Action = "COMPENSATED"
)

var actions = []Action{ReportSubmitted, Routed, StatusChanged, Reassigned, Escalated, Compensated}

// Role is the caller's role, as the Role enum names in proto/common/v1.
type Role string

// Event is one audit entry. It is never changed or deleted after it is stored.
type Event struct {
	ID         string
	OccurredAt time.Time
	Service    string
	Action     Action
	ReportID   string // "" when the action is not about a report
	ActorID    string // "" when the system acted
	OnBehalfOf string // staff member without email a supervisor acts for (FR-24)
	Details    map[string]string
}

// Input is what a service sends to Append.
type Input struct {
	OccurredAt time.Time // zero means now
	Service    string
	Action     string
	ReportID   string
	ActorID    string
	OnBehalfOf string
	Details    map[string]string
}

// NewEvent checks the input and returns the entry to store.
func NewEvent(id string, in Input, now time.Time) (Event, error) {
	service := strings.TrimSpace(in.Service)
	if service == "" || utf8.RuneCountInString(service) > MaxServiceLen {
		return Event{}, fmt.Errorf("%w: service is required, at most %d characters", ErrInvalidEvent, MaxServiceLen)
	}
	action := Action(in.Action)
	if !slices.Contains(actions, action) {
		return Event{}, fmt.Errorf("%w: unknown action %q", ErrInvalidEvent, in.Action)
	}
	if in.ReportID != "" && !IsUUID(in.ReportID) {
		return Event{}, fmt.Errorf("%w: report_id must be a UUID", ErrInvalidEvent)
	}
	if in.ActorID != "" && !IsUUID(in.ActorID) {
		return Event{}, fmt.Errorf("%w: actor_id must be a UUID", ErrInvalidEvent)
	}
	if utf8.RuneCountInString(in.OnBehalfOf) > MaxValueLen {
		return Event{}, fmt.Errorf("%w: on_behalf_of is longer than %d characters", ErrInvalidEvent, MaxValueLen)
	}
	if err := checkDetails(in.Details); err != nil {
		return Event{}, err
	}

	occurredAt := in.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}
	return Event{
		ID:         id,
		OccurredAt: occurredAt,
		Service:    service,
		Action:     action,
		ReportID:   in.ReportID,
		ActorID:    in.ActorID,
		OnBehalfOf: in.OnBehalfOf,
		Details:    maps.Clone(in.Details),
	}, nil
}

func checkDetails(d map[string]string) error {
	if len(d) > MaxDetails {
		return fmt.Errorf("%w: at most %d details", ErrInvalidEvent, MaxDetails)
	}
	for k, v := range d {
		if k == "" || utf8.RuneCountInString(k) > MaxKeyLen {
			return fmt.Errorf("%w: detail keys must be 1 to %d characters", ErrInvalidEvent, MaxKeyLen)
		}
		if utf8.RuneCountInString(v) > MaxValueLen {
			return fmt.Errorf("%w: detail %q is longer than %d characters", ErrInvalidEvent, k, MaxValueLen)
		}
	}
	return nil
}

// CheckCanList allows only the System Administrator (docs/design/api-routes.md).
func CheckCanList(role Role) error {
	if role != "SYSTEM_ADMIN" {
		return fmt.Errorf("%w: only the System Administrator can read the audit log", ErrPermissionDenied)
	}
	return nil
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s looks like a UUID.
func IsUUID(s string) bool { return uuidPattern.MatchString(s) }
