// Package checker does one escalation run (FR-15): it asks sla-tracking for
// timers past their deadline, which sla-tracking marks LAPSED and announces with
// sla.lapsed, then sets each report to ESCALATED in status-lifecycle. It talks
// to other services only through the two interfaces below, so tests use fakes.
package checker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

const (
	// BatchSize is the limit sent to sla-tracking ListDueTimers (its default).
	BatchSize = 100
	// MaxBatches bounds one run; anything left is picked up by the next run.
	MaxBatches = 10
)

// ErrNotEscalatable: status-lifecycle refused the change because the report
// is no longer open for escalation (for example it was resolved meanwhile).
var ErrNotEscalatable = errors.New("report cannot be escalated")

// Lapsed is one timer that sla-tracking has just marked LAPSED.
type Lapsed struct {
	ReportID string
	Deadline time.Time
}

// DueTimers is sla-tracking.
type DueTimers interface {
	// LapseDue marks up to limit due timers LAPSED and returns them.
	LapseDue(ctx context.Context, limit int) ([]Lapsed, error)
}

// StatusChanger is status-lifecycle.
type StatusChanger interface {
	// Escalate sets the report to ESCALATED as the system.
	// It wraps ErrNotEscalatable when the flow does not allow it.
	Escalate(ctx context.Context, reportID, reason string) error
}

// Result is what one run did. It is the function's JSON answer.
type Result struct {
	Lapsed    int `json:"lapsed"`    // timers sla-tracking marked LAPSED
	Escalated int `json:"escalated"` // reports set to ESCALATED
	Skipped   int `json:"skipped"`   // resolved or already escalated meanwhile
	Errors    int `json:"errors"`    // status-lifecycle failed for the report
}

// Checker runs the escalation.
type Checker struct {
	timers DueTimers
	status StatusChanger
}

// New returns a Checker that uses the given services.
func New(timers DueTimers, status StatusChanger) *Checker {
	return &Checker{timers: timers, status: status}
}

// Run does one escalation run. It returns an error only when sla-tracking
// cannot be asked; a failure for one report is counted and the rest continue.
func (c *Checker) Run(ctx context.Context) (Result, error) {
	var res Result
	for batch := 0; batch < MaxBatches; batch++ {
		due, err := c.timers.LapseDue(ctx, BatchSize)
		if err != nil {
			return res, fmt.Errorf("sla-tracking ListDueTimers: %w", err)
		}
		res.Lapsed += len(due)
		for _, l := range due {
			c.escalate(ctx, l, &res)
		}
		if len(due) < BatchSize {
			break
		}
	}
	return res, nil
}

func (c *Checker) escalate(ctx context.Context, l Lapsed, res *Result) {
	reason := "SLA lapsed: deadline was " + l.Deadline.UTC().Format(time.RFC3339)
	err := c.status.Escalate(ctx, l.ReportID, reason)
	switch {
	case err == nil:
		res.Escalated++
		slog.InfoContext(ctx, "report escalated", "report_id", l.ReportID, "deadline", l.Deadline)
	case errors.Is(err, ErrNotEscalatable):
		res.Skipped++
		slog.InfoContext(ctx, "report not escalated", "report_id", l.ReportID, "reason", err.Error())
	default:
		res.Errors++
		slog.ErrorContext(ctx, "escalation failed", "report_id", l.ReportID, "error", err)
	}
}
