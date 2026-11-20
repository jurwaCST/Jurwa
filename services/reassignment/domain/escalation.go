package domain

import (
	"errors"
	"strings"
	"time"
)

// ErrNotEscalated: there is no escalation to undo (the current assignment is
// not ESCALATED, for example because it was already undone).
var ErrNotEscalated = errors.New("current assignment is not an escalation")

// CategoryHead returns the first ROUTED assignment of a report's history: the
// Category Head that routing chose. Escalation hands the report back to them.
func CategoryHead(history []Assignment) (Assignment, error) {
	for _, a := range history {
		if a.Kind == Routed {
			return a, nil
		}
	}
	return Assignment{}, ErrNotFound
}

// EscalationPosition is the position a late report goes up to, by the position
// routing gave it to (SCRUM-158): a report the SSO gave out goes to the DSA, one
// an HoD holds goes to the DAA. "" means nobody is above it, so the report goes
// back to the routed head.
func EscalationPosition(routedTo string) string {
	switch {
	case routedTo == "SSO":
		return "DSA"
	case strings.HasPrefix(routedTo, "HOD_"):
		return "DAA"
	default:
		return ""
	}
}

// NewEscalated is the Saga step after an SLA lapse: the report goes up to head,
// the person it is escalated to. The system acts, so AssignedBy is empty. If
// head already holds the report, the escalation is still recorded, so the
// history and the report.reassigned event show it.
func NewEscalated(id string, head Assignment, now time.Time) Assignment {
	to := "the Category Head"
	if head.AssigneePositionCode != "" {
		to = "the " + head.AssigneePositionCode
	}
	return Assignment{
		ID:                   id,
		ReportID:             head.ReportID,
		AssigneeUserID:       head.AssigneeUserID,
		AssigneePositionCode: head.AssigneePositionCode,
		Kind:                 Escalated,
		Reason:               "SLA lapsed: escalated to " + to,
		Current:              true,
		AssignedAt:           now,
	}
}

// NewCompensated undoes an escalation: the report goes back to previous, the
// assignment before the escalation.
func NewCompensated(id string, escalated, previous Assignment, reason string, now time.Time) (Assignment, error) {
	if escalated.Kind != Escalated {
		return Assignment{}, ErrNotEscalated
	}
	return Assignment{
		ID:                   id,
		ReportID:             escalated.ReportID,
		AssigneeUserID:       previous.AssigneeUserID,
		AssigneePositionCode: previous.AssigneePositionCode,
		Kind:                 Compensated,
		Reason:               "Escalation undone: " + reason,
		Current:              true,
		AssignedAt:           now,
	}, nil
}
