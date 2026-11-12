# Real CST locations (SCRUM-146)

Owner: Gayley Choden. For Sonam Choki, who loads this list into location-taxonomy (SCRUM-150).

The location list in `services/location-taxonomy/adapters/out/memory/sample.go` is sample data ("Hostel A", "Wing 1", "Lecture Hall 1"). This file holds the real names to replace it. The web app has no names written into it: the report form reads them from the location service, so only the service data changes.

**Status: waiting for the names.** Section 3 is filled in from the college and the campus; no name is entered here unless it is checked.

## 1. Categories and heads

Checked against `services/category-taxonomy/domain/category.go` on 12 November 2026:

| Category | Covers | Category Head | Still right? |
|---|---|---|---|
| Hostel | Rooms, bathrooms, water and electricity in the hostels | DSA | |
| Academic Block | Classrooms, labs and equipment in the academic blocks | DAA | |
| Others | Campus-wide issues such as Wi-Fi, roads and grounds | SSO | |

## 2. Rules for the names (from the pilot)

From `docs/testing/pilot-usability-report.md`, problems 1, 2 and 5:

1. **Use the names students already use.** Room "HA-101", not "Room 101"; lecture theatre "LT-01", not "Lecture Hall 1" (P2, P3).
2. **A wing only when students know it.** "Wing 1 / Wing 2" and "East / West Wing" confused students (P1, P4). If a building has no wing or floor that students use, leave the wing empty.
3. **Name every canteen.** CST has 3 canteens; "Canteen" alone is not enough (P5).
4. **Others are places, not services.** "Wi-Fi" and "Estate" were unclear as places (P5). List the place (mess, library, gate, football ground); a Wi-Fi problem is reported at the place where it happens.

What the service accepts (`services/location-taxonomy/domain/location.go`):
- Building is required; wing and area are optional.
- Others have no wing: their building is always `Campus`, and each area is one place.
- A name that is no longer used is kept and set inactive, not deleted, so old reports still show it.

## 3. The list

| Category | Building | Wing or floor | Area |
|---|---|---|---|
| Hostel | | | |
| Academic Block | | | |
| Others | Campus | (none) | |

## 4. Sources

| What | Who or where | Date |
|---|---|---|
| | | |
