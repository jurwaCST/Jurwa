package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Kind is the kind of position, the same strings as the database check.
type Kind string

const (
	KindCategoryHead Kind = "CATEGORY_HEAD"
	KindHoD          Kind = "HOD"
	KindCouncillor   Kind = "COUNCILLOR"
	KindStaff        Kind = "STAFF"
)

var (
	ErrInvalidCode     = errors.New("code is required")
	ErrInvalidPosition = errors.New("invalid position")
	ErrNotFound        = errors.New("not found")
)

// Department is one of the 6 academic departments, each with one HoD position
// (Mechanical's is held by the acting Programme Leader).
type Department struct {
	Code string
	Name string
}

// Position is a role someone holds, for example "DSA" or "HOD_CT".
type Position struct {
	Code           string
	Title          string
	Kind           Kind
	DepartmentCode string // "" when not a department position
	HolderUserID   string // "" while nobody holds it
	Active         bool
}

// Validate checks the fields an administrator must fill in.
func (p Position) Validate() error {
	if strings.TrimSpace(p.Code) == "" {
		return ErrInvalidCode
	}
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidPosition)
	}
	switch p.Kind {
	case KindCategoryHead, KindCouncillor, KindStaff:
	case KindHoD:
		if p.DepartmentCode == "" {
			return fmt.Errorf("%w: a HoD needs a department", ErrInvalidPosition)
		}
	default:
		return fmt.Errorf("%w: unknown kind %q", ErrInvalidPosition, p.Kind)
	}
	return nil
}

// DefaultDepartments are the 6 academic departments of CST, named as on
// https://cst.edu.bt/department/. Each gets one HoD position. On 2 November 2026
// the staff directory (https://cst.edu.bt/staff-directory/) listed an HoD for
// every department except Mechanical Engineering, so the Mechanical Programme
// Leader acts as HoD (actingHoD). The System Administrator assigns the holder.
func DefaultDepartments() []Department {
	return []Department{
		{Code: "ARCH", Name: "Architecture"},
		{Code: "CIVIL", Name: "Civil & Environmental Engineering"},
		{Code: "CT", Name: "Computing Technologies"},
		{Code: "EEE", Name: "Electrical & Electronics Engineering"},
		{Code: "MECH", Name: "Mechanical Engineering"},
		{Code: "SH", Name: "Science & Humanities"},
	}
}

// actingHoD marks departments with no HoD, whose Programme Leader acts as HoD.
// The position code stays HOD_<dept>, so reassignment works the same way.
var actingHoD = map[string]bool{"MECH": true}

// DefaultPositions are the Category Heads (decision 2), one HoD per department
// and the Hostel reassignment targets named in the proposal. Nobody holds them yet.
func DefaultPositions() []Position {
	ps := []Position{
		{Code: "DSA", Title: "Dean of Student Affairs", Kind: KindCategoryHead},
		{Code: "DAA", Title: "Dean of Academic Affairs", Kind: KindCategoryHead},
		{Code: "SSO", Title: "Student Service Officer", Kind: KindCategoryHead},
		{Code: "SSO_MALE", Title: "Student Service Officer (male)", Kind: KindStaff},
		{Code: "SSO_FEMALE", Title: "Student Service Officer (female)", Kind: KindStaff},
		{Code: "BLOCK_COUNCILLOR", Title: "Block Councillor", Kind: KindCouncillor},
		{Code: "MAINTENANCE_COUNCILLOR", Title: "Maintenance Councillor", Kind: KindCouncillor},
		{Code: "DEPUTY_CHIEF_COUNCILLOR_MESS", Title: "Deputy Chief Councillor (Mess)", Kind: KindCouncillor},
		{Code: "CHIEF_COUNCILLOR", Title: "Chief Councillor", Kind: KindCouncillor},
	}
	for _, d := range DefaultDepartments() {
		title := "Head of Department, " + d.Name
		if actingHoD[d.Code] {
			title += " (Programme Leader, acting)"
		}
		ps = append(ps, Position{Code: "HOD_" + d.Code, Title: title, Kind: KindHoD, DepartmentCode: d.Code})
	}
	for i := range ps {
		ps[i].Active = true
	}
	return ps
}

// demoHolderIDs are the demo accounts that identity seeds (SEED_DEMO_USERS): the
// three Category Heads, so a report can be routed end to end, and the staff
// account, so a report can be reassigned to someone who can log in. Not real people.
var demoHolderIDs = map[string]string{
	"DSA":      "00000000-0000-4000-8000-000000000001",
	"DAA":      "00000000-0000-4000-8000-000000000002",
	"SSO":      "00000000-0000-4000-8000-000000000003",
	"SSO_MALE": "00000000-0000-4000-8000-000000000010",
}

// WithDemoHolders returns a copy of positions where the Category Heads have demo holders.
func WithDemoHolders(positions []Position) []Position {
	out := make([]Position, len(positions))
	copy(out, positions)
	for i := range out {
		if id, ok := demoHolderIDs[out[i].Code]; ok {
			out[i].HolderUserID = id
		}
	}
	return out
}

// HeldPositions returns the positions that have a holder, in the same order:
// the people a report can be reassigned to.
func HeldPositions(positions []Position) []Position {
	out := []Position{}
	for _, p := range positions {
		if p.HolderUserID != "" {
			out = append(out, p)
		}
	}
	return out
}

// normalize trims a code and makes it upper case, so " dsa " finds "DSA".
func normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// FindPosition returns the position with the given code.
func FindPosition(positions []Position, code string) (Position, error) {
	code = normalize(code)
	if code == "" {
		return Position{}, ErrInvalidCode
	}
	for _, p := range positions {
		if p.Code == code {
			return p, nil
		}
	}
	return Position{}, fmt.Errorf("%w: position %q", ErrNotFound, code)
}

// FindHoD returns the Head of Department position of one department.
func FindHoD(depts []Department, positions []Position, deptCode string) (Position, error) {
	deptCode = normalize(deptCode)
	if deptCode == "" {
		return Position{}, ErrInvalidCode
	}
	known := false
	for _, d := range depts {
		if d.Code == deptCode {
			known = true
			break
		}
	}
	if !known {
		return Position{}, fmt.Errorf("%w: department %q", ErrNotFound, deptCode)
	}
	for _, p := range positions {
		if p.Kind == KindHoD && p.DepartmentCode == deptCode {
			return p, nil
		}
	}
	return Position{}, fmt.Errorf("%w: HoD of %q", ErrNotFound, deptCode)
}
