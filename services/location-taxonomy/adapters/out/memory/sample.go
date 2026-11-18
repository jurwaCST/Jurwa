package memory

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

// SampleEntries is the location list: the real CST buildings, floors, rooms and
// places from docs/design/real-locations.md (SCRUM-146, loaded in SCRUM-150),
// followed by the first sample rows, switched off so old reports that name them
// still make sense. Ids are fixed: a real place gets the same id after every
// restart (see placeID).
func SampleEntries() []domain.Entry {
	entries := realHostels()
	entries = append(entries, realAcademicBlock()...)
	entries = append(entries, realCampusPlaces()...)
	return append(entries, oldSampleRows()...)
}

// hostel describes one hostel from the floor plans. Rooms are numbered by
// floor: the first digit is the floor, so room n on floor f is "Room f0n".
type hostel struct {
	building  string
	floors    int
	firstRoom int      // first room number on every floor (RKB starts at 13)
	lastRoom  int      // last room number on every floor
	extra     []string // more rooms per floor, "x" is replaced by the floor number
}

// The 9 hostels (docs/design/real-locations.md section 3). HE is handled in
// realHostels because its two floors have different room counts.
var hostels = []hostel{
	{building: "HA", floors: 3, firstRoom: 1, lastRoom: 13},
	{building: "HB", floors: 3, firstRoom: 1, lastRoom: 13},
	{building: "HC", floors: 3, firstRoom: 1, lastRoom: 13},
	{building: "HD", floors: 3, firstRoom: 1, lastRoom: 13},
	{building: "HE", floors: 2},
	{building: "HF", floors: 3, firstRoom: 1, lastRoom: 24},
	{building: "NK", floors: 4, firstRoom: 1, lastRoom: 8, extra: []string{"Kx01", "Kx02"}},
	{building: "RKA", floors: 4, firstRoom: 1, lastRoom: 12},
	{building: "RKB", floors: 4, firstRoom: 13, lastRoom: 24},
}

// heRooms is the last room number on each HE floor: floor 1 has 101 to 112,
// floor 2 has 201 to 233.
var heRooms = map[int]int{1: 12, 2: 33}

// Every hostel floor is split at the middle entrance into a left and a right
// wing, each with a hallway and a bathroom.
var sharedFloorAreas = []string{"Left wing hallway", "Right wing hallway", "Left wing bathroom", "Right wing bathroom"}

func realHostels() []domain.Entry {
	var out []domain.Entry
	for _, h := range hostels {
		for f := 1; f <= h.floors; f++ {
			floor := fmt.Sprintf("Floor %d", f)
			first, last := h.firstRoom, h.lastRoom
			if h.building == "HE" {
				first, last = 1, heRooms[f]
			}
			var areas []string
			for n := first; n <= last; n++ {
				areas = append(areas, fmt.Sprintf("Room %d%02d", f, n))
			}
			for _, x := range h.extra {
				areas = append(areas, "Room "+strings.ReplaceAll(x, "x", fmt.Sprint(f)))
			}
			areas = append(areas, sharedFloorAreas...)
			for _, a := range areas {
				out = append(out, place(domain.CategoryHostel, domain.Location{Building: h.building, Wing: floor, Area: a}))
			}
		}
	}
	return out
}

// The academic rooms are in one building named after the category, without wings.
func realAcademicBlock() []domain.Entry {
	var out []domain.Entry
	for _, prefix := range []string{"CR", "IT", "Lab", "LH"} {
		for n := 1; n <= 5; n++ {
			out = append(out, place(domain.CategoryAcademicBlock, domain.Location{Building: "Academic Block", Area: fmt.Sprintf("%s%d", prefix, n)}))
		}
	}
	return out
}

