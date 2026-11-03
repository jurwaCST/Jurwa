# EVM log: Sprint 2 (28 October to 3 November 2026)

Card: SCRUM-82. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Planned hours, logged hours and statuses were read from Jira on 3 November 2026. All values are person-hours.

## 1. Result

PV is still the baseline 93.3 h per sprint, because action A3 from week 1 (revise the per-sprint PV) has not been done yet.

| Metric | Sprint 2 | Cumulative (Sprint 1 + 2) | How |
|---|---|---|---|
| PV (Planned Value) | 93.3 h | 186.7 h | Baseline |
| EV (Earned Value) | 135 h | 273 h | Planned hours of the cards that are Done |
| AC (Actual Cost) | 129 h | 225 h | Hours logged in Jira |
| CV = EV - AC | +6 h | +48 h | |
| SV = EV - PV | +41.7 h | +86.3 h | |
| **CPI = EV / AC** | **135 / 129 = 1.05** | **273 / 225 = 1.21** | Close to 1: the work took about the hours planned |
| **SPI = EV / PV** | **135 / 93.3 = 1.45** | **273 / 186.7 = 1.46** | Above 1, mainly because the baseline PV is too low (see section 3) |

## 2. Per card

| Owner | Card | Planned (h) | Logged (h) | Status |
|---|---|---|---|---|
| Sonam Choki | SCRUM-21 internal-notes | 5 | 4 | Done |
| | SCRUM-22 report-history | 6 | 5 | Done |
| | SCRUM-23 location-taxonomy | 6 | 4 | Done |
| | SCRUM-28 photo-media | 6 | 5 | Done |
| | SCRUM-40 containerize all services | 5 | 7 | Done |
| | SCRUM-83 issue log and risk review | 3 | 3 | Done |
| | SCRUM-94 BDD scenarios | 5 | 5 | Done |
| | *Subtotal* | *36* | *33* | |
| Sonam Zangmo | SCRUM-24 category-taxonomy | 5 | 5 | Done |
| | SCRUM-32 department-directory | 5 | 4 | Done |
| | SCRUM-38 api-gateway | 8 | 8 | Done |
| | SCRUM-90 routing-engine | 6 | 9 | Done |
| | SCRUM-91 resilience patterns | 6 | 7 | Done |
| | SCRUM-93 specification and state-transition tests | 6 | 5 | Done |
| | SCRUM-114 api-gateway CheckPermission | 2 | 1 | Done |
| | SCRUM-115 photo spec tests | 1 | 1 | Done |
| | *Subtotal* | *39* | *40* | |
| Nyendrak | SCRUM-29 sla-policy | 5 | 5 | Done |
| | SCRUM-86 sla-tracking | 6 | 5 | Done |
| | SCRUM-88 reassignment | 6 | 4 | Done |
| | SCRUM-27 sla-escalation-checker | 6 | 5 | Done |
| | SCRUM-42 audit-log | 5 | 4 | Done |
| | SCRUM-92 unit tests, TDD and gomock | 5 | 5 | Done |
| | *Subtotal* | *33* | *28* | |
| Gayley Choden | SCRUM-14 authorization | 5 | 6 | Done |
| | SCRUM-26 notification | 6 | 6 | Done |
| | SCRUM-31 analytics | 5 | 5 | Done |
| | SCRUM-33 feedback | 5 | 4 | Done |
| | SCRUM-97 digest-dispatcher | 6 | 7 | Done |
| | *Subtotal* | *27* | *28* | |
| **Team** | | **135** | **129** | |

Not counted (not Done, no planned hours yet, carry over to Sprint 3): SCRUM-89 escalation Saga (needs the event bus), SCRUM-104 Knative and PingSource, SCRUM-35 student report form and SCRUM-98 PWA installability (both need the React app in `web/`).

## 3. Reading

- **Cost (CPI 1.05).** The team used 129 h for work planned at 135 h, so the estimates were close this sprint. That is much closer to 1 than Sprint 1 (1.44), mostly because three of the four members entered their planned hours before starting. The biggest overruns were SCRUM-90 routing-engine (9 h for 6 h) and SCRUM-40 containerization (7 h for 5 h). Both are shared work that other cards depended on.
- **Schedule (SPI 1.45).** All 27 estimated cards are Done. SPI is high again because the baseline gives each sprint 93.3 h, while the team plans about 135 h per sprint. So SPI mostly shows the baseline is too low, not that the team is far ahead. Revising the PV (A3) is now overdue.
- **Carry-overs.** The four open cards are Nyendrak's. They wait on shared pieces that have no owner yet (the Redis Streams event bus, and the React app in `web/`), not on late work.
- **Cumulative.** After two sprints, EV is 273 h of the 280 h remaining-work baseline, but Sprint 3 still holds deployment, integration and end-to-end testing, the UI and the final presentation. This confirms the baseline under-counts the work (open question 1 in `schedule-evm.md`: BAC 280 h or 360 h).

## 4. Week 1 actions

| Action | Result |
|---|---|
| A1 Estimate before starting | Partly met. Gayley, Sonam Choki and Sonam Zangmo estimated first. Nyendrak's planned hours were entered at the end of the sprint |
| A2 Log hours daily | Not met. Most hours were logged on 2 and 3 November |
| A3 Revise the per-sprint PV | Not done |
| A4 Merge shared blockers first | Met for the template, contracts and resilience package. New shared blockers appeared (event bus, `web/`) |

## 5. Actions for Sprint 3

1. **Revise the PV before Sprint 3 work starts** (A3), with the Scrum Master and Team Lead. Use about 135 h per sprint, and decide BAC 280 h or 360 h. Then update `docs/pm/schedule-evm.md`.
2. **Give every shared piece an owner and a card at planning**: the React app in `web/` and the Redis Streams event bus.
3. **Planned hours before In Progress**, and only one card per person In Progress.
4. **Log hours the same day** the work is done.

## 6. Check by hand

- EV: 36 + 39 + 33 + 27 = 135
- AC: 33 + 40 + 28 + 28 = 129
- CPI: 135 / 129 = 1.047
- SPI: 135 / 93.3 = 1.447
- Cumulative EV: 138 + 135 = 273; AC: 96 + 129 = 225
- Cumulative CPI: 273 / 225 = 1.213; SPI: 273 / 186.7 = 1.462
