# EVM log: Sprint 3 (3 to 8 November 2026, with work until the rehearsal on 10 November)

Card: SCRUM-82. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Planned hours, logged hours and statuses were read from Jira on 10 November 2026 (46 cards in Sprint 3). All values are person-hours.

## 1. Result

PV is still the baseline 93.3 h per sprint, because the revision asked for in weeks 1 and 2 was not made. The team's own Sprint 3 plan (the estimates on the Sprint 3 cards) is shown next to it, because it is the better yardstick (section 3).

| Metric | Sprint 3 | Cumulative (Sprint 1 to 3) | How |
|---|---|---|---|
| PV (Planned Value), baseline | 93.3 h | 280 h | `schedule-evm.md`: 280 h / 3 sprints |
| PV, team's Sprint 3 plan | 186.5 h | | Estimates of the Sprint 3 cards, SCRUM-98 moved to Sprint 4 left out |
| EV (Earned Value) | 174.5 h | 447.5 h | Planned hours of the cards that are Done |
| AC (Actual Cost) | 199 h | 424 h | Hours logged in Jira |
| CV = EV - AC | -24.5 h | +23.5 h | |
| SV = EV - PV (baseline) | +81.2 h | +167.5 h | |
| SV = EV - PV (team plan) | -12 h | | |
| **CPI = EV / AC** | **174.5 / 199 = 0.88** | **447.5 / 424 = 1.06** | Below 1 this sprint: the finished work took about 14% more hours than planned |
| **SPI = EV / PV (baseline)** | **174.5 / 93.3 = 1.87** | **447.5 / 280 = 1.60** | Far above 1 only because the baseline PV is too low |
| **SPI = EV / PV (team plan)** | **174.5 / 186.5 = 0.94** | | Close to 1: 94% of the planned work is Done |

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
| | SCRUM-99 President dashboard | 4 | 3 | In Review |
| | SCRUM-110 accessibility tests | 4 | 4 | In Review |
| | SCRUM-85 closure report | 4 | 2.5 | To Do |
| | *Subtotal* | *60* | *59.5* | *EV 48* |
| Nyendrak | SCRUM-89 escalation Saga | 7 | 10 | Done |
| | SCRUM-107 property-based tests (gopter) | 3 | 3 | Done |
| | SCRUM-104 Knative and PingSource | 4 | 4 | Done |
| | SCRUM-35 student report form | 4 | 5 | Done |
| | SCRUM-36 staff view | 6 | 7 | Done |
| | SCRUM-37 admin dashboard (counts) | 6 | 7 | Done |
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
| **Team** | | **186.5** | **199** | **EV 174.5** |

Not counted: SCRUM-98 PWA installability (Nyendrak, 7 h planned, 0 h logged) moved to Sprint 4, together with a new card for editing routing rules and SLA hours on the admin page, which waits for gateway admin routes. SCRUM-99, SCRUM-110 and SCRUM-85 (Gayley, 12 h planned) are not Done yet; their 9.5 logged hours are already in AC, so EV rises to 186.5 h and CPI to 0.94 when they close.

## 3. Reading

- **Cost (CPI 0.88).** For the first time CPI is below 1: the team logged 199 h for work planned at 174.5 h. The overruns are on integration work that had to be done twice or that others waited on: the escalation Saga (SCRUM-89, 10 h for 7 h, a race found by repeated test runs), the Kubernetes manifests (SCRUM-123), Argo CD (SCRUM-102), the three database adapter cards (SCRUM-124, 125, 127) and the visual design (SCRUM-138). The three UI pages (SCRUM-35, 36, 37) each took one hour more than planned because they were first built in the plain style and then rebuilt when the shared design and components arrived in the middle of the sprint. Cumulatively the project is still under cost (CPI 1.06).
- **Schedule.** Against the team's own plan, SPI is 0.94: everything is Done except one moved card and three of Gayley's cards in review. Against the baseline SPI is 1.87, which again only shows that the baseline PV (93.3 h) is about half of what the team plans and does each sprint (135 h in Sprint 2, 186.5 h in Sprint 3).
- **Scope added during the sprint.** 19 cards with 70.5 planned hours (SCRUM-122 to 142) were created during the sprint, from 6 November: the event bus, Kubernetes manifests, database adapters, gateway routes, permissions, event consumers and the visual design. All of them are Done. Without them the Sprint 3 plan was 116 h, so more than a third of the sprint was found while working: pieces the pages and the demo needed that were not on a card at planning.
- **Cumulative.** EV after three sprints is 447.5 h, more than the 280 h tracking baseline and the 360 h BAC. The baseline under-counted the work (open question 1 in `schedule-evm.md`); EV against it no longer says how much is left.

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
4. **Close the cards in review** (SCRUM-99, 110, 85) before the closure report, so the final EV is complete.
5. **Log hours the same day** the work is done.
