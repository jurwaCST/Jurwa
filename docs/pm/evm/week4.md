# EVM log: Sprint 4 (11 to 17 November 2026)

Card: SCRUM-153. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Planned hours, logged hours and statuses were read from Jira on 18 November 2026 with `sprint = 69` (17 cards in Sprint 4). All values are person-hours.

Same method as week 3: hours logged on a carried-over card stay in the sprint where they were spent. SCRUM-98 had nothing logged in Sprint 3, so all 7 h are Sprint 4. SCRUM-85 had 2.5 of 4 h logged in Sprint 3, so its Sprint 4 plan and AC are the remaining 1.5 h. SCRUM-153 (this log) is counted as Done with its 2 h, because this file is what it delivers.

## 1. Result

Sprint 4 is outside the baseline: `schedule-evm.md` spreads 280 h over three sprints and has no PV for a fourth. The Sprint 4 PV is therefore the team's own plan, the estimates of the Sprint 4 cards (week 3, action 1).

| Metric | Sprint 4 | Cumulative (Sprint 1 to 4) | How |
|---|---|---|---|
| PV (Planned Value), team's Sprint 4 plan | 51.5 h | | Estimates of the 17 Sprint 4 cards; SCRUM-85 at its remaining 1.5 h |
| PV, baseline | 0 h (not in the baseline) | 280 h | `schedule-evm.md`: 280 h over Sprints 1 to 3 |
| EV (Earned Value) | 42 h | 497.5 h | Planned hours of the cards that are Done |
| AC (Actual Cost) | 51.5 h | 475.5 h | Hours logged in Jira |
| CV = EV - AC | -9.5 h | +22 h | |
| SV = EV - PV (team plan) | -9.5 h | | |
| **CPI = EV / AC** | **42 / 51.5 = 0.82** | **497.5 / 475.5 = 1.05** | Below 1 only because four cards are logged but not Done yet (section 3) |
| **SPI = EV / PV (team plan)** | **42 / 51.5 = 0.82** | | 13 of 17 cards Done |
| SPI = EV / PV (baseline) | | 497.5 / 280 = 1.78 | Above 1 only because the baseline PV is too low |

Cumulative values add Sprints 1 to 3 from `week3.md` (EV 455.5 h, AC 424 h).

## 2. Per card

