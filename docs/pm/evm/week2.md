# EVM log: Sprint 2 (28 October to 3 November 2026)

Card: SCRUM-82. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Planned and logged hours and statuses were read from Jira on 2 November 2026 at 22:45. All values are person-hours.

> **DRAFT.** Logged hours are still missing for Sonam Choki, Sonam Zangmo and Nyendrak, and Nyendrak's planned hours are not in Jira yet. Values marked *pending* are filled in from Jira before this is merged.

## 1. Result

PV is still the baseline 93.3 h per sprint: action A3 from week 1 (revise the per-sprint PV) has not been done yet.

| Metric | Sprint 2 only | Cumulative (Sprint 1 + 2) | How |
|---|---|---|---|
| PV (Planned Value) | 93.3 h | 186.7 h | Baseline |
| EV (Earned Value) | *pending* (135 h if Nyendrak's planned hours are 33 h) | *pending* (273 h) | Planned hours of Done cards |
| AC (Actual Cost) | *pending* | *pending* | Hours logged in Jira |
| CV = EV - AC | *pending* | *pending* | |
| SV = EV - PV | *pending* (+41.7 h) | *pending* (+86.3 h) | |
| CPI = EV / AC | *pending* | *pending* | |
| SPI = EV / PV | *pending* (135 / 93.3 = 1.45) | *pending* (273 / 186.7 = 1.46) | |

## 2. Per card

| Owner | Card | Planned (h) | Logged (h) | Status |
|---|---|---|---|---|
| Sonam Choki | SCRUM-21 internal-notes | 5 | *pending* | Done |
| | SCRUM-22 report-history | 6 | *pending* | Done |
| | SCRUM-23 location-taxonomy | 6 | *pending* | Done |
| | SCRUM-28 photo-media | 6 | *pending* | Done |
| | SCRUM-40 containerize all services | 5 | *pending* | Done |
| | SCRUM-83 issue log and risk review | 3 | *pending* | Done |
| | SCRUM-94 BDD scenarios | 5 | *pending* | Done |
| | *Subtotal* | *36* | *pending* | |
| Sonam Zangmo | SCRUM-24 category-taxonomy | 5 | *pending* | Done |
| | SCRUM-32 department-directory | 5 | *pending* | Done |
| | SCRUM-38 api-gateway | 8 | *pending* | Done |
| | SCRUM-90 routing-engine | 6 | *pending* | Done |
| | SCRUM-91 resilience patterns | 6 | *pending* | Done |
| | SCRUM-93 specification and state-transition tests | 6 | *pending* | Done |
| | SCRUM-114 api-gateway CheckPermission | 2 | *pending* | Done |
| | SCRUM-115 photo spec tests | 1 | *pending* | Done |
| | *Subtotal* | *39* | *pending* | |
| Nyendrak | SCRUM-29 sla-policy | *pending* | *pending* | Done |
| | SCRUM-86 sla-tracking | *pending* | *pending* | Done |
| | SCRUM-88 reassignment | *pending* | *pending* | Done |
| | SCRUM-27 sla-escalation-checker | *pending* | *pending* | Done |
| | SCRUM-42 audit-log | *pending* | *pending* | Done |
| | SCRUM-92 unit tests, TDD and gomock | *pending* | *pending* | Done |
| | *Subtotal* | *pending* | *pending* | |
| Gayley Choden | SCRUM-14 authorization | 5 | 6 | Done |
| | SCRUM-26 notification | 6 | 6 | Done |
| | SCRUM-31 analytics | 5 | 5 | Done |
| | SCRUM-33 feedback | 5 | 4 | Done |
| | SCRUM-97 digest-dispatcher | 6 | 7 | Done |
| | *Subtotal* | *27* | *28* | |
| **Team** | | *pending* | *pending* | |

Not counted (not Done, carry over to Sprint 3): SCRUM-89 escalation Saga (needs the event bus), SCRUM-104 Knative and PingSource, SCRUM-35 student report form and SCRUM-98 PWA installability (both need the React app in `web/`, which has no owner yet).

## 3. Reading

*To be written when the numbers are complete.* Points already known:

- **Schedule.** Every card except Nyendrak's four carry-overs is Done. EV is again above the 93.3 h baseline, so A3 (revise the per-sprint PV) is now overdue.
- **Estimates.** Week 1 action A1 (estimate before starting) was met by Gayley, Sonam Choki and Sonam Zangmo. Nyendrak's planned hours were entered at the end of the sprint.
- **Logging.** Week 1 action A2 (log hours daily) was not met: on 2 November only Gayley's hours were in Jira.
- **Blockers.** The carry-overs come from shared pieces with no owner (event bus, `web/` setup), not from late work. See action 4 of week 1.

## 4. Actions for Sprint 3

*To be agreed at the retro.* Proposed:

1. Revise the per-sprint PV now (A3), before Sprint 3 starts.
2. Give every shared piece an owner and a card at planning: the React app in `web/` and the Redis Streams event bus.
3. Every card gets planned hours before it moves to In Progress, and only one card per person is In Progress.

## 5. Check by hand

*To be filled in with the final numbers.*
