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
	CategoryHostel        Category = "HOSTEL"
	CategoryAcademicBlock Category = "ACADEMIC_BLOCK"
	CategoryOthers        Category = "OTHERS"
)

// Location says where the problem is. Building is required; wing and area are
// optional and "" when not used.
type Location struct {
	Building string
	Wing     string
	Area     string
}

// Entry is one location students can pick for a category.
type Entry struct {
	ID       string
	Category Category
	Location Location
	Active   bool // switched-off entries are kept but never listed or matched
}

// Reasons returned by CheckLocation when a location is not valid.
const (
	ReasonUnknownBuilding   = "unknown building"
	ReasonUnknownWingOrArea = "unknown wing or area for this building"
)

// ErrInvalidQuery is wrapped by every request rule error, so callers can map all
// of them to one response (INVALID_ARGUMENT) with errors.Is.
var ErrInvalidQuery = errors.New("invalid query")

var (
	ErrMissingCategory = fmt.Errorf("%w: category is required", ErrInvalidQuery)
	ErrInvalidCategory = fmt.Errorf("%w: category must be HOSTEL, ACADEMIC_BLOCK or OTHERS", ErrInvalidQuery)
	ErrMissingBuilding = fmt.Errorf("%w: location building is required", ErrInvalidQuery)
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s has the canonical UUID form.
func IsUUID(s string) bool {
	return uuidPattern.MatchString(s)
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

// CheckLocation says whether loc exactly matches one of the active entries.
// The caller passes the entries of one category. Case and outer spaces are
// ignored. reason is "" when valid. err is set only when loc has no building.
func CheckLocation(entries []Entry, loc Location) (valid bool, reason string, err error) {
	if strings.TrimSpace(loc.Building) == "" {
		return false, "", ErrMissingBuilding
	}
	want := normalize(loc)
	knownBuilding := false
	for _, e := range ActiveOnly(entries) {
		got := normalize(e.Location)
		if got.Building != want.Building {
			continue
		}
		knownBuilding = true
		if got == want {
			return true, "", nil
		}
	}
	if knownBuilding {
		return false, ReasonUnknownWingOrArea, nil
	}
	return false, ReasonUnknownBuilding, nil
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
