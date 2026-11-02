package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Category is the report category. The values match the codes stored in the databases.
type Category string

const (
	CategoryHostel        Category = "HOSTEL"         // hostels: building, wing, area
	CategoryAcademicBlock Category = "ACADEMIC_BLOCK" // academic blocks: building, wing, area
	CategoryOthers        Category = "OTHERS"         // campus-wide: a flat list of areas, no wing
)

// CampusBuilding is the building of every OTHERS entry, so its areas form a flat list.
const CampusBuilding = "Campus"

// Location says where the problem is. Building is required; wing and area are
// optional and "" when not used.
type Location struct {
	Building string
	Wing     string
	Area     string
}

// Entry is one row of the location list, like the locations table in
// docs/design/schema/location-taxonomy.sql. The rows together form the
// building, wing, area hierarchy of each category.
type Entry struct {
	ID       string
	Category Category
	Location Location
	Active   bool // switched-off rows are kept but never listed or matched
}

// Reasons returned by CheckLocation when a location is not valid.
const (
	ReasonUnknownBuilding = "unknown building"
	ReasonUnknownWing     = "unknown wing for this building"
	ReasonUnknownArea     = "unknown area for this building"
)

// ErrInvalidInput is wrapped by every rule error, so callers can map all of
// them to one response (INVALID_ARGUMENT) with errors.Is.
var ErrInvalidInput = errors.New("invalid input")

var (
	ErrInvalidID       = fmt.Errorf("%w: id must be a UUID", ErrInvalidInput)
	ErrMissingCategory = fmt.Errorf("%w: category is required", ErrInvalidInput)
	ErrInvalidCategory = fmt.Errorf("%w: category must be HOSTEL, ACADEMIC_BLOCK or OTHERS", ErrInvalidInput)
	ErrMissingBuilding = fmt.Errorf("%w: location building is required", ErrInvalidInput)
	ErrWingNotAllowed  = fmt.Errorf("%w: campus-wide (OTHERS) locations have no wing", ErrInvalidInput)
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s has the canonical UUID form.
func IsUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// NewEntry checks the hierarchy rules and returns an active entry.
func NewEntry(id string, category Category, loc Location) (Entry, error) {
	if !IsUUID(id) {
		return Entry{}, ErrInvalidID
	}
	if err := ValidateCategory(category); err != nil {
		return Entry{}, err
	}
	if strings.TrimSpace(loc.Building) == "" {
		return Entry{}, ErrMissingBuilding
	}
	if category == CategoryOthers && strings.TrimSpace(loc.Wing) != "" {
		return Entry{}, ErrWingNotAllowed
	}
	return Entry{ID: id, Category: category, Location: loc, Active: true}, nil
}

// ValidateCategory checks that c is one of the three categories.
func ValidateCategory(c Category) error {
	switch c {
	case "":
		return ErrMissingCategory
	case CategoryHostel, CategoryAcademicBlock, CategoryOthers:
		return nil
	default:
		return ErrInvalidCategory
	}
}

// CheckLocation says whether loc exists among the active entries. The caller
// passes the entries of one category. Wing and area are optional, but each one
// given must belong to the building (and the area to the wing). Case and outer
// spaces are ignored. reason is "" when valid. err is set only when loc has no
// building.
func CheckLocation(entries []Entry, loc Location) (valid bool, reason string, err error) {
	want := normalize(loc)
	if want.Building == "" {
		return false, "", ErrMissingBuilding
	}

	var inBuilding []Location
	for _, e := range ActiveOnly(entries) {
		if got := normalize(e.Location); got.Building == want.Building {
			inBuilding = append(inBuilding, got)
		}
	}
	if len(inBuilding) == 0 {
		return false, ReasonUnknownBuilding, nil
	}

	wingFound := want.Wing == ""
	for _, got := range inBuilding {
		if want.Wing != "" && got.Wing == want.Wing {
			wingFound = true
		}
		if matches(got, want) {
			return true, "", nil
		}
	}
	if !wingFound {
		return false, ReasonUnknownWing, nil
	}
	return false, ReasonUnknownArea, nil
}

// matches reports whether the request want fits the stored row got (same building).
// An empty wing in the request means "no wing", unless the area is empty too,
// in which case the building alone is enough.
func matches(got, want Location) bool {
	switch {
	case want.Wing != "" && got.Wing != want.Wing:
		return false
	case want.Wing == "" && want.Area != "" && got.Wing != "":
		return false
	case want.Area != "" && got.Area != want.Area:
		return false
	default:
		return true
	}
}

// ActiveOnly returns the entries that are switched on, in the same order.
func ActiveOnly(entries []Entry) []Entry {
	active := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.Active {
			active = append(active, e)
		}
	}
	return active
}

func normalize(l Location) Location {
	return Location{
		Building: strings.ToLower(strings.TrimSpace(l.Building)),
		Wing:     strings.ToLower(strings.TrimSpace(l.Wing)),
		Area:     strings.ToLower(strings.TrimSpace(l.Area)),
	}
}