// Others are places, not services: building is always Campus and there is no wing.
func realCampusPlaces() []domain.Entry {
	var out []domain.Entry
	for _, a := range []string{
		"Library", "Mess", "Football ground", "Basketball court (indoor)",
		"Basketball court (outdoor)", "Volleyball court", "Gym",
	} {
		out = append(out, place(domain.CategoryOthers, domain.Location{Building: domain.CampusBuilding, Area: a}))
	}
	return out
}

// The first sample rows, switched off. Their ids stay the same.
func oldSampleRows() []domain.Entry {
	hostel := domain.CategoryHostel
	block := domain.CategoryAcademicBlock
	others := domain.CategoryOthers
	campus := domain.CampusBuilding
	return []domain.Entry{
		{ID: "668753c7-2008-46e2-9691-fdd4f9d58106", Category: hostel, Location: domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}},
		{ID: "8eb717f8-f5f6-489d-a4cd-361c0c034e51", Category: hostel, Location: domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Room 101"}},
		{ID: "69062735-57e0-40d5-9d9b-bef089d902cb", Category: hostel, Location: domain.Location{Building: "Hostel A", Wing: "Wing 2", Area: "Common Room"}},
		{ID: "0803a9df-c7da-4366-983d-30b5eb834c11", Category: hostel, Location: domain.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}},
		{ID: "7e21adbe-8aae-4704-a99d-c928b2235461", Category: hostel, Location: domain.Location{Building: "Hostel B", Wing: "Ground Floor", Area: "Laundry"}},
		{ID: "1bf3061b-a6a4-41c2-9880-9b810062d4e2", Category: block, Location: domain.Location{Building: "Academic Block 1", Wing: "East Wing", Area: "Lecture Hall 1"}},
		{ID: "e1191761-a50d-4016-ae5b-10e4da7187bd", Category: block, Location: domain.Location{Building: "Academic Block 1", Wing: "West Wing", Area: "Lab 1"}},
		{ID: "35c4a8b4-55fb-4403-b790-8233927f2723", Category: block, Location: domain.Location{Building: "Academic Block 2", Wing: "Ground Floor", Area: "Lab 3"}},
		{ID: "e8c1a92b-182f-49eb-986c-7bcd78ed48bb", Category: block, Location: domain.Location{Building: "Academic Block 2", Wing: "First Floor", Area: "Classroom 204"}},
		{ID: "3da5b988-e449-425a-91e2-9b986801ae94", Category: others, Location: domain.Location{Building: campus, Area: "Wi-Fi"}},
		{ID: "5ba787d7-b197-4772-bc88-67232b77cfe4", Category: others, Location: domain.Location{Building: campus, Area: "Estate"}},
		{ID: "5ff22d8f-fd94-45c5-9aca-b8fa900ab827", Category: others, Location: domain.Location{Building: campus, Area: "Canteen"}},
		{ID: "3f8a0bc9-8455-411f-bb43-2562b123fdb7", Category: others, Location: domain.Location{Building: campus, Area: "Old Workshop"}},
	}
}

func place(c domain.Category, loc domain.Location) domain.Entry {
	return domain.Entry{ID: placeID(c, loc), Category: c, Location: loc, Active: true}
}

// locationNamespace is the fixed namespace for placeID (a random UUID made once).
var locationNamespace = [16]byte{0x6b, 0x2f, 0x1c, 0x4e, 0x93, 0xa7, 0x4d, 0x15, 0x8e, 0x0c, 0x52, 0xd1, 0x7a, 0x3b, 0x9f, 0x60}

// placeID is a name-based UUID (version 5, RFC 9562) of the category and the
// location, so a place keeps its id after a restart and two places never share one.
func placeID(c domain.Category, loc domain.Location) string {
	h := sha1.New()
	h.Write(locationNamespace[:])
	h.Write([]byte(strings.Join([]string{string(c), loc.Building, loc.Wing, loc.Area}, "|")))
	b := h.Sum(nil)[:16]
	b[6] = (b[6] & 0x0f) | 0x50 // version 5
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 9562 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
