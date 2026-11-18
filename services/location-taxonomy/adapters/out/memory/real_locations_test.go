package memory

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

// The real CST locations from docs/design/real-locations.md (SCRUM-146, SCRUM-150).

// activeIn returns the active entries of one category.
func activeIn(c domain.Category) []domain.Entry {
	var out []domain.Entry
	for _, e := range domain.ActiveOnly(SampleEntries()) {
		if e.Category == c {
			out = append(out, e)
		}
	}
	return out
}

// areasOn returns the areas of one hostel floor, sorted.
func areasOn(building, floor string) []string {
	var areas []string
	for _, e := range activeIn(domain.CategoryHostel) {
		if e.Location.Building == building && e.Location.Wing == floor {
			areas = append(areas, e.Location.Area)
		}
	}
	sort.Strings(areas)
	return areas
}

// rooms returns "Room <prefix><n>" for n from first to last, with n as two digits.
func rooms(prefix string, first, last int) []string {
	var out []string
	for n := first; n <= last; n++ {
		out = append(out, fmt.Sprintf("Room %s%02d", prefix, n))
	}
	return out
}

var sharedAreas = []string{"Left wing bathroom", "Left wing hallway", "Right wing bathroom", "Right wing hallway"}

func withShared(roomList ...[]string) []string {
	var out []string
	for _, r := range roomList {
		out = append(out, r...)
	}
	out = append(out, sharedAreas...)
	sort.Strings(out)
	return out
}

func TestRealLocationCounts(t *testing.T) {
	assert.Len(t, activeIn(domain.CategoryHostel), 525, "9 hostels: rooms plus 4 hallway and bathroom areas per floor")
	assert.Len(t, activeIn(domain.CategoryAcademicBlock), 20, "CR1-5, IT1-5, Lab1-5, LH1-5")
	assert.Len(t, activeIn(domain.CategoryOthers), 7, "7 campus places")
}

func TestHostelBuildingsAndFloors(t *testing.T) {
	floors := map[string]map[string]bool{}
	for _, e := range activeIn(domain.CategoryHostel) {
		if floors[e.Location.Building] == nil {
			floors[e.Location.Building] = map[string]bool{}
		}
		floors[e.Location.Building][e.Location.Wing] = true
	}

	want := map[string]int{"HA": 3, "HB": 3, "HC": 3, "HD": 3, "HE": 2, "HF": 3, "NK": 4, "RKA": 4, "RKB": 4}
	assert.Len(t, floors, len(want), "exactly the 9 hostels")
	for building, n := range want {
		require.Contains(t, floors, building)
		assert.Len(t, floors[building], n, "%s has %d floors", building, n)
		for f := 1; f <= n; f++ {
			assert.True(t, floors[building][fmt.Sprintf("Floor %d", f)], "%s has Floor %d", building, f)
		}
	}
}

func TestHostelRoomsPerFloor(t *testing.T) {
	tests := []struct {
		building, floor string
		want            []string
	}{
		{"HA", "Floor 1", withShared(rooms("1", 1, 13))},
		{"HC", "Floor 2", withShared(rooms("2", 1, 13))},
		{"HD", "Floor 3", withShared(rooms("3", 1, 13))},
		{"HE", "Floor 1", withShared(rooms("1", 1, 12))},
		{"HE", "Floor 2", withShared(rooms("2", 1, 33))},
		{"HF", "Floor 3", withShared(rooms("3", 1, 24))},
		{"NK", "Floor 3", withShared(rooms("3", 1, 8), []string{"Room K301", "Room K302"})},
		{"RKA", "Floor 1", withShared(rooms("1", 1, 12))},
		{"RKB", "Floor 3", withShared(rooms("3", 13, 24))},
		{"RKB", "Floor 4", withShared(rooms("4", 13, 24))},
	}
	for _, tt := range tests {
		t.Run(tt.building+" "+tt.floor, func(t *testing.T) {
			assert.Equal(t, tt.want, areasOn(tt.building, tt.floor))
		})
	}
}

