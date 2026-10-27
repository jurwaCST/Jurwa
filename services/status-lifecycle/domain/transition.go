package domain

import (
	"fmt"
	"time"
)

// userTransitions are the changes a person (Category Head or staff) may make.
// The system may make these too.
var userTransitions = map[Status][]Status{
	Open:       {Assigned},
	Assigned:   {InProgress},
	InProgress: {Resolved},
	Escalated:  {InProgress}, // the new assignee starts work on an escalated report
}

// canEscalate are the statuses the system may escalate from when the SLA passes.
var canEscalate = map[Status]bool{Open: true, Assigned: true, InProgress: true}

// CheckTransition returns nil if a report may go from "from" to "to".
// previous is the status before the current one; it is needed to undo an escalation.
// bySystem is true when no user is acting (routing, SLA escalation, Saga compensation).
func CheckTransition(from, to, previous Status, bySystem bool) error {
	if !to.Valid() {
		return fmt.Errorf("%w: %d", ErrInvalidStatus, int32(to))
	}
	for _, allowed := range userTransitions[from] {
		if to == allowed {
			return nil
		}
	}
	if bySystem {
		// SLA escalation.
		if to == Escalated && canEscalate[from] {
			return nil
		}
		// Saga compensation: undo an escalation by going back to the previous status.
		if from == Escalated && to == previous && canEscalate[previous] {
			return nil
		}
	}
	return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, from, to)
}

// ReportStatus is the current status of one report.
type ReportStatus struct {
	ReportID  string
	Status    Status
	Previous  Status // StatusUnspecified until the first change
	Version   int32  // starts at 1, goes up by 1 on every change
	UpdatedBy string // "" when the system made the change
	UpdatedAt time.Time
}

// NewReportStatus returns the status of a newly submitted report: OPEN, version 1.
func NewReportStatus(reportID string, now time.Time) ReportStatus {
	return ReportStatus{ReportID: reportID, Status: Open, Version: 1, UpdatedAt: now}
}

// ChangeTo moves the report to a new status. actorID "" means the system is acting.
// On error the report is left unchanged.
func (r *ReportStatus) ChangeTo(to Status, actorID string, now time.Time) error {
	if err := CheckTransition(r.Status, to, r.Previous, actorID == ""); err != nil {
		return err
	}
	r.Previous = r.Status
	r.Status = to
	r.Version++
	r.UpdatedBy = actorID
	r.UpdatedAt = now
	return nil
}
