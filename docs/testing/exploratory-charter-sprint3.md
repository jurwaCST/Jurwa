# Exploratory testing charter, Sprint 3

Tester: Gayley Choden. Card: SCRUM-112. Method: one member explores the running app off script each sprint (quality plan section 3, "Structured charters").

**Status: not run yet.** The sections below "Session" are filled in during and after the session. Nothing in them is written in advance.

## Charter

| | |
|---|---|
| Mission | Explore the student's report flow and the Category Head's view to find behaviour that the scripted tests miss: confusing messages, wrong roles, lost data, odd inputs |
| Build | `main` at the commit written below, started with `web/cypress/start-stack.sh` and `npm run dev` (all services in memory) |
| Time box | 45 minutes |
| Accounts | A new student registered during the session; the demo accounts from the identity seed (`dsa.demo`, `staff.demo`, `president.demo`, `admin.demo`) |
| Out of scope | Load and accessibility (they have their own tests); the Kubernetes cluster |

## Areas and ideas to try

| Area | Ideas |
|---|---|
| Sign in and register | Upper-case email, spaces around the email, a very long name, signing in twice in two tabs, signing out and pressing Back |
| Report form | Exactly 500 and 501 characters, Dzongkha text, emoji, only spaces, pasting text, a 4.9 MB and a 5.1 MB photo, a GIF renamed to `.png`, removing the photo after choosing it, pressing Submit twice quickly |
| Roles | Opening `/report/new` as staff, a Category Head page as a student, the dashboard as a student, a report id of another student typed in the address bar |
| Category Head | The assigned list after a new Hostel report, changing the status out of order, two heads changing the same report (409) |
| Errors | Stopping one service (for example photo-media) while the form is open, then submitting; slow network in the browser's developer tools |
| Phone | The app at phone width in the browser's device mode: can every field be reached and read |

## Session

| | |
|---|---|
| Date and time | |
| Commit (`git rev-parse --short HEAD`) | |
| Actual time spent | |

### Notes

(What was tried, in order, with what happened.)

### Bugs found

| # | What happened | Steps to reproduce | Expected | Owner told | Defect log id |
|---|---|---|---|---|---|
| | | | | | |

Each bug is sent as a message to the owner and added to the defect log in `docs/testing/test-plan.md` section 6.

### Questions

(Things that are unclear: not bugs, but worth asking the owner or the team.)