func TestAcademicBlockIsOneBuildingWithoutWings(t *testing.T) {
	var areas []string
	for _, e := range activeIn(domain.CategoryAcademicBlock) {
		assert.Equal(t, "Academic Block", e.Location.Building)
		assert.Empty(t, e.Location.Wing, "academic rooms have no wing")
		areas = append(areas, e.Location.Area)
	}
	var want []string
	for _, prefix := range []string{"CR", "IT", "Lab", "LH"} {
		for n := 1; n <= 5; n++ {
			want = append(want, fmt.Sprintf("%s%d", prefix, n))
		}
	}
	assert.ElementsMatch(t, want, areas)
}

func TestOthersAreTheSevenCampusPlaces(t *testing.T) {
	var areas []string
	for _, e := range activeIn(domain.CategoryOthers) {
		assert.Equal(t, domain.CampusBuilding, e.Location.Building)
		assert.Empty(t, e.Location.Wing)
		areas = append(areas, e.Location.Area)
	}
	assert.ElementsMatch(t, []string{
		"Library", "Mess", "Football ground", "Basketball court (indoor)",
		"Basketball court (outdoor)", "Volleyball court", "Gym",
	}, areas, "no canteens, Wi-Fi or Estate (pilot problems 2 and 5)")
}

// The old sample rows stay in the list, switched off, with their ids, so old
// reports that name them still make sense.
func TestOldSampleRowsAreKeptSwitchedOff(t *testing.T) {
	old := map[string]string{
		"668753c7-2008-46e2-9691-fdd4f9d58106": "Hostel A",
		"8eb717f8-f5f6-489d-a4cd-361c0c034e51": "Hostel A",
		"69062735-57e0-40d5-9d9b-bef089d902cb": "Hostel A",
		"0803a9df-c7da-4366-983d-30b5eb834c11": "Hostel B",
		"7e21adbe-8aae-4704-a99d-c928b2235461": "Hostel B",
		"1bf3061b-a6a4-41c2-9880-9b810062d4e2": "Academic Block 1",
		"e1191761-a50d-4016-ae5b-10e4da7187bd": "Academic Block 1",
		"35c4a8b4-55fb-4403-b790-8233927f2723": "Academic Block 2",
		"e8c1a92b-182f-49eb-986c-7bcd78ed48bb": "Academic Block 2",
		"3da5b988-e449-425a-91e2-9b986801ae94": domain.CampusBuilding,
		"5ba787d7-b197-4772-bc88-67232b77cfe4": domain.CampusBuilding,
		"5ff22d8f-fd94-45c5-9aca-b8fa900ab827": domain.CampusBuilding,
		"3f8a0bc9-8455-411f-bb43-2562b123fdb7": domain.CampusBuilding,
	}
	byID := map[string]domain.Entry{}
	for _, e := range SampleEntries() {
		byID[e.ID] = e
	}
	for id, building := range old {
		e, ok := byID[id]
		require.True(t, ok, "old row %s is kept", id)
		assert.Equal(t, building, e.Location.Building, "old row %s keeps its place", id)
		assert.False(t, e.Active, "old row %s is switched off", id)
	}
}

// Ids are fixed: the same place gets the same id after every restart, and no
// two places share one.
func TestRealLocationIDsAreStableAndUnique(t *testing.T) {
	first, second := SampleEntries(), SampleEntries()
	require.Equal(t, len(first), len(second))

	seen := map[string]domain.Location{}
	for i, e := range first {
		assert.True(t, domain.IsUUID(e.ID), "id %q is a UUID", e.ID)
		assert.Equal(t, e.ID, second[i].ID, "id of %v is the same on every call", e.Location)
		if prev, dup := seen[e.ID]; dup {
			t.Errorf("id %s is used by %v and %v", e.ID, prev, e.Location)
		}
		seen[e.ID] = e.Location
	}
}
