# Real CST locations (SCRUM-146)

Owner: Gayley Choden. For Sonam Choki, who loads this list into location-taxonomy (SCRUM-150).

The location list in `services/location-taxonomy/adapters/out/memory/sample.go` is sample data ("Hostel A", "Wing 1", "Lecture Hall 1"). This file holds the real names to replace it. The web app has no names written into it: the report form reads them from the location service, so only the service data changes.

## 1. Categories and heads

From `services/category-taxonomy/domain/category.go` on 12 November 2026:

| Category | Covers | Category Head |
|---|---|---|
| Hostel | Rooms, bathrooms, water and electricity in the hostels | DSA |
| Academic Block | Classrooms, labs and equipment in the academic blocks | DAA |
| Others | Campus-wide issues such as Wi-Fi, roads and grounds | SSO |

## 2. Rules for the names (from the pilot)

From `docs/testing/pilot-usability-report.md`, problems 1, 2 and 5:

1. **Use the names students already use:** "HA", "CR1", "LH1", not "Hostel A" or "Lecture Hall 1" (P2, P3).
2. **Wings only where students use them.** "Wing 1 / Wing 2" and "East / West Wing" confused students (P1, P4). Hostels use the floor ("Floor 1", "Floor 2") in the wing field, as the hostel floor plans do; academic rooms have no wing.
3. **No canteens.** Canteens cannot be reported in Jurwa. This replaces the pilot's request to name the 3 canteens (P5).
4. **Others are places, not services.** "Wi-Fi" and "Estate" were unclear as places (P5), so they are removed. A Wi-Fi problem is reported at the place where it happens.

What the service accepts (`services/location-taxonomy/domain/location.go`):
- Building is required; wing and area are optional.
- Others have no wing: their building is always `Campus`, and each area is one place.
- A sample name that is no longer used is set inactive, not deleted, so old reports still show it.

## 3. The list

**Hostel** (9 buildings). From the hostel floor plans (screenshots of each hostel's first floor, and HE's second floor). Rooms are numbered by floor: the first digit is the floor. There is no laundry room.

For every hostel, each floor gets one row per room and one row per restroom area:

- Wing = the floor: `Floor 1`, `Floor 2`, ...
- Area = the room number (`101`) or the restroom area named on the plan (`Restroom`, or `Bathroom` and `Toilet`).

| Building | Floors | Rooms on each floor | Restroom areas on each floor | Rooms per floor |
|---|---|---|---|---|
| HA | 3 | x01 to x13 | Restroom | 13 |
| HB | 3 | x01 to x13 | Restroom | 13 |
| HC | 3 | x01 to x13 (floor 1: 101 to 113) | Restroom | 13 |
| HD | 3 | x01 to x13 | Restroom | 13 |
| HE | 2 | Floor 2: 201 to 233. Floor 1: **waiting for** the plan | **Waiting for** the plan | 33 on floor 2 |
| HF | 3 | x01 to x24 (floor 1: 101 to 124) | Restroom | 24 |
| NK | 4 | x01 to x08, Kx01, Kx02 (floor 1: 101 to 108, K101, K102) | Restroom | 10 |
| RKA | 4 | x01 to x12 (floor 1: 101 to 112) | Bathroom, Toilet | 12 |
| RKB | 4 | x13 to x24 (floor 1: 113 to 124) | Bathroom, Toilet | 12 |

"x" is the floor number, so HC floor 2 is 201 to 213 and RKB floor 3 is 313 to 324. HA, HB and HD have the same layout as HC.

**To check:** only floor 1 was seen for most hostels (floor 2 for HE). Floors 2 and up are taken to repeat floor 1 with their own first digit. NK's K rooms on the upper floors (K201, K301, ...) follow the same rule, but this is not confirmed. Some floors have two restrooms on the plan (HC, HF, NK); they are one `Restroom` area, and the student says which in the description.

Example rows (HC, floor 1):

| Category | Building | Wing | Area |
|---|---|---|---|
| Hostel | HC | Floor 1 | 101 |
| Hostel | HC | Floor 1 | 102 |
| Hostel | HC | Floor 1 | ... |
| Hostel | HC | Floor 1 | 113 |
| Hostel | HC | Floor 1 | Restroom |

**Academic Block** (20 rooms, in one building named after the category, no wing):

| Category | Building | Wing | Area |
|---|---|---|---|
| Academic Block | Academic Block | | CR1 |
| Academic Block | Academic Block | | CR2 |
| Academic Block | Academic Block | | CR3 |
| Academic Block | Academic Block | | CR4 |
| Academic Block | Academic Block | | CR5 |
| Academic Block | Academic Block | | IT1 |
| Academic Block | Academic Block | | IT2 |
| Academic Block | Academic Block | | IT3 |
| Academic Block | Academic Block | | IT4 |
| Academic Block | Academic Block | | IT5 |
| Academic Block | Academic Block | | Lab1 |
| Academic Block | Academic Block | | Lab2 |
| Academic Block | Academic Block | | Lab3 |
| Academic Block | Academic Block | | Lab4 |
| Academic Block | Academic Block | | Lab5 |
| Academic Block | Academic Block | | LH1 |
| Academic Block | Academic Block | | LH2 |
| Academic Block | Academic Block | | LH3 |
| Academic Block | Academic Block | | LH4 |
| Academic Block | Academic Block | | LH5 |

**Others** (7 places, building `Campus`, no wing):

| Category | Building | Wing | Area |
|---|---|---|---|
| Others | Campus | | Library |
| Others | Campus | | Mess |
| Others | Campus | | Football ground |
| Others | Campus | | Basketball court (indoor) |
| Others | Campus | | Basketball court (outdoor) |
| Others | Campus | | Volleyball court |
| Others | Campus | | Gym |

## 4. Sources

| What | Who | Date |
|---|---|---|
| Hostel, Academic Block and Others names; no canteens; Gym | Gayley Choden | 12 November 2026 |
| Hostel floors, room numbers and restrooms | Hostel floor plans: first floor of HC, HF, NK, RKA, RKB, second floor of HE (screenshots from Gayley Choden; HA, HB and HD are the same as HC) | 12 November 2026 |
| Categories and heads | `category-taxonomy` code | 12 November 2026 |
