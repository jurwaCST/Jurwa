package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Category codes, the same strings as the proto enum names.
type Category string

const (
	Hostel        Category = "HOSTEL"
	AcademicBlock Category = "ACADEMIC_BLOCK"
	Others        Category = "OTHERS"
)

var (
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidLocation = errors.New("location building is required")
	ErrRuleNotFound    = errors.New("no routing rule matches")
	ErrNoHolder        = errors.New("target position has no holder")
	ErrUnavailable     = errors.New("a service routing depends on is unavailable")
)

// Valid reports whether c is one of the three agreed categories.
func (c Category) Valid() bool {
	switch c {
	case Hostel, AcademicBlock, Others:
		return true
	}
	return false
}

// Request is everything routing may look at: the category and the building.
// There is deliberately no description field (decision 1, FR-11).
type Request struct {
	ReportID string // "" for a dry run: no event is published
	Category Category
	Building string
}

// Validate checks the request before any other service is called.
func (r Request) Validate() error {
	if !r.Category.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidCategory, r.Category)
	}
	if strings.TrimSpace(r.Building) == "" {
		return ErrInvalidLocation
	}
	return nil
}

// Rule is the part of a routing-rules rule that routing needs.
type Rule struct {
	ID             string
	TargetPosition string // position code in department-directory, for example "DSA"
}

// Decision says who receives a report and why.
type Decision struct {
	ReportID       string
	Category       Category
	CategoryHeadID string // user id of the person holding TargetPosition
	TargetPosition string
	RuleID         string
}

// Decide combines the matched rule and the holder of its position.
func Decide(req Request, rule Rule, holderID string) (Decision, error) {
	if holderID == "" {
		return Decision{}, fmt.Errorf("%w: %s", ErrNoHolder, rule.TargetPosition)
	}
	return Decision{
		ReportID:       req.ReportID,
		Category:       req.Category,
		CategoryHeadID: holderID,
		TargetPosition: rule.TargetPosition,
		RuleID:         rule.ID,
	}, nil
}
