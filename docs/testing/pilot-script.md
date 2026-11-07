# Pilot usability test: script (SCRUM-113)

Owner: Sonam Choki. Goal (proposal objective 2): a student submits a report with a photo in **under two minutes**. Results go in `docs/testing/pilot-usability-report.md`.

## Who

| Code | Role | How many |
|---|---|---|
| P1 to P5 (or more) | Students | at least 5 |
| C1, C2 | Category Head or staff (demo account `dsa.demo@jurwa.example` or `staff.demo@jurwa.example`) | 1 or 2 |

**Consent (read out before starting):** "This tests the app, not you. We time the task and write down what is unclear. We do not record your name or face; you are P1 to P6 in our notes. You can stop at any time." Start only after a yes.

## Setup (once, before the first participant)

1. Cluster running, all services up: `bash scripts/check-all-health.sh --k8s` shows `20 up, 0 down`.
2. Demo accounts switched on in identity (password from `dev/.env`, never written down here).
3. Gateway reachable from the laptop: `kubectl -n jurwa port-forward svc/api-gateway 8080:8080`.
4. Web app: in `web/`, `.env.local` contains `VITE_API_URL=/api`, then `npm run dev` and open http://localhost:5173.
5. For each student, a fresh account is registered beforehand (email `p1.pilot@rub.edu.bt` and so on), so the timing measures the report, not sign-up.
6. A JPG photo of a real problem (for example a leaking tap) is on the laptop's desktop, under 5 MB.
7. Stopwatch on a phone. One printed copy of the notes sheet below per participant.

## Tasks

Say the task once, then watch without helping. If the participant is stuck for 60 seconds, help, and write "helped" in the notes.

| # | Who | Task as read out | Done when |
|---|---|---|---|
| 1 | Student | "Log in with this email and password." | Home page shows the student's name |
| 2 | Student | "The tap in your hostel bathroom is leaking. Report it, with the photo on the desktop." | Confirmation shown after submit |
| 3 | Student | "Find the report you just sent and tell me its status." | Participant reads the status |
| 4 | Category Head or staff | "A new hostel report has come to you. Open it and change its status to In Progress." | Status shows In Progress |

**State of the app on 8 Nov 2026 (checked before the sessions):** tasks 1 and 2 can be done in the app (login page, report form with photo upload, confirmation "Report sent. Your report number is ..."). Task 3 needs a "my reports" page and task 4 the staff view (Nyendrak, SCRUM-36); both pages are not in the app yet, and the gateway has no "assigned to me" route yet. If they are still missing on the test day, write "not available" for tasks 3 and 4 in the report and run tasks 1 and 2 only. Tasks 1 and 2 are the card's done-condition.

**Timing for task 2:** start when the report form is on screen, stop when the confirmation appears. Write it in seconds.

## Notes sheet (one per participant)

| Field | Write |
|---|---|
| Code and role | P_ , Student / Category Head / Staff |
| Task 1 done? helped? | |
| Task 2 time (seconds) | |
| Task 2 errors (wrong field, error message, back button, retried upload) | |
| Task 3 done? helped? | |
| Task 4 done? helped? (staff only) | |
| "What confused you?" (their words) | |
| Observer notes | |

## After all sessions

Copy the sheets into `docs/testing/pilot-usability-report.md`: the table of all participants, the average and slowest time for task 2 against 120 seconds, the top 5 problems, and for each problem the owner and card. Send each problem to its owner by message; do not edit their cards. Photos of the notes (no faces) go in `docs/evidence/sonam-choki/sprint3/pilot/`.
