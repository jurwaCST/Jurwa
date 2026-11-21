package domain

import (
	"errors"
	"fmt"
	"sort"
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

// DepartmentHoDs maps each CST department, as the student picks it for an
// Academic Block report, to its HoD position in department-directory
// (SCRUM-158). The names are the Academic Block "buildings" in location-taxonomy.
var DepartmentHoDs = map[string]string{
	"Architecture":                         "HOD_ARCH",
	"Civil & Environmental Engineering":    "HOD_CIVIL",
	"Computing Technologies":               "HOD_CT",
	"Electrical & Electronics Engineering": "HOD_EEE",
	"Mechanical Engineering":               "HOD_MECH",
	"Science & Humanities":                 "HOD_SH",
}

// DefaultRules are the agreed rules (SCRUM-158): Hostel and Others go to the
// SSO, who hands them to a councillor; an Academic Block report goes to the HoD
// of the department the student picked. The Academic Block rule without a
// building catches a report that names no department (an old one): the DAA.
func DefaultRules() []Rule {
	rules := []Rule{
		{Category: Hostel, TargetPosition: "SSO", Priority: DefaultPriority, Active: true},
		{Category: AcademicBlock, TargetPosition: "DAA", Priority: DefaultPriority, Active: true},
		{Category: Others, TargetPosition: "SSO", Priority: DefaultPriority, Active: true},
	}
	for _, d := range sortedDepartments() {
		rules = append(rules, Rule{Category: AcademicBlock, Building: d, TargetPosition: DepartmentHoDs[d], Priority: DefaultPriority, Active: true})
	}
	return rules
}

func sortedDepartments() []string {
	out := make([]string, 0, len(DepartmentHoDs))
	for d := range DepartmentHoDs {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}
