package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Category codes, the same strings as the proto enum names and the database.
type Category string

const (
	Hostel        Category = "HOSTEL"
	AcademicBlock Category = "ACADEMIC_BLOCK"
	Others        Category = "OTHERS"
)

// DefaultPriority is used when a rule has no priority. Lower numbers win.
const DefaultPriority = 100

var (
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidRule     = errors.New("invalid rule")
	ErrRuleNotFound    = errors.New("no routing rule matches")
)

// Valid reports whether c is one of the three agreed categories.
func (c Category) Valid() bool {
	switch c {
	case Hostel, AcademicBlock, Others:
		return true
	}
	return false
}

// Rule sends reports of a category, optionally only for one building, to a
// Category Head position in department-directory.
type Rule struct {
	ID             string
	Category       Category
	Building       string // "" means any building in the category
	TargetPosition string // position code, for example "DSA"
	Priority       int    // lower wins when two rules are equally specific
	Active         bool
	UpdatedAt      time.Time
}

// Validate checks the fields an administrator must fill in.
func (r Rule) Validate() error {
	if !r.Category.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidCategory, r.Category)
	}
	if strings.TrimSpace(r.TargetPosition) == "" {
		return fmt.Errorf("%w: target position is required", ErrInvalidRule)
	}
	if r.Priority < 0 {
		return fmt.Errorf("%w: priority must not be negative", ErrInvalidRule)
	}
	return nil
}

// DefaultRules are the three agreed rules (decision 2): Hostel to DSA,
// Academic Block to DAA, Others to SSO.
func DefaultRules() []Rule {
	return []Rule{
		{Category: Hostel, TargetPosition: "DSA", Priority: DefaultPriority, Active: true},
		{Category: AcademicBlock, TargetPosition: "DAA", Priority: DefaultPriority, Active: true},
		{Category: Others, TargetPosition: "SSO", Priority: DefaultPriority, Active: true},
	}
}
