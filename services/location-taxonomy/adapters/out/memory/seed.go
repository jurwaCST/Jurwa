package memory

import "github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"

// SeedEntries is the starting location list. The names are examples that match
// the wireframes; replace them with the real CST buildings when the team has them.
// The ids are fixed so they stay the same between restarts.
func SeedEntries() []domain.Entry {
	hostel := domain.CategoryHostel
	block := domain.CategoryAcademicBlock
	others := domain.CategoryOthers
	return []domain.Entry{
		// Hostels (routed to DSA)
		{ID: "668753c7-2008-46e2-9691-fdd4f9d58106", Category: hostel, Location: domain.Location{Building: "Hostel A"}, Active: true},
		{ID: "8eb717f8-f5f6-489d-a4cd-361c0c034e51", Category: hostel, Location: domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}, Active: true},
		{ID: "69062735-57e0-40d5-9d9b-bef089d902cb", Category: hostel, Location: domain.Location{Building: "Hostel B"}, Active: true},
		{ID: "0803a9df-c7da-4366-983d-30b5eb834c11", Category: hostel, Location: domain.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Bathroom 2"}, Active: true},
		{ID: "7e21adbe-8aae-4704-a99d-c928b2235461", Category: hostel, Location: domain.Location{Building: "Hostel B", Wing: "Wing 2", Area: "Common Room"}, Active: true},
		{ID: "d1745b9d-391c-4f80-a697-08dbf3234c10", Category: hostel, Location: domain.Location{Building: "Hostel C", Area: "Laundry"}, Active: true},
		// Academic blocks (routed to DAA)
		{ID: "1bf3061b-a6a4-41c2-9880-9b810062d4e2", Category: block, Location: domain.Location{Building: "Academic Block 1"}, Active: true},
		{ID: "e1191761-a50d-4016-ae5b-10e4da7187bd", Category: block, Location: domain.Location{Building: "Academic Block 1", Area: "Lecture Hall 1"}, Active: true},
		{ID: "35c4a8b4-55fb-4403-b790-8233927f2723", Category: block, Location: domain.Location{Building: "Academic Block 2"}, Active: true},
		{ID: "e8c1a92b-182f-49eb-986c-7bcd78ed48bb", Category: block, Location: domain.Location{Building: "Academic Block 2", Area: "Lab 3"}, Active: true},
		// Campus-wide (routed to SSO)
		{ID: "3da5b988-e449-425a-91e2-9b986801ae94", Category: others, Location: domain.Location{Building: "Library"}, Active: true},
		{ID: "5ba787d7-b197-4772-bc88-67232b77cfe4", Category: others, Location: domain.Location{Building: "Canteen"}, Active: true},
		{ID: "5ff22d8f-fd94-45c5-9aca-b8fa900ab827", Category: others, Location: domain.Location{Building: "Sports Ground"}, Active: true},
		{ID: "3f8a0bc9-8455-411f-bb43-2562b123fdb7", Category: others, Location: domain.Location{Building: "Old Workshop"}, Active: false},
	}
}