| Owner | Card | Planned (h) | Logged (h) | Status |
|---|---|---|---|---|
| Gayley Choden | SCRUM-143 web: My reports and report timeline | 6 | 6 | Done |
| | SCRUM-144 web: notifications page and bell | 5 | 5 | Done |
| | SCRUM-145 web: feedback stars | 4 | 4 | Done |
| | SCRUM-146 web: Home wording and real CST list | 2 | 2 | Done |
| | SCRUM-155 notification: messages from report events | 4 | 4 | In review (PR #134, approved, not merged) |
| | SCRUM-156 web: Home shows the report route only to students | 1 | 1 | To Do |
| | SCRUM-85 closure report (rest, carried from Sprint 3) | 1.5 | 1.5 | In Progress |
| | *Subtotal* | *23.5* | *23.5* | *EV 17* |
| Nyendrak | SCRUM-98 PWA installability (carried from Sprint 3) | 7 | 7 | Done |
| | SCRUM-152 tests and evidence for the student pages | 4 | 4 | Done |
| | SCRUM-153 Sprint 4 EVM and demo script | 2 | 2 | Done with this log |
| | SCRUM-154 report form: Floor instead of Wing | 1 | 1 | Done |
| | *Subtotal* | *14* | *14* | *EV 14* |
| Sonam Choki | SCRUM-150 real CST buildings, floors, areas and categories | 3 | 4 | Done |
| | SCRUM-151 secrets check on the public repository | 1 | 0 | Done, hours not logged yet |
| | SCRUM-157 timeline status words match the badges | 1 | 1 | Done |
| | *Subtotal* | *5* | *5* | *EV 5* |
| Sonam Zangmo | SCRUM-147 gateway: notifications and feedback routes | 4 | 4 | Done |
| | SCRUM-148 department-directory: 5 vs 6 departments | 2 | 2 | Done |
| | SCRUM-149 demo rehearsal and final recording | 3 | 3 | In Progress |
| | *Subtotal* | *9* | *9* | *EV 6* |
| **Team** | | **51.5** | **51.5** | **EV 42** |

SCRUM-151 is Done in Jira with no hours logged. When they are logged, AC rises by that amount (1 h as planned gives AC 52.5 h, CPI 0.80, cumulative AC 476.5 h).

## 3. Reading

- **Cost (CPI 0.82).** The cards that are Done cost exactly what was planned: 42 h logged for 42 h of EV (CPI 1.00 on finished work). Only SCRUM-150 went over (4 h for 3 h, the real CST list was longer than expected), and SCRUM-151 shows 0 h only because it is not logged yet. The whole gap is the four cards that are logged but not Done: SCRUM-155 (4 h, waiting for its merge), SCRUM-149 (3 h, the rehearsal), SCRUM-85 (1.5 h, the closure report) and SCRUM-156 (1 h). Each has all its planned hours logged and no time remaining, so when they are Done, EV becomes 51.5 h and CPI about 1.00 (0.98 with SCRUM-151's hour).
- **Schedule (SPI 0.82).** 13 of 17 cards are Done. The four open cards are the last steps before the presentation (rehearsal, closure report) or are finished and waiting for review (SCRUM-155). There is no Sprint 5, so they are not carried over: they close in the days before the presentation.
- **Scope added during the sprint.** 4 cards with 7 planned hours were created after the sprint started: SCRUM-154 (Floor label, from Choki's SCRUM-150 data, 12 Nov), SCRUM-155 (the notification consumer, found while building the notifications page, 12 Nov), SCRUM-156 (17 Nov) and SCRUM-157 (from the SCRUM-152 screenshots, 18 Nov). The plan at the start was 44.5 h. Much less was found mid-sprint than in Sprint 3 (19 cards, 70.5 h), because the routes and data the pages needed were planned as their own cards (SCRUM-147, 150).
- **Size of the sprint.** 51.5 h is much smaller than Sprints 2 and 3 (135 h and 182.5 h): Sprint 4 was finish and polish, with the presentation in the same week.
- **Cumulative.** After four sprints EV is 497.5 h and AC 475.5 h (CPI 1.05): the project as a whole is under cost. EV is 138% of the 360 h BAC; as in week 3, the baseline under-counted the work, so EV against it does not say how much is left.

## 4. Sprint 3 actions

| Action | Result |
|---|---|
| 1. Rebaseline from Jira | Partly. `schedule-evm.md` still has the 280 h baseline. This log uses the team's Sprint 4 plan as the only PV for Sprint 4 |
| 2. Shared design before pages | Met. Gayley's pages (SCRUM-143, 144, 145) were built on the shared components, and SCRUM-152 started only after they were merged |
| 3. List the gateway routes in each page card | Met. The notifications and feedback routes had their own card at planning (SCRUM-147) |
| 4. Create Sprint 4 and move the carried cards | Met for SCRUM-98 and SCRUM-85. The admin editing card was not created: it needs gateway admin routes that nobody planned, so it goes to the closure report as future work |
| 5. Log hours the same day | Not met. Most Sprint 4 hours were logged on 18 November, after the sprint ended and when asked for this log; SCRUM-151 is still not logged |

## 5. Actions before the closure report

1. **Log SCRUM-151** (Choki), then update section 1 and 2 of this file with the real hours.
2. **Close the four open cards**: merge SCRUM-155 (PR #134, approved), finish SCRUM-156, SCRUM-149 (rehearsal and recording) and SCRUM-85 (closure report), and move them to Done.
3. **Closure report EVM**: use the cumulative numbers from this log (EV 497.5 h, AC 475.5 h, CPI 1.05), and say that the 280 h baseline covered three sprints while the project took four, with 507 h of planned work done or nearly done (EV 497.5 h plus the 9.5 h of the four open cards).
4. **Future work**: editing routing rules and SLA hours on the admin page (needs gateway admin routes), and a database adapter for sla-tracking so timers survive a restart.
