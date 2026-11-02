package domain

import (
	"errors"
	"fmt"
	"time"
)

// Errors returned by the domain. Adapters map them to gRPC codes.
var (
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidWindow   = errors.New("SLA window must be greater than 0")
	ErrNotFound        = errors.New("timer not found")
	ErrAlreadyExists   = errors.New("timer already exists for this report")
	ErrNotDue          = errors.New("timer is not running or its deadline has not passed")
)

// Category uses the same numbers as jurwa.common.v1.Category.
type Category int32

const (
	CategoryUnspecified Category = 0
	Hostel              Category = 1
	AcademicBlock       Category = 2
	Others              Category = 3
)

// Valid reports whether c is one of the three report categories.
func (c Category) Valid() bool { return c >= Hostel && c <= Others }

func (c Category) String() string {
	switch c {
	case Hostel:
		return "HOSTEL"
	case AcademicBlock:
		return "ACADEMIC_BLOCK"
	case Others:
		return "OTHERS"
	default:
		return "CATEGORY_UNSPECIFIED"
	}
}

// State uses the same numbers as jurwa.slatracking.v1.TimerState.
type State int32

const (
	StateUnspecified State = 0
	Running          State = 1
	Stopped          State = 2
	Lapsed           State = 3
)

func (s State) String() string {
	switch s {
	case Running:
		return "RUNNING"
	case Stopped:
		return "STOPPED"
	case Lapsed:
		return "LAPSED"
	default:
		return "TIMER_STATE_UNSPECIFIED"
	}
}

// Timer is the SLA timer of one report.
type Timer struct {
	ReportID  string
	Category  Category
	StartedAt time.Time
	Deadline  time.Time
	State     State
}

// NewTimer starts a running timer whose deadline is now plus the SLA window.
func NewTimer(reportID string, category Category, window time.Duration, now time.Time) (Timer, error) {
	if !category.Valid() {
		return Timer{}, fmt.Errorf("%w: %d", ErrInvalidCategory, int32(category))
	}
	if window <= 0 {
		return Timer{}, fmt.Errorf("%w: got %s", ErrInvalidWindow, window)
	}
	return Timer{
		ReportID:  reportID,
		Category:  category,
		StartedAt: now,
		Deadline:  now.Add(window),
		State:     Running,
	}, nil
}

// IsDue reports whether the timer is running and its deadline is at or before now.
func (t Timer) IsDue(now time.Time) bool {
	return t.State == Running && !t.Deadline.After(now)
}

// Lapse marks a due timer as LAPSED. A timer lapses only once.
func (t *Timer) Lapse(now time.Time) error {
	if !t.IsDue(now) {
		return fmt.Errorf("%w: state %s, deadline %s", ErrNotDue, t.State, t.Deadline.Format(time.RFC3339))
	}
	t.State = Lapsed
	return nil
}

// Stop stops the timer when the report is resolved. It works from any state,
// so stopping twice or after a lapse is safe.
func (t *Timer) Stop() {
	t.State = Stopped
}
