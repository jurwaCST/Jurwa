# Lessons learned (SCRUM-85)

Facilitator: Gayley Choden (Scrum Master). Whole team, 30 minutes, after the final demo.

**Status: the session is not held yet.** Section 2 lists the lessons the team already wrote down in the retrospectives and saw again in Sprints 3 and 4, as the starting point. Section 3 is filled in during the session; nothing in it is written in advance.

## 1. Agenda

1. What worked (10 minutes)
2. What hurt (10 minutes)
3. What we would do differently next time, with a name against each action (10 minutes)

## 2. Lessons already recorded

| # | Lesson | Seen in | Evidence |
|---|---|---|---|
| L1 | **Estimates entered late give a misleading EVM.** Sprint 1 estimates and hours were entered at the end, so CPI 1.44 was optimistic; once estimates came first in Sprint 2, CPI was 1.05. In Sprint 4 most hours were again logged after the sprint ended, when asked for the EVM | Sprint 1, 2, 4 | `docs/pm/evm/week1.md`, `week2.md`, `week4.md` section 4; issue I4 |
| L2 | **Work lands on the last day.** Cards and pull requests were merged on the final day of each sprint, so reviews were rushed and the board moved in bulk | Sprint 1, 2, 3 | Sprint 2 retro A4; risk R4, R12 |
| L3 | **Shared pieces need an owner and a card at planning.** The template and contracts landed a day late and blocked every service slice; the event bus had no owner at the start of Sprint 3 | Sprint 1, 2 | Issue I1; Sprint 2 retro A3; risk R7 |
| L4 | **A role rule that does not match the design.** internal-notes let the System Administrator add notes, against `api-routes.md`, and was found after the merge | Sprint 2 | Issue I8; Sprint 2 retro A6; defect D-01 |
| L5 | **A shared folder change breaks other services' images.** Services that started using `pkg/` had Dockerfiles that did not copy it, three times; the Images check caught it each time | Sprint 2, 3 | Defects D-03, D-04, D-05 |
| L6 | **The baseline PV was never revised.** Action A3 (revise the per-sprint PV) was open after Sprint 1 and still open after Sprint 2, so SPI stayed high for the wrong reason. The baseline covered three sprints and the project took four, so Sprint 4 had no baseline PV at all | Sprint 1, 2, 3, 4 | `docs/pm/evm/week2.md` section 4, `week4.md` section 1 |
| L7 | **Tests found real defects before users did.** The JMeter run found that retries can make duplicates (D-07); the end-to-end check found new reports missing from the timeline (D-12) | Sprint 3 | `docs/testing/test-plan.md` section 6 |
| L8 | **Know the free limits of every tool before the last week.** The free GitHub Actions minutes for a private repository ran out two days before the demo and blocked every merge. Making the repository public fixed it at zero cost, but a check of the usage page each sprint would have shown it coming | Sprint 3 | Issue I10 |
| L9 | **Re-run the tests after a redesign, on the build that is deployed.** Re-running the accessibility tests on the final design found a real problem (D-16) that the earlier passing run could not; and tests against the dev server failed at random under load, while the built app passed | Sprint 3 | `docs/testing/accessibility.md`; defect D-16 |
| L10 | **Plan what fills a page, not only its route.** The notifications page and its gateway route were planned, but nothing created a notification, so the inbox would have been empty in the demo. It was found while building the page and became its own card (SCRUM-155) | Sprint 4 | SCRUM-147 review; `docs/pm/evm/week4.md` section 3 |
| L11 | **Build test data from the real answer.** The report page passed its tests but was blank against the real gateway: the test data wrapped the report in an object, the gateway does not. Writing test data from a `curl` of the real route would have caught it | Sprint 4 | SCRUM-143 commits (report page no longer blank) |
| L12 | **Look at every page as every role, on a phone.** Signing in as each demo account showed that staff, admin and the President saw the student's "Where your report goes" on Home (SCRUM-156), and screenshots from another member showed the timeline's status words in capitals (SCRUM-157) | Sprint 4 | `docs/evidence/gayley/sprint4/scrum-156-*.png`; SCRUM-152 screenshots |

## 3. Session notes

| | |
|---|---|
| Date and time | |
| Present | |

### What worked

### What hurt

### What we would do differently

| Action | Owner |
|---|---|
| | |
