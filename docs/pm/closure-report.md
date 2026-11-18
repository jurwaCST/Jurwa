# Closure report: Jurwa (SCRUM-85)

Owner: Gayley Choden (Scrum Master). Draft, updated 19 November 2026 after Sprint 4. Module: SWE303.

Every number below has a source file in the repository, named next to it. Parts marked **Waiting for** need something that does not exist yet (the recorded rehearsal, the presentation and the sign-off); they are filled in before this report is merged.

## 1. Objectives (charter section 3)

| # | Objective | Result | Evidence |
|---|---|---|---|
| 1 | 20 microservices that route a report to the correct Category Head with 100% accuracy on the test rule-set | **Met.** 20 services in `services/`. Every category and location routes correctly (Hostel to DSA, Academic Block to DAA, Others to SSO), and the description never changes the route. In Sprint 4 the real CST buildings, floors and areas replaced the sample list (552 locations, SCRUM-150) | `tests/spec/routing_rules_test.go`; `docs/testing/test-plan.md` S-08, S-09; `docs/evidence/sonam-choki/sprint4/` |
| 2 | An installable PWA in which a student submits a report with a photo in under two minutes, verified by usability testing | **Met.** Pilot with 5 students: all submitted a report with a photo in under two minutes, average **62 s**, slowest **90 s**, nobody needed help. Since Sprint 4 the app is installable (SCRUM-98: manifest, a hand-written service worker that keeps only the app's own files and never an API answer, and an offline banner; Lighthouse on mobile: accessibility 100, best practices 100, SEO 92, performance 79). A student can now follow the report after sending it: My reports with the timeline (SCRUM-143), notifications with a bell (SCRUM-144, with the messages sent by SCRUM-155) and feedback stars once it is resolved (SCRUM-145) | `docs/testing/pilot-usability-report.md`; `docs/evidence/nyendrak/sprint4/scrum-98-*`; `docs/evidence/gayley/sprint4/` |
| 3 | Automatic SLA escalation by a scheduled serverless function, shown in the final demo | **Met in the cluster; live demo to come.** The sla-escalation-checker runs as a Knative Service started by a PingSource every 2 minutes and scales to zero between runs (SCRUM-104); the CronJob fallback is kept for the demo laptop. The escalation Saga and its compensation pass their tests (SG-01 to SG-05). **Waiting for** the recorded rehearsal (SCRUM-149): whether a real escalation is shown live | `docs/evidence/nyendrak/sprint3/knative.txt`, `deploy/knative/`, `tests/saga/` |
| 4 | A test pyramid tracked in SonarQube, at least 70% coverage on routing, SLA and status-flow code | **Met.** 10 of 10 critical packages at 98.2% or more (Sprint 3 run); 9 gopter properties; JMeter p95 252 ms with 0 errors. Web unit tests went from 112 to 173 in Sprint 4. Browser tests: 26 of 26 Cypress tests pass (report flow 8, accessibility 15, student pages 3), and axe-core finds 0 violations on 15 page states, including the three new student pages. SonarQube (6 Nov scan): coverage 61.7%, no issue in a Go service | `docs/testing/test-plan.md` section 5, `docs/testing/sonarqube.md`, `docs/testing/accessibility.md`, `docs/evidence/nyendrak/sprint4/scrum-152-tests.txt` |
| 5 | Containers on Kubernetes (k3d) through GitOps (GitHub Actions and Argo CD), shown in a live demo | **Met for the pipeline.** CI builds and tests every pull request; Argo CD syncs the cluster from `main` and heals a change made by hand (api-gateway scaled to 0, back to 1 in 2 seconds). The web app runs in the cluster too, behind Traefik on `/` next to the API on `/api` (SCRUM-139, SCRUM-140). **Waiting for** the live demo | `.github/workflows/`, `deploy/argocd/`, `docs/evidence/sonam-zangmo/sprint3/argocd.txt`, `docs/evidence/sonam-choki/sprint3/web-deploy-cluster.txt` |
| 6 | Within the schedule and a zero budget, tracked weekly with EVM | **Budget met:** only free and open-source tools. **Schedule:** one sprint longer than planned, see section 3 | `docs/pm/evm/` |

## 2. Scope

**Built: 20 services and 2 serverless functions, not the 25 services of the first proposal.**

| Change | From | To | Why | Source |
|---|---|---|---|---|
| C1 | 25 services | 20 services: export, holiday-calendar, digest, email-template and user-profile removed; digest emails are sent by the `digest-dispatcher` function | Less time (C2). Risk R4: the lowest-priority services are cut first; testing and documentation are never cut | `docs/pm/schedule-evm.md` section 5 |
| C2 | 8 weeks | 3 five-day sprints, presentation about 9 November | The presentation date moved | `docs/pm/schedule-evm.md` section 5 |
| C3 | 6 or 7 services per member | 5 services per member, plus one function for two members | An even load | `docs/pm/schedule-evm.md` section 5 |
| C4 | 3 sprints | A fourth sprint, 11 to 20 November (first planned to end on 17 November) | The rehearsal and presentation moved; the student pages the pilot found missing were built in it | `docs/pm/evm/week4.md` |

Holidays and weekends are therefore not counted in SLA windows (the holiday-calendar service was one of the five removed).

**Added in Sprint 3 beyond the plan:** a student sign-up page (SCRUM-138), the campus design with CST photos, and shared web components for the remaining pages (SCRUM-141). The President dashboard (SCRUM-99, a Stretch card) was built and merged. The staff view (SCRUM-36) and the admin dashboard (SCRUM-37) were merged on the last day.

**Built in Sprint 4** (17 cards, 51.5 h planned):

| Area | Cards |
|---|---|
| Student pages: My reports and the timeline, notifications and the bell, feedback stars | SCRUM-143, 144, 145, with the gateway routes in SCRUM-147 |
| Messages to students from report events (received, in progress, resolved, escalated) | SCRUM-155 |
| Installable PWA | SCRUM-98 |
| Real CST locations and categories, and the form saying "Floor" for hostels | SCRUM-150, 154 |
| Wording and role fixes: Home wording, departments 5 or 6, timeline status words, the report route shown only to students | SCRUM-146, 148, 157, 156 |
| Secrets check on the public repository | SCRUM-151 |
| Cypress and axe tests for the student pages | SCRUM-152 |
| EVM week 4, rehearsal, this report | SCRUM-153, 149, 85 |

Four of these cards were found during the sprint (7 h): SCRUM-154, 155, 156 and 157.

**Not built (future work):**
- Editing routing rules and SLA hours on the admin page: the gateway has no admin routes for them.
- A database adapter for sla-tracking, so its timers survive a restart (until then, the demo rule: no restart of sla-tracking or sla-policy after T minus 90).
- "Mark as read" for notifications: there is no route, so the browser remembers the newest message seen, per device.
- Staff without an institutional email (R5, FR-24 is a Could), Jaeger tracing, the godog runner for the BDD scenarios.
- Defect D-08 (report-submission saves a report before it publishes the event).

## 3. Schedule and cost (EVM)

| Sprint | PV (baseline) | PV (team plan) | EV | AC | CPI | SPI (team plan) | Source |
|---|---|---|---|---|---|---|---|
| 1 (22 to 27 Oct) | 93.3 h | | 138 h | 96 h | 1.44 | | `docs/pm/evm/week1.md` |
| 2 (28 Oct to 3 Nov) | 93.3 h | | 135 h | 129 h | 1.05 | | `docs/pm/evm/week2.md` |
| 3 (4 to 10 Nov) | 93.3 h | 182.5 h | 182.5 h | 199 h | 0.92 | 1.00 | `docs/pm/evm/week3.md` |
| 4 (11 to 20 Nov) | 0 h (not in the baseline) | 51.5 h | 42 h | 52.5 h | 0.80 | 0.82 | `docs/pm/evm/week4.md` |
| **Cumulative** | **280 h** | | **497.5 h** | **476.5 h** | **1.04** | | `docs/pm/evm/week4.md` |

Read honestly:
- **The project as a whole is under cost** (cumulative CPI 1.04), but each sprint was closer to 1 or below it: 1.44, 1.05, 0.92, 0.80.
- **Sprint 4's CPI 0.80 is mostly timing.** The finished cards cost almost exactly their plan (43 h logged for 42 h of EV). Four cards were logged but not Done when the log was taken (SCRUM-155, 156, 149 and 85, 9.5 h). When they are Done, Sprint 4 EV is 51.5 h and CPI 0.98 (`week4.md` section 3).
- **The baseline was too small and never revised.** It spread 280 h over three sprints; the project took four, and 507 h of planned work is done or nearly done (EV 497.5 h plus the 9.5 h of the four open cards). Against the baseline, SPI is 1.78, which only shows that the baseline under-counted the work (lesson L6).
- **Estimates came late in Sprint 1** (issue I4), so its CPI of 1.44 is optimistic. In Sprint 4, most hours were again logged after the sprint ended, when asked for the EVM (`week4.md` section 4, action 5).

## 4. Risks and issues at the end

| Risk | What happened | Status at the end | Source |
|---|---|---|---|
| R1 scope grows | Scope went down, not up: 25 to 20 services (C1). Sprint 4 added 4 small cards found while building (7 h) | Closed: did not happen | `risk-register.md` 2.2; `evm/week4.md` |
| R2 learning curve | Local setup cost time in Sprint 1. Sprint 2 services were built from the template without new tool problems; Sprint 3 had new Windows problems with Testcontainers and Cypress, solved and documented | Reduced | `risk-register.md` 2.2 |
| R3 cluster or Knative unstable near the demo | Argo CD deploys the cluster; Knative runs the escalation function with a PingSource, and the CronJob fallback is ready for the demo laptop. **Waiting for** the recorded rehearsal | Reduced | `risk-register.md` section 2.3; SCRUM-83, SCRUM-104 |
| R4 schedule slips | Work landed on the last day of each sprint, and the project needed a fourth sprint (C4). Must cards were finished and Stretch cards cut, in the planned order | Happened, handled by cutting Stretch work and adding Sprint 4 | `risk-register.md` 2.2; `evm/week4.md` |
| R5 staff without institutional email | Not built (FR-24 is a Could) | Open, future work | `test-plan.md` section 4 |

**Issue log at the end** (`docs/pm/issue-log.md`): I1, I3, I9 and I10 closed; I2, I4 and I8 in progress; I5, I6 and I7 open. I6 (old Jira cards from the 25-service design) is closed by closing those cards as "Won't do". I10: the free GitHub Actions minutes for a private repository ran out, so every check was blocked and no pull request could merge. Fixed on 9 November at zero cost by making the repository public, after checking that the whole Git history held no secret (no paid budget, NFR-11); SCRUM-151 then scanned the public repository for secrets with gitleaks.

**Defects:** 16 defects found in Sprints 2 and 3 (D-01 to D-16) are listed with how each was found and who fixed it in `docs/testing/test-plan.md` section 6. Retries after a timeout that could create duplicates (D-07) were fixed in SCRUM-132. D-08 is still open (section 2, future work).

## 5. What was delivered

| Area | Where |
|---|---|
| Services and functions | `services/`, `functions/`, one README each |
| Contracts | `proto/` |
| Web app (installable PWA) | `web/` |
| Deployment | `deploy/k8s/`, `deploy/argocd/`, `deploy/knative/`, `deploy/observability/` |
| CI | `.github/workflows/` |
| Tests | `tests/`, `web/cypress/`, `docs/testing/` |
| Project management | `docs/pm/`, `docs/sprint-reviews/` |
| Evidence per member | `docs/evidence/gayley/`, `docs/evidence/nyendrak/`, `docs/evidence/sonam-choki/`, `docs/evidence/sonam-zangmo/` |

## 6. Acceptance

**Waiting for** the recorded rehearsal (SCRUM-149), the final presentation and the tutors' sign-off. The team accepts the delivery when the demo script (`docs/pm/demo-script.md`) runs end to end, or its recorded fallback is shown.

| Name | Role | Accepted |
|---|---|---|
| Sonam Zangmo | Team Lead | |
| Gayley Choden | Scrum Master | |
| Nyendrak | Developer | |
| Sonam Choki | Developer | |
