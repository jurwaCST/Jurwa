# Sprint 2 review and retrospective

Sprint 2: 28 October to 3 November 2026. Scrum Master: Gayley Choden.
Sources: Jira project SCRUM (read again on 4 November 2026, after all hours were logged), merged pull requests on `main`, and each member's status and retro.

## 1. Sprint goal

Build the other 16 services and the two serverless functions as thin slices: built from the template, health check SERVING, one real call working, unit tests written first, merged.

**Result: met.** All **20 services** and both **functions** (`sla-escalation-checker`, `digest-dispatcher`) are on `main`. Every service has a separate "tests first" commit and evidence in `docs/evidence/<name>/`. Storage is in memory; real databases and Redis Streams events are Sprint 3.

## 2. What got done

| Owner | Card | PR |
|---|---|---|
| Sonam Choki | SCRUM-23 location-taxonomy | #31 (first attempt #29 reverted in #30) |
| | SCRUM-28 photo-media | #42 |
| | SCRUM-22 report-history | #43 |
| | SCRUM-21 internal-notes | #45 |
| | SCRUM-40 containerize all services | #48 |
| | SCRUM-83 risk register review and issue log | #49 |
| | SCRUM-94 BDD scenarios in Gherkin (stretch) | #51 |
| Sonam Zangmo | SCRUM-24 category-taxonomy | #32 |
| | SCRUM-32 department-directory | #33 |
| | SCRUM-90 routing-engine (carried from Sprint 1) | #34 |
| | SCRUM-91 resilience package (timeout, retry, circuit breaker) | #35 |
| | SCRUM-38 api-gateway (carried from Sprint 1) | #37 |
| | SCRUM-93 specification and state-transition tests (stretch) | #38 |
| | SCRUM-114 api-gateway uses authorization CheckPermission | #44 |
| | SCRUM-115 photo size and type spec tests | #50 |
| Nyendrak Yoezer Zangmo | SCRUM-29 sla-policy | #36 |
| | SCRUM-86 sla-tracking | #53 |
| | SCRUM-88 reassignment | #56 |
| | SCRUM-27 sla-escalation-checker function | #57, #61 |
| | SCRUM-42 audit-log | #58 |
| | SCRUM-92 unit tests and TDD (100% domain coverage, gomock) | #59 |
| Gayley Choden | SCRUM-14 authorization | #40 |
| | SCRUM-26 notification | #46 |
| | SCRUM-33 feedback | #47 |
| | SCRUM-31 analytics | #54 |
| | SCRUM-97 digest-dispatcher function | #55 |

