# EVM log: Sprint 1 (22 to 27 October 2026)

Card: SCRUM-82. Author: Nyendrak Yoezer Zangmo. Baseline: `docs/pm/schedule-evm.md`.
Data read from Jira (project SCRUM, open sprint) and GitHub on 27 October 2026.

## 1. Numbers

| Metric | Value | Source |
|---|---|---|
| PV (Planned Value, Sprint 1) | 93.3 h | Baseline, section 3 (280 h / 3 sprints) |
| EV (Earned Value) | **unknown** | Needs the original estimate of each Done card. No Sprint 1 card has an estimate in Jira |
| AC (Actual Cost) | **unknown** | Needs logged hours. No hours are logged on any Sprint 1 card in Jira |
| CPI = EV / AC | **cannot be calculated** | |
| SPI = EV / PV | **cannot be calculated** | |

The numbers are left blank on purpose. The baseline rule is that hours are never invented.

## 2. Progress by card count (what can be measured today)

21 cards are in Sprint 1.

| Jira status | Cards | Count |
|---|---|---|
| Done | SCRUM-13, 17, 44, 71, 72, 75, 76, 77, 78, 79, 80, 81, 100, 101 | 14 |
| In Review | SCRUM-18 | 1 |
| In Progress | SCRUM-19, 73, 74, 82, 90 | 5 |
| To Do | SCRUM-38 | 1 |

14 of 21 cards (67%) are Done in Jira. This is only a rough indicator, because cards differ in size. It is not EV.

**Jira does not match GitHub.** The pull requests for SCRUM-18 (#24), SCRUM-73 (#11) and SCRUM-74 (#19) are merged, but the cards are not Done in Jira. Once those three are moved, 17 of 21 cards (81%) are Done. SCRUM-19 is waiting for review (pull request #25).

## 3. Reading

- **Schedule:** most Sprint 1 work is finished. The template (SCRUM-73) and contracts (SCRUM-74) landed on 27 Oct instead of 26 Oct, so the service thin slices were squeezed into the last day. SCRUM-38 (api-gateway) and SCRUM-90 (routing-engine) carry over to Sprint 2.
- **Cost:** unknown. Without logged hours the team cannot tell whether Sprint 1 took more or less effort than planned. This is the main monitoring gap (risk R4).

## 4. Actions for Sprint 2

1. **Every owner adds an original estimate** to each Sprint 2 card before the sprint starts on 28 Oct, so the estimates add up to the Sprint 2 PV (93.3 h).
2. **Every owner logs hours on their Sprint 1 cards by 28 Oct.** This log is then updated with real EV, AC, CPI and SPI.
3. **Move cards to Done when their pull request merges** (SCRUM-18, 73, 74 now). The Scrum Master adds "log hours, move cards" to each stand-up.

## 5. To fill in once hours are logged

| Card | Owner | Estimate (h) | Logged (h) | Done? |
|---|---|---|---|---|
| (one row per Sprint 1 card) | | | | |
| **Total** | | EV = sum of estimates of Done cards | AC = sum of logged | |

Then: CPI = EV / AC, SPI = EV / 93.3.
