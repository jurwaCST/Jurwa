package domain

import (
	"errors"
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

// NewEscalated is the Saga step after an SLA lapse: the report goes back up to
// the Category Head. The system acts, so AssignedBy is empty. If the head
// already holds the report, the escalation is still recorded, so the history
// and the report.reassigned event show it.
func NewEscalated(id string, head Assignment, now time.Time) Assignment {
	return Assignment{
		ID:                   id,
		ReportID:             head.ReportID,
		AssigneeUserID:       head.AssigneeUserID,
		AssigneePositionCode: head.AssigneePositionCode,
		Kind:                 Escalated,
		Reason:               "SLA lapsed: escalated to the Category Head",
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
