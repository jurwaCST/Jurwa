# SCRUM-92: unit tests and TDD for routing, SLA and status flow

Card: done when routing, SLA policy and status-flow logic are written test-first with Go testing, Testify and gomock, reaching at least 70% coverage on those packages (quality plan Q2 and section 6).

## Coverage (target at least 70%)

From `scrum-92-coverage.txt`, run on `main` on 2 Nov 2026.

| Package (quality plan section 6) | Owner | Coverage |
|---|---|---|
| `services/routing-rules/domain` | Sonam Zangmo | 100% |
| `services/routing-engine/domain` | Sonam Zangmo | 100% |
| `services/sla-policy/domain` | Nyendrak | 100% |
| `services/sla-tracking/domain` | Nyendrak | 100% |
| `services/status-lifecycle/domain` | Nyendrak | 100% |
| Total of the five | | 100% |

The application and adapter packages of the same services are between 97.1% and 100%. Only the `cmd/server` packages are low (2.9% to 9.4%), because `main()` starts a real server and is checked with grpcurl instead.

## Test-first evidence

| Card | Evidence |
|---|---|
| SCRUM-90 routing-engine | commit `4c6f005 SCRUM-90 tests first`, `docs/evidence/sonam-zangmo/scrum-90-tests-first-red.txt` |
| SCRUM-91 routing resilience | `docs/evidence/sonam-zangmo/scrum-91-tests-first-red.txt` |
| SCRUM-86 sla-tracking | `scrum-86-tests-first-fail.txt` |
| SCRUM-88 reassignment | `scrum-88-tests-first-fail.txt` |
| SCRUM-27 sla-escalation-checker | `scrum-27-tests-first-fail.txt` |
| SCRUM-42 audit-log | `scrum-42-tests-first-fail.txt` (in PR for SCRUM-42) |

SCRUM-19 status-lifecycle and SCRUM-29 sla-policy were also written test first, but the failing run was not saved, so there is no file for them.

## Tools

- Go `testing` and Testify (`assert`, `require`) in every test, table-driven for the rules (transitions, SLA windows, timer due times, roles).
- Test doubles: every service uses small hand-written fakes for its ports (repositories, publishers, gRPC clients), and `bufconn` for real gRPC calls in memory. **gomock is not used yet.** The quality plan lists it, but adding it changes the shared `go.mod`, so it is a team decision: either add `go.uber.org/mock` for the ports, or change the quality plan to say hand-written fakes.
