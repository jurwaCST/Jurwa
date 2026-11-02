package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Category codes, the same strings as the proto enum names and the database.
type Category string

const (
	Hostel        Category = "HOSTEL"
	AcademicBlock Category = "ACADEMIC_BLOCK"
	Others        Category = "OTHERS"
)

// Role is the caller's role code, the same string as the proto enum name.
type Role string

// SystemAdmin is the only role allowed to change a category.
const SystemAdmin Role = "SYSTEM_ADMIN"

// MaxNameLength is the longest display name the PWA form can show.
const MaxNameLength = 50

var (
	ErrInvalidCategory  = errors.New("invalid category")
	ErrNotFound         = errors.New("category not found")
	ErrInvalidName      = errors.New("invalid display name")
	ErrPermissionDenied = errors.New("only a system administrator can change categories")
	ErrRequiredCategory = errors.New("category cannot be removed, routing depends on it")
)

// Info is one category as shown on the report form.
type Info struct {
	Category     Category
	Name         string
	Description  string
	SortOrder    int
	Active       bool
	HeadPosition string // position code in department-directory, for example "DSA"
}

// DefaultCategories are the three agreed categories in display order (decision 2).
func DefaultCategories() []Info {
	return []Info{
		{Category: Hostel, Name: "Hostel", Description: "Rooms, bathrooms, water and electricity in the hostels", SortOrder: 1, Active: true, HeadPosition: "DSA"},
		{Category: AcademicBlock, Name: "Academic Block", Description: "Classrooms, labs and equipment in the academic blocks", SortOrder: 2, Active: true, HeadPosition: "DAA"},
		{Category: Others, Name: "Others", Description: "Campus-wide issues such as Wi-Fi, roads and grounds", SortOrder: 3, Active: true, HeadPosition: "SSO"},
	}
}

// Find returns the category c from list. An empty code is invalid; a code that
// is not in the list is not found.
func Find(list []Info, c Category) (Info, error) {
	if c == "" {
		return Info{}, ErrInvalidCategory
	}
	for _, info := range list {
		if info.Category == c {
			return info, nil
		}
	}
	return Info{}, fmt.Errorf("%w: %q", ErrNotFound, c)
}

// Rename returns info with a new display name and, when desc is not empty, a
// new help text. The code, order and Category Head never change, so routing is
// not affected.
func Rename(info Info, name, desc string, role Role) (Info, error) {
	if role != SystemAdmin {
		return Info{}, ErrPermissionDenied
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Info{}, fmt.Errorf("%w: name is required", ErrInvalidName)
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return Info{}, fmt.Errorf("%w: at most %d characters", ErrInvalidName, MaxNameLength)
	}
	info.Name = name
	if d := strings.TrimSpace(desc); d != "" {
		info.Description = d
	}
	return info, nil
}

// CheckRemovable always refuses the three categories, because routing-rules
// and the report form depend on them.
func CheckRemovable(c Category) error {
	if _, err := Find(DefaultCategories(), c); err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", ErrRequiredCategory, c)
}
