# EVM log: Sprint 3 (3 to 8 November 2026, with work until the rehearsal on 10 November)

Card: SCRUM-82. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Planned hours, logged hours and statuses were read from Jira on 10 November 2026 (46 cards in Sprint 3), and again after the rehearsal-day updates. All values are person-hours.

Cards that are not Done carry over to Sprint 4 with the hours not yet spent: SCRUM-98 (7 h, nothing logged) and SCRUM-85 (the closure report, 2.5 of 4 h logged). They are left out of the Sprint 3 plan and EV; the hours already logged on them stay in Sprint 3 AC, because they were spent in Sprint 3.

## 1. Result

PV is still the baseline 93.3 h per sprint, because the revision asked for in weeks 1 and 2 was not made. The team's own Sprint 3 plan (the estimates on the Sprint 3 cards) is shown next to it, because it is the better yardstick (section 3).

| Metric | Sprint 3 | Cumulative (Sprint 1 to 3) | How |
|---|---|---|---|
| PV (Planned Value), baseline | 93.3 h | 280 h | `schedule-evm.md`: 280 h / 3 sprints |
| PV, team's Sprint 3 plan | 182.5 h | | Estimates of the Sprint 3 cards, without SCRUM-98 and SCRUM-85 (moved to Sprint 4) |
| EV (Earned Value) | 182.5 h | 455.5 h | Planned hours of the cards that are Done |
| AC (Actual Cost) | 199 h | 424 h | Hours logged in Jira |
| CV = EV - AC | -16.5 h | +31.5 h | |
| SV = EV - PV (baseline) | +89.2 h | +175.5 h | |
| SV = EV - PV (team plan) | 0 h | | |
| **CPI = EV / AC** | **182.5 / 199 = 0.92** | **455.5 / 424 = 1.07** | Below 1 this sprint: the finished work took about 9% more hours than planned |
| **SPI = EV / PV (baseline)** | **182.5 / 93.3 = 1.96** | **455.5 / 280 = 1.63** | Far above 1 only because the baseline PV is too low |
| **SPI = EV / PV (team plan)** | **182.5 / 182.5 = 1.00** | | Every card kept in Sprint 3 is Done |

Cumulative values add Sprint 1 (EV 138 h, AC 96 h, `week1.md`) and Sprint 2 (EV 135 h, AC 129 h, `week2.md`).

## 2. Per card

| Owner | Card | Planned (h) | Logged (h) | Status |
|---|---|---|---|---|
| Gayley Choden | SCRUM-105 observability | 6 | 6 | Done |
| | SCRUM-106 integration tests (Testcontainers) | 5 | 4 | Done |
| | SCRUM-108 end-to-end tests (Cypress) | 7 | 6 | Done |
| | SCRUM-112 test plan | 5 | 4 | Done |
| | SCRUM-118 web: React PWA setup | 5 | 6 | Done |
| | SCRUM-126 identity: demo accounts | 2 | 2 | Done |
| | SCRUM-127 database adapters | 6 | 8 | Done |
| | SCRUM-137 authorization: report.list_assigned | 1 | 1 | Done |
| | SCRUM-138 web: visual design | 6 | 8 | Done |
| | SCRUM-141 web: shared components | 5 | 5 | Done |
| | SCRUM-99 President dashboard | 4 | 3 | Done |
| | SCRUM-110 accessibility tests | 4 | 4 | Done |
| | SCRUM-85 closure report | (4) | 2.5 | Moved to Sprint 4 |
| | *Subtotal* | *56* | *59.5* | *EV 56* |
| Nyendrak | SCRUM-89 escalation Saga | 7 | 10 | Done |
| | SCRUM-107 property-based tests (gopter) | 3 | 3 | Done |
| | SCRUM-104 Knative and PingSource | 4 | 4 | Done |
| | SCRUM-35 student report form | 4 | 5 | Done |
| | SCRUM-36 staff view | 6 | 7 | Done |
| | SCRUM-37 admin dashboard (counts) | 6 | 7 | Done |
| | SCRUM-98 PWA installability | (7) | 0 | Moved to Sprint 4 |
| | *Subtotal* | *30* | *36* | *EV 30* |
| Sonam Choki | SCRUM-39 health checks | 3 | 3 | Done |
| | SCRUM-103 Traefik and cert-manager TLS | 5 | 5 | Done |
| | SCRUM-111 SonarQube | 5 | 4 | Done |
| | SCRUM-113 pilot usability test | 5 | 5 | Done |
| | SCRUM-116 internal-notes permissions | 1 | 1 | Done |
| | SCRUM-122 pkg/eventbus (Redis Streams) | 6 | 6 | Done |
| | SCRUM-123 Kubernetes manifests | 8 | 10 | Done |
| | SCRUM-125 database adapters | 6 | 8 | Done |
| | SCRUM-130 report-history event consumers | 3 | 3 | Done |
| | SCRUM-131 report-history ListReportsByAssignee | 2 | 2 | Done |
| | SCRUM-135 k8s gateway addresses | 0.5 | 0.5 | Done |
| | SCRUM-139 k8s: deploy the web app | 3 | 3 | Done |
| | *Subtotal* | *47.5* | *50.5* | *EV 47.5* |
| Sonam Zangmo | SCRUM-84 final presentation and rehearsal | 6 | 6 | Done |
| | SCRUM-102 Argo CD and Argo Rollouts | 8 | 10 | Done |
| | SCRUM-109 performance tests (JMeter) | 5 | 6 | Done |
| | SCRUM-117 gateway: notes routes | 2 | 2 | Done |
| | SCRUM-119 gateway: history routes | 2 | 2 | Done |
| | SCRUM-120 gateway: photo routes | 3 | 3 | Done |
| | SCRUM-121 resilience: gomock tests | 1 | 1 | Done |
| | SCRUM-124 database adapters | 6 | 7 | Done |
| | SCRUM-128 gateway: status routes | 3 | 3 | Done |
| | SCRUM-129 gateway: dashboard routes | 3 | 3 | Done |
| | SCRUM-132 resilience: no retry after a timeout | 2 | 2 | Done |
| | SCRUM-133 routing-engine: report.routed | 3 | 3 | Done |
| | SCRUM-136 gateway: assigned reports route | 2 | 2 | Done |
| | SCRUM-142 gateway: reassign and positions routes | 3 | 3 | Done |
| | *Subtotal* | *49* | *53* | *EV 49* |
| **Team** | | **182.5** | **199** | **EV 182.5** |

