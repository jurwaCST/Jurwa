package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidCategory: the category is not Hostel, Academic Block or Others.
	ErrInvalidCategory = errors.New("invalid category")
	// ErrInvalidWindow: the SLA window must be more than 0 hours.
	ErrInvalidWindow = errors.New("window_hours must be greater than 0")
	// ErrNotFound: no SLA window has been set for this category yet.
	ErrNotFound = errors.New("no policy set for this category")
)

// Category of a report. The numbers match jurwa.common.v1.Category,
// but the domain does not import the generated proto code.
type Category int32

const (
	CategoryUnspecified Category = 0
	Hostel              Category = 1
	AcademicBlock       Category = 2
	Others              Category = 3
)

var categoryNames = map[Category]string{
	Hostel:        "HOSTEL",
	AcademicBlock: "ACADEMIC_BLOCK",
	Others:        "OTHERS",
}

// Valid reports whether c is one of the three categories.
func (c Category) Valid() bool {
	_, ok := categoryNames[c]
	return ok
}

// String returns the upper-case code, the same one stored in the database.
func (c Category) String() string {
	if name, ok := categoryNames[c]; ok {
		return name
	}
	return "CATEGORY_UNSPECIFIED"
}

// Policy is the SLA window for one category: how long a report may stay
// unresolved before it is escalated.
type Policy struct {
	Category    Category
	WindowHours int32
	UpdatedBy   string // System Administrator who last set it
	UpdatedAt   time.Time
}

// NewPolicy checks the rules and returns a policy.
func NewPolicy(category Category, windowHours int32, updatedBy string, now time.Time) (Policy, error) {
	if !category.Valid() {
		return Policy{}, fmt.Errorf("%w: %d", ErrInvalidCategory, int32(category))
	}
	if windowHours <= 0 {
		return Policy{}, fmt.Errorf("%w: got %d", ErrInvalidWindow, windowHours)
	}
	return Policy{Category: category, WindowHours: windowHours, UpdatedBy: updatedBy, UpdatedAt: now}, nil
}

// Window returns the SLA window as a duration.
func (p Policy) Window() time.Duration {
	return time.Duration(p.WindowHours) * time.Hour
}