Also merged during Sprint 2: the quality plan update for SCRUM-80 (#60, Gayley). SCRUM-80 is a Sprint 1 card, so it is not in the 26 cards or the hours below.

## 3. Carried over to Sprint 3

| Card | Owner | Reason |
|---|---|---|
| SCRUM-89 escalation Saga | Nyendrak | Needs the event bus (Redis Streams) |
| SCRUM-104 Knative Serving, Eventing and PingSource | Nyendrak | Deployment work, Sprint 3 |
| SCRUM-35 student report form | Nyendrak | Needs the React app in `web/` (SCRUM-118) |
| SCRUM-98 PWA installability | Nyendrak | Needs the React app in `web/` (SCRUM-118) |

None of the four has an original estimate yet.

## 4. Hours (from Jira)

| Member | Cards Done | Estimate of Done cards | Hours logged |
|---|---|---|---|
| Gayley Choden | 5 | 27 h | 28 h |
| Nyendrak Yoezer Zangmo | 6 | 33 h | 28 h |
| Sonam Choki | 7 | 36 h | 33 h |
| Sonam Zangmo | 8 | 39 h | 40 h |
| **Team** | **26** | **135 h** | **129 h** |

- All estimates and hours are read from Jira. Sonam Choki's and Sonam Zangmo's hours were logged after the sprint ended and match the per-card totals they sent on 2 November.
- First reading: EV 135 h (estimates of the 26 Done cards), AC 129 h, so CPI = 135 / 129 = 1.05 (about on budget). The Sprint 2 EVM log (Nyendrak) gives the final figures.
- Estimates were added at the end of the sprint again, so the indexes look better than they really are (risk R8, issue I4).

## 5. Sprint review

No review meeting was held; each member sent a written status. Highlights:

- **Escalation works end to end:** a real one-hour SLA lapse run escalated an open report and skipped a resolved one (Nyendrak, SCRUM-27).
- **The gateway checks roles for real:** it calls authorization `CheckPermission` on every signed-in route and returns 503 if authorization is down (Zangmo, SCRUM-114).
- **All services build as images:** `scripts/docker-build-all.sh` built every service with 0 failures, and the containers pass the gRPC health check (Choki, SCRUM-40).
- **Services tested against each other:** feedback against status-lifecycle, and digest-dispatcher against analytics, department-directory and notification (Gayley).

## 6. Retrospective

### What went well

- **Sonam Choki:** all seven cards, including the stretch card, merged with tests first and evidence. The photo upload, error messages and the notes role rule were agreed in the group chat before building, so the service, the gateway and the form match.
- **Sonam Zangmo:** all cards merged with tests first, and every service was checked live before its PR.
- **Nyendrak Yoezer Zangmo:** six cards merged with tests and evidence; the escalation chain works end to end; gomock was agreed quickly.
- **Gayley Choden:** all five cards (four services and the digest function) merged with tests first and 100% coverage on domain, application and adapters. Feedback and the digest were tested live against teammates' services. The quality plan was updated to match the code.

### What did not go well

- **Work landed on the last day.** Most PRs merged on 2 November, and three were stacked (SCRUM-90, 91, 38) so they had to merge in order (Zangmo, Gayley).
- **Estimates and hours came late again.** Estimates were entered at the end of the sprint; the time-tracking field was missing from Task cards until it was added; two members' hours were not logged in Jira by the end of the sprint (since entered) (Nyendrak, Choki, Gayley).
- **Shared pieces had no owner.** Nobody owned `web/` or the event bus, which blocks the report form, the PWA and the Saga (Nyendrak).
- **Cards sat In Progress without being started** (Nyendrak).
- **Review rules slipped.** One PR (#49) was merged before approval (Choki).
- **A role rule did not match the design.** internal-notes let the System Administrator add notes, against `api-routes.md`; found after the merge and fixed in SCRUM-116 (Choki).
- **Rework and setup:** the first location-taxonomy slice was reverted and redone; Docker Desktop filled a disk and blocked SCRUM-40; the feedback Dockerfile did not copy `pkg/` and was fixed in #55 (Choki, Gayley).

### What to change in Sprint 3

| # | Action | Owner |
|---|---|---|
| A1 | Every card gets an original estimate at planning, before it moves to In Progress | Everyone; Gayley checks |
| A2 | Log hours in Jira every day, with "Log work", not only the estimate, so no hours are entered after the sprint | Everyone |
| A3 | Shared pieces get an owner and a card at planning: `web/` is SCRUM-118 (Gayley). **The event bus (Redis Streams) still needs an owner and a card** | Gayley at planning |
| A4 | Open a PR as soon as a card is done; avoid stacked PRs; no card waits for the last day | Everyone |
| A5 | Merge only after the green Approved | Everyone |
| A6 | Check `api-routes.md` and the authorization policy before writing a role rule | Everyone |
| A7 | Settle open design questions at the start of the sprint (for example, the vacant Mechanical HoD, to confirm with the tutors) | Zangmo |
| A8 | Only one card In Progress per person, and only when work has really started | Everyone |

## 7. Sprint 3

Sprint 3: 3 to 8 November 2026. Goal: deploy to k3d with Argo CD, real databases and Redis Streams events, the React PWA, the full test suite, and the final demo (presentation about 9 November).
