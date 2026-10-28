# EVM log: Sprint 1 (22 to 27 October 2026)

Card: SCRUM-82. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Planned hours per card were agreed by the team. Logged hours and statuses were read from Jira on 28 October 2026. All values are person-hours.

## 1. Result

| Metric | Value | How |
|---|---|---|
| PV (Planned Value) | 93.3 h | Baseline: 280 h / 3 sprints |
| EV (Earned Value) | 138 h | Sum of planned hours of the 18 cards that are Done |
| AC (Actual Cost) | 90 h | Sum of hours logged in Jira |
| CV = EV - AC | +48 h | |
| SV = EV - PV | +44.7 h | |
| **CPI = EV / AC** | **138 / 90 = 1.53** | Above 1: the finished work took fewer hours than planned |
| **SPI = EV / PV** | **138 / 93.3 = 1.48** | Above 1: more planned work was finished than the baseline expected for Sprint 1 |

## 2. Per card

| Owner | Card | Planned (h) | Logged (h) | Status |
|---|---|---|---|---|
| Sonam Choki | SCRUM-77 charter | 5 | 5 | Done |
| | SCRUM-78 requirements | 11 | 7 | Done |
| | SCRUM-76 wireframes | 8 | 5 | Done |
| | SCRUM-100 k3d cluster | 12 | 7 | Done |
| | SCRUM-101 CI | 8 | 4 | Done |
| | SCRUM-17 report-submission | 7 | **not logged** | Done |
| | *Subtotal* | *51* | *28* | |
| Sonam Zangmo | SCRUM-73 template | 10 | 8 | Done |
| | SCRUM-74 contracts | 8 | 7 | Done |
| | SCRUM-72 bounded contexts | 4 | 4 | Done |
| | SCRUM-18 routing-rules | 7 | 4 | Done |
| | *Subtotal* | *29* | *23* | |
| Nyendrak | SCRUM-75 database schema | 7 | 6 | Done |
| | SCRUM-79 schedule and EVM baseline | 8 | 5 | Done |
| | SCRUM-19 status-lifecycle | 7 | 5 | Done |
| | *Subtotal* | *22* | *16* | |
| Gayley Choden | SCRUM-71 learning spike | 8 | 6 | Done |
| | SCRUM-80 quality plan | 7 | 4 | Done |
| | SCRUM-81 communications plan | 5 | 3 | Done |
| | SCRUM-44 PM documentation | 9 | 4 | Done |
| | SCRUM-13 identity | 7 | 6 | Done |
| | *Subtotal* | *36* | *23* | |
| **Team** | | **138** | **90** | |

Not counted (no planned hours, not Done): SCRUM-38 api-gateway (To Do), SCRUM-90 routing-engine (In Progress), SCRUM-82 this log (In Progress). All three carry over to Sprint 2.

## 3. Reading

- **Schedule (SPI 1.48).** All 18 estimated cards were finished. SPI is high mainly because the cards the team planned for Sprint 1 add up to 138 h, while the baseline gave Sprint 1 only 93.3 h. Sprint 1 also carried the initiation, planning and design work (WP 1.1 to 1.3, 80 h in the proposal). This supports counting those 80 h inside the plan (open question 1 in the baseline). The baseline PV per sprint should be revised for Sprint 2.
- **Cost (CPI 1.53).** The team used 90 h for work planned at 138 h. Treat this as optimistic, for three reasons:
  1. The planned hours were entered at the end of the sprint, not before it started, so they may lean high.
  2. SCRUM-17 has no logged hours, so AC is too low. Without SCRUM-17, CPI is 131 / 90 = 1.46.
  3. AI assistance (allowed by the module) shortened some tasks.
- **Delivery.** The template and contracts merged on the last day, so every service slice was squeezed into 27 October. Two of Sonam Zangmo's cards (SCRUM-38, SCRUM-90) carry over.

## 4. Actions for Sprint 2

1. **Estimate before starting.** Every Sprint 2 card gets its planned hours on 28 October, before work begins.
2. **Log hours daily**, and log SCRUM-17 now so this log can be corrected.
3. **Revise the per-sprint PV** with the Scrum Master and Team Lead, using the Sprint 1 estimates (138 h) and the 360 h vs 280 h decision. Then update `docs/pm/schedule-evm.md`.
4. **Merge shared blockers first.** Shared cards that block others (template, contracts) come first in each sprint.

## 5. Check by hand

- EV: 51 + 29 + 22 + 36 = 138
- AC: 28 + 23 + 16 + 23 = 90
- CPI: 138 / 90 = 1.533
- SPI: 138 / 93.3 = 1.479
