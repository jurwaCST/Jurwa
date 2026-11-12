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
2. **No wings.** "Wing 1 / Wing 2" and "East / West Wing" confused students (P1, P4), so no entry has a wing.
3. **No canteens.** Canteens cannot be reported in Jurwa. This replaces the pilot's request to name the 3 canteens (P5).
4. **Others are places, not services.** "Wi-Fi" and "Estate" were unclear as places (P5), so they are removed. A Wi-Fi problem is reported at the place where it happens.

What the service accepts (`services/location-taxonomy/domain/location.go`):
- Building is required; wing and area are optional.
- Others have no wing: their building is always `Campus`, and each area is one place.
- A sample name that is no longer used is set inactive, not deleted, so old reports still show it.

## 3. The list

**Hostel** (7 buildings, no wing, no area):

| Category | Building | Wing | Area |
|---|---|---|---|
| Hostel | HA | | |
| Hostel | HB | | |
| Hostel | HC | | |
| Hostel | HD | | |
| Hostel | HE | | |
| Hostel | HF | | |
| Hostel | NK | | |

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

**Others** (6 places, building `Campus`, no wing):

| Category | Building | Wing | Area |
|---|---|---|---|
| Others | Campus | | Library |
| Others | Campus | | Mess |
| Others | Campus | | Football ground |
| Others | Campus | | Basketball court (indoor) |
| Others | Campus | | Basketball court (outdoor) |
| Others | Campus | | Volleyball court |

## 4. Sources

| What | Who | Date |
|---|---|---|
| Hostel, Academic Block and Others names; no canteens | Gayley Choden | 12 November 2026 |
| Categories and heads | `category-taxonomy` code | 12 November 2026 |
