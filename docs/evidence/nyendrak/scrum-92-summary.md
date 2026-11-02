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
- Test doubles, as the team agreed on 2 Nov: hand-written fakes are the default (they check results), and gomock (`go.uber.org/mock` v0.6.0) is used for a few ports where the calls themselves matter. `bufconn` is used for real gRPC calls in memory.

## gomock in sla-tracking

Mocks are generated from `services/sla-tracking/ports/ports.go` into `ports/mocks/` (`go generate ./services/sla-tracking/ports/`). The tests are in `services/sla-tracking/application/service_gomock_test.go`:

| Test | What it checks that a fake cannot |
|---|---|
| `TestListDueTimers_PublishesThenSavesInOrder` | For each due timer, `sla.lapsed` is published before the timer is saved as LAPSED, one timer at a time (`gomock.InOrder`) |
| `TestListDueTimers_PublishFails_NothingSaved` | When publishing fails, `Update` is never called and the next timer is not touched |
| `TestListDueTimers_BadLimit_NoCalls` | Bad input is rejected before any port is called |
| `TestStartTimer_AsksPolicyOnceThenCreates` | sla-policy is asked exactly once, then the timer is created with that deadline |
| `TestStartTimer_PolicyDown_NoCreate` | If sla-policy fails, no timer is created |

Evidence: `scrum-92-gomock-tests.txt` (all pass), and `scrum-92-gomock-catches-wrong-order.txt`: saving before publishing was put in on purpose, the order test failed with "should be called before", and the mistake was removed.