Planned hours in brackets are not counted in Sprint 3.

Carried over to Sprint 4: SCRUM-98 PWA installability (Nyendrak, 7 h), the rest of SCRUM-85 closure report (Gayley, 1.5 h of 4 h), and a new card for editing routing rules and SLA hours on the admin page (Nyendrak, 4 h), which waits for gateway admin routes. Jira has no Sprint 4 yet; it has to be created before these cards can move.

## 3. Reading

- **Cost (CPI 0.92).** For the first time CPI is below 1: the team logged 199 h for work planned at 182.5 h. 2.5 h of that is the closure report, which carries over, so the Done work itself is closer to plan (CPI 0.93 without it). The overruns are on integration work that had to be done twice or that others waited on: the escalation Saga (SCRUM-89, 10 h for 7 h, a race found by repeated test runs), the Kubernetes manifests (SCRUM-123), Argo CD (SCRUM-102), the three database adapter cards (SCRUM-124, 125, 127) and the visual design (SCRUM-138). The three UI pages (SCRUM-35, 36, 37) each took one hour more than planned because they were first built in the plain style and then rebuilt when the shared design and components arrived in the middle of the sprint. Cumulatively the project is still under cost (CPI 1.06).
- **Schedule.** Against the team's own plan, SPI is 1.00: every card kept in Sprint 3 is Done, and the two unfinished cards carry over to Sprint 4 with their remaining hours. Against the baseline SPI is 1.96, which again only shows that the baseline PV (93.3 h) is about half of what the team plans and does each sprint (135 h in Sprint 2, 182.5 h in Sprint 3).
- **Scope added during the sprint.** 19 cards with 70.5 planned hours (SCRUM-122 to 142) were created during the sprint, from 6 November: the event bus, Kubernetes manifests, database adapters, gateway routes, permissions, event consumers and the visual design. All of them are Done. Without them the Sprint 3 plan was 112 h, so more than a third of the sprint was found while working: pieces the pages and the demo needed that were not on a card at planning.
- **Cumulative.** EV after three sprints is 455.5 h, more than the 280 h tracking baseline and the 360 h BAC. The baseline under-counted the work (open question 1 in `schedule-evm.md`); EV against it no longer says how much is left.

## 4. Sprint 2 actions

| Action | Result |
|---|---|
| 1. Revise the PV before Sprint 3 | Not done. This log adds the team's plan as a second PV |
| 2. An owner and a card for every shared piece | Met: event bus SCRUM-122 (Sonam Choki), React app SCRUM-118 and shared components SCRUM-138 and 141 (Gayley) |
| 3. Planned hours before In Progress, one card In Progress per person | Partly met. All 46 cards have planned hours. Nyendrak had three UI cards In Progress at once while waiting for the design |
| 4. Log hours the same day | Not met for Nyendrak: all Sprint 3 hours were logged on 10 November |

## 5. Actions for Sprint 4 and the closure report

1. **Rebaseline from Jira**: set PV per sprint to the team's planned hours, and replace the 280 h and 360 h BAC with the sum of the estimates still open, so the closure report's EVM uses one consistent baseline.
2. **Shared design before pages**: a page card starts only when the components it uses are merged, so pages are not built twice.
3. **List the gateway routes in each page card** at planning (for example the admin routes for editing routing rules and SLA hours), so the routes are planned and not added mid-sprint.
4. **Create Sprint 4 in Jira** and move SCRUM-98, SCRUM-85 and the new admin editing card into it, with their remaining hours as the Sprint 4 plan.
5. **Log hours the same day** the work is done.
