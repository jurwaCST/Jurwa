# Sprint 1 review and retrospective

Sprint 1: 22 to 27 October 2026. Scrum Master: Gayley Choden.
Sources: Jira project SCRUM, merged pull requests on `main`, the EVM log (`docs/pm/evm/week1.md`) and each member's retro.

## 1. Sprint goal

Charter, requirements, plans, wireframes, service template, gRPC contracts, k3d cluster, CI, and one thin-slice service per owner (from `docs/pm/schedule-evm.md`).

**Result: met.** All 18 estimated cards are Done and merged. All four thin-slice services build and their tests pass on `main`.

## 2. What got done

| Owner | Card | PR |
|---|---|---|
| Sonam Choki | SCRUM-77 project charter | #6 |
| | SCRUM-78 requirements | #7 |
| | SCRUM-76 UI/UX wireframes | #10 |
| | SCRUM-100 k3d cluster | #12 |
| | SCRUM-101 CI pipeline | #13, #16 |
| | SCRUM-17 report-submission thin slice | #22 |
| Sonam Zangmo | SCRUM-73 service template | #11 |
| | SCRUM-74 gRPC contracts | #19 |
| | SCRUM-72 bounded contexts | #21 |
| | SCRUM-18 routing-rules thin slice | #24 |
| Nyendrak Yoezer Zangmo | SCRUM-75 database schema | #14 |
| | SCRUM-79 schedule and EVM baseline | #15 |
| | SCRUM-19 status-lifecycle thin slice | #25 |
| | SCRUM-82 first EVM log | #26 |
| Gayley Choden | SCRUM-71 learning spike | #2 |
| | SCRUM-80 quality management plan | #3 |
| | SCRUM-81 communications plan | #4 |
| | SCRUM-44 PM documentation (stakeholder register, risk register, WBS) | #20 |
| | SCRUM-13 identity thin slice | #23 |

Evidence for each member is in `docs/evidence/<name>/`.

## 3. Carried over to Sprint 2

| Card | Owner | Reason |
|---|---|---|
| SCRUM-90 routing-engine | Sonam Zangmo | Stretch card, started but not finished |
| SCRUM-38 api-gateway | Sonam Zangmo | Stretch card, not started |

## 4. EVM (from `docs/pm/evm/week1.md`)

| PV | EV | AC | CPI | SPI |
|---|---|---|---|---|
| 93.3 h | 138 h | 96 h | 1.44 | 1.48 |

Both indexes are above 1, but treat them as optimistic. Estimates were entered at the end of the sprint, and the Sprint 1 cards add up to 138 h against a baseline of 93.3 h, so the per-sprint PV needs revising.

## 5. Sprint review

Demo: each owner runs their thin slice and calls it with grpcurl.

| Service | Owner | Shown |
|---|---|---|
| identity | Gayley Choden | Register, Login, ValidateToken |
| report-submission | Sonam Choki | SubmitReport, GetReport |
| routing-rules | Sonam Zangmo | GetRule |
| status-lifecycle | Nyendrak Yoezer Zangmo | Valid and invalid status changes |

## 6. Retrospective

### What went well

- **Sonam Choki:** all 7 cards merged within the sprint. report-submission was written test-first with 100% coverage on domain, application and adapter code. The k3d cluster and CI were ready early, so other services could build on them.
- **Sonam Zangmo:** the template and contracts merged with all three approvals. routing-rules was built test-first with 100% coverage on the domain.
- **Nyendrak Yoezer Zangmo:** the database schema and EVM baseline gave every owner a shared reference. status-lifecycle was built test-first with every valid and invalid status change tested. The first EVM log used real Jira hours.
- **Gayley Choden:** all 5 cards merged. identity was built test-first with over 87% coverage on domain, application and gRPC code. The PM docs match the proposal and the EVM baseline.

### What did not go well

- **Late blockers.** The template and contracts landed about a day late, so every service slice was squeezed into the last day (Zangmo, Nyendrak, Choki).
- **Late estimates and hours.** Estimates and hours were only entered at the end of the sprint, so the first EVM log was late and CPI and SPI look better than they are (Nyendrak, Gayley).
- **CI went red** after the template merged ("matched no packages"). Fixed in PR #16, but some runs failed first (Choki).
- **Template bug found late.** `go mod tidy` removed testify from the template; this was only found during SCRUM-74 (Zangmo).
- **Local setup cost time:** broken protoc installs and PATH, Docker Desktop paused, wrong registry name, port 9001 in use, k3d networking on Windows, a library trying to move go.mod to Go 1.26, and `make lint` failing on Windows because of CRLF line endings (Choki, Gayley).

### What to change in Sprint 2

| # | Action | Owner |
|---|---|---|
| A1 | Blocking cards (shared contracts, template changes) go first and are agreed at planning, so no PR waits on another | Sonam Zangmo, whole team |
| A2 | Every card gets an original estimate on day 1, before work starts. Everyone logs hours daily | Everyone |
| A3 | Revise the per-sprint PV (138 h vs 93.3 h, and the 360 h vs 280 h question) and update `schedule-evm.md` | Nyendrak, Gayley |
| A4 | Add `make template-check` to CI so template problems are caught on every PR | Sonam Zangmo |
| A5 | Run `make test` and `make lint` locally before every push | Everyone |
| A6 | Add a `.gitattributes` file so line endings are always LF | Gayley |
| A7 | Only one card In Progress per person at a time | Everyone |
