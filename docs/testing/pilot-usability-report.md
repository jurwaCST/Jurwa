# Pilot usability test: report (SCRUM-113)

Owner: Sonam Choki. Script: `docs/testing/pilot-script.md`. Notes: `docs/evidence/sonam-choki/sprint3/pilot/notes.md` (typed) and the photos of the sheets next to it.

**Result: objective 2 is met.** All 5 students submitted a report with a photo in under two minutes: average 62 seconds, slowest 90 seconds, nobody needed help. The problems they had were all about choosing the location.

## 1. Setup

| Item | Value |
|---|---|
| Date | 9 November 2026 |
| Where | Demo laptop: k3d cluster deployed by Argo CD from `main`, 20 of 20 services up; web app with `npm run dev`, gateway through `kubectl port-forward` |
| Participants | 5 students (P1 to P5). No staff: task 4 could not be run (section 2) |
| Accounts | `p1.pilot@rub.edu.bt` to `p5.pilot@rub.edu.bt`, created before the sessions through `POST /auth/register`, because the web app has no sign-up page (problem 3) |
| Photo | One JPG of a problem, on the laptop's desktop |
| Consent | Read out before each session; no names or faces recorded |

Students were free to report a problem in any category, so the reports cover Hostel (P1, P2), Academic Block (P3, P4) and Others (P5).

## 2. Tasks

| # | Task | Run? | Why not |
|---|---|---|---|
| 1 | Log in | Yes, 5 of 5 | |
| 2 | Report a problem with a photo | Yes, 5 of 5 | |
| 3 | Find your report and read its status | Not available | The student home page has only "Report an issue"; there is no "my reports" page yet |
| 4 | Staff: open a report, set In Progress | Not available | `/staff` is still a placeholder ("Staff ticket view: Nyendrak, SCRUM-36") |

Tasks 1 and 2 are the card's done-condition (pilot script, section "Tasks").

## 3. Results

| Code | Category | Task 1 | Task 2 time | Helped | Errors | What confused them (their words, shortened) |
|---|---|---|---|---|---|---|
| P1 | Hostel | Done | 60 s | No | None | "Wing 1/2" under hostel |
| P2 | Hostel | Done | 90 s | No | None | Room and bathroom could be more specific ("HA-101") |
| P3 | Academic | Done | 45 s | No | None | Nothing; would like the hall number ("LT-01") |
| P4 | Academic | Done | 65 s | No | None | East / West wing is confusing |
| P5 | Others | Done | 50 s | No | None | Wi-Fi and Estate; CST has 3 canteens |

| Measure | Value | Target |
|---|---|---|
| Task 1 completed | 5 of 5 | |
| Task 2 completed | 5 of 5 | |
| Task 2 average | **62 s** | under 120 s |
| Task 2 slowest | **90 s** (P2) | under 120 s |
| Task 2 fastest | 45 s (P3) | |
| Help needed | 0 of 5 | |
| Error messages seen by participants | 0 | |

P1's time was noted as "1 min" and is counted as 60 s.

## 4. Top 5 problems

Ranked by how many participants it affected and how much it would matter for real users.

| # | Problem | Seen by | Owner | Card |
|---|---|---|---|---|
| 1 | **The wing step is confusing.** "Wing 1 / Wing 2" in the hostels and "East Wing / West Wing" in Academic Block 1 mean nothing to students; they asked what a wing is or wanted it removed | P1, P4 | Sonam Choki (location data, `location-taxonomy`) and Nyendrak (form text, `ReportForm.tsx`) | New card for the location data; form hint under SCRUM-35 |
| 2 | **Areas are not specific enough.** Students expected real room and hall names: "HA-101" instead of "Room 101", "LT-01" instead of "Lecture Hall 1", and which of the 3 canteens instead of one "Canteen" | P2, P3, P5 | Sonam Choki (`location-taxonomy` sample data) | New card: replace the sample locations with the real CST list |
| 3 | **No sign-up page in the web app.** A new student cannot create an account in the app; the pilot accounts were made through the API | Observer (setup) | Gayley Choden (`Login.tsx`, identity) | New card |
| 4 | **A student cannot see their reports, and staff cannot work on them** (tasks 3 and 4 not available) | Observer (setup) | Nyendrak (`StaffView.tsx`, SCRUM-36); "my reports" page has no owner yet (gateway route from SCRUM-136 exists) | SCRUM-36; new card for "my reports" |
| 5 | **"Wi-Fi" and "Estate" are unclear as places** under Others: students did not know what to pick for a problem that is not in a building | P5 | Sonam Choki (`location-taxonomy`) | Same card as problem 2 |

### Other findings (not seen by participants)

| Finding | Owner |
|---|---|
| The gateway accepted a report with building "x", a location that does not exist: the location is not checked on submit | Sonam Choki (`report-submission`) |
| The first submit after report-submission restarted failed with "The service is busy or unreachable. Please try again." (gateway 2 s timeout); the retry worked | Sonam Choki (`report-submission`), Sonam Zangmo (gateway timeout) |
| The confirmation shows the report number as a 36-character id, which a student cannot remember or read out | Nyendrak (`ReportForm.tsx`, SCRUM-35) |

## 5. Limits

- 5 students, one observer, one laptop, one day. The numbers show the form works, not that it is fast for every student.
- All participants were CST students who knew the campus and the observer, and computer students may be faster than others.
- The locations are sample data, so problems 1, 2 and 5 may change once the real list is entered.
- Tasks 3 and 4 were not tested; staff were not part of this pilot.

## 6. Follow-up

Each problem is sent to its owner by message; their cards are not edited by the pilot owner. If the staff view (SCRUM-36) is merged before the end of the project, task 4 can be run with 1 or 2 staff and added here as section 7.
