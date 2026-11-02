# Quality Management Plan (SCRUM-80)

Project: Jurwa, campus facility issue reporting PWA. Owner: Gayley Choden (Scrum Master). Version 1.0, 26 October 2026.

This plan describes how quality will be built in and checked. No automated tests exist yet; everything below is what the team will do from Sprint 1 onward.

## 1. Quality objectives

Taken from the proposal's objectives and testing approach.

| # | Objective | Measure | Checked by |
|---|---|---|---|
| Q1 | Reports reach the correct Category Head | 100% of cases in the routing test rule-set pass | Unit and specification-based tests on routing-engine and routing-rules |
| Q2 | Critical logic is well tested | At least 70% statement coverage on routing, SLA policy and status-flow packages | `go test -cover`, SonarQube |
| Q3 | No report is silently forgotten | Every valid and invalid status change is tested; an overdue report is always escalated | State transition tests, gopter property tests, Cypress end-to-end test |
| Q4 | A student can submit a report with a photo in under two minutes | Timed usability test with pilot users (Week 7) | Usability session notes |
| Q5 | The PWA is accessible | No critical axe-core violations; Lighthouse accessibility checked on every main screen | axe-core, Lighthouse |
| Q6 | Report submission holds up under load | JMeter load test on report submission, results recorded | Apache JMeter |
| Q7 | No untested or unreviewed code reaches `main` | Every pull request has passing unit tests and one approval | GitHub branch protection, CI |

## 2. Test pyramid

Many fast unit tests at the bottom, fewer slow end-to-end tests at the top.

| Level | What it covers | Tools | Runs |
|---|---|---|---|
| Unit | Domain logic in each service: routing, SLA policy, status flow, validation, identity | Go `testing`, Testify for assertions, table-driven tests; hand-written fakes and in-memory adapters by default, gomock for interaction checks on chosen ports | Every push, in CI |
| Integration | One service against real PostgreSQL, MongoDB and Redis | Testcontainers | Every pull request touching adapters |
| System and end-to-end | The full flow in the browser: submit, route, update, resolve, escalate | Cypress | Before each sprint review, from Sprint 2 |
| Non-functional | Load on report submission; accessibility of the PWA | Apache JMeter; axe-core and Lighthouse | Week 6 and 7 |

## 3. Tools and where each is used

Every testing tool named in the proposal.

| Tool | Where it is used |
|---|---|
| Go `testing` | All unit tests in every service (`*_test.go` next to the code) |
| Testify | Assertions only (`assert`, `require`) in all Go tests |
| Hand-written fakes and in-memory adapters | The default test doubles for ports: small Go structs in the test file (for example `fakeStatus` in feedback, `fakeSender` in notification) and the `adapters/out/memory` of services that store data, so domain and application logic is tested without a database or network |
| Table-driven tests | Specification-based tests: equivalence classes and boundary values for the report form (description about 500 characters, photo about 5 MB, required fields) and routing rules; state transition tables for the status flow |
| `go test -cover` | Structural coverage on routing, SLA and status-flow packages; target at least 70% |
| gomock | Only for interaction checks on chosen ports, where a test must prove a call happened with the right arguments, the right number of times or in the right order: sla-tracking (SCRUM-92) and retry in `pkg/resilience`. Not used for other port doubles |
| gopter | Property-based tests of SLA time-window logic (sla-policy, sla-tracking) with random timestamps, for example "a report older than its SLA window is always overdue" |
| Testcontainers | Integration tests that start real PostgreSQL, MongoDB and Redis containers for each service's adapters |
| Cypress | End-to-end tests of the PWA: submit a report, route it, update its status, resolve it |
| Apache JMeter | Load test on report submission through the api-gateway |
| axe-core | Automated WCAG checks on PWA screens |
| Lighthouse | Accessibility and PWA audit (installability) of the front end |
| SonarQube Community Build | Coverage, code smells and security issues; quality gate on critical packages |
| Structured charters | Manual exploratory testing: one member explores the build off-script each sprint and logs findings as Jira bugs |

## 4. Test-driven development (TDD)

Routing rules, SLA policy and the status flow are written test-first: write a failing test, write the smallest code that passes it, then refactor.

| Area | Services | Owner |
|---|---|---|
| Routing (Hostel to DSA, Academic Block to DAA, Others to SSO; category and location only, never the description) | routing-rules, routing-engine | Sonam Zangmo |
| SLA policy and timers | sla-policy, sla-tracking | Nyendrak |
| Status flow (Open, Assigned, In Progress, Resolved; Escalated set automatically) | status-lifecycle | Nyendrak |

Reviewers check that the test commits come before or with the code, not after.

## 5. Behaviour-driven development (BDD)

The report lifecycle is described in Gherkin scenarios, written from the student's and Category Head's point of view. They later become the Cypress end-to-end tests.

Example:

```gherkin
Feature: Report lifecycle

  Scenario: A hostel report is routed to the DSA and escalated when ignored
    Given a signed-in student
    When the student submits a report with category "Hostel" and location "Hostel A, Block 2, Wing B, Washroom"
    Then the report status is "Assigned"
    And the report is assigned to the DSA
    When the SLA window for the report passes with no status change
    Then the report status is "Escalated"
    And the escalation is recorded in the audit log
```

## 6. Coverage target

- At least 70% statement coverage on the domain packages of routing-rules, routing-engine, sla-policy, sla-tracking and status-lifecycle.
- Other services: unit tests for their domain logic are required before merge; no fixed number.
- Measured with `go test -cover` locally and in CI, tracked in SonarQube.

## 7. Code review rules

- Nobody commits to `main`. One branch and one pull request per Jira card, with the Jira key in the title.
- Every pull request needs one approval. Nobody merges their own pull request.
- `.github/CODEOWNERS` sends each pull request to the owner of the changed folder. Changes to `proto/` need Sonam Zangmo and the owners of every service that uses the contract.
- Reviewer pairs: Sonam Zangmo is reviewed by Nyendrak; Nyendrak by Sonam Zangmo (Gayley if busy); Sonam Choki by Gayley; Gayley by Sonam Choki.
- The reviewer checks: only the author's folders changed; no secrets, `.env` or `node_modules`; tests pass; the pull request template checklist is filled in.
- Review within one working day.

## 8. Definition of done

A card is Done when:

1. The code or document is on `main` through an approved, merged pull request.
2. Unit tests for the domain logic exist and pass (code cards).
3. CI is green (once CI exists).
4. Coverage target met for routing, SLA and status-flow packages.
5. The service README says what it does, its RPCs, and how to run and test it.
6. No secrets in the repo; configuration comes from environment variables.
7. The Jira card is moved to Done and hours are logged.

## 9. CI quality gates

Sonam Choki builds the GitHub Actions pipeline (SCRUM-101). A pull request cannot be merged unless:

| Gate | Check |
|---|---|
| Format and vet | `gofmt` reports no changes; `go vet` passes |
| Build | Every service compiles |
| Unit tests | `go test ./...` passes |
| Coverage | At least 70% on the critical packages listed in section 6 |
| Image | The service Dockerfile builds |
| SonarQube | Quality gate passes on critical packages (once SonarQube is set up) |

## 10. Defect handling in Jira

1. Anyone who finds a defect creates a Jira **Bug** with: steps to reproduce, expected result, actual result, evidence (screenshot or output), and the affected service.
2. Priority: **High** blocks the report flow or the demo; **Medium** wrong behaviour with a workaround; **Low** cosmetic.
3. The bug is assigned to the service owner (see README). High bugs are raised at the next stand-up and fixed in the current sprint.
4. The fix goes through a normal pull request that includes a test reproducing the bug.
5. The person who reported the bug confirms the fix before it is moved to Done.

## 11. Responsibilities

| Who | Quality responsibility |
|---|---|
| Gayley Choden (Scrum Master) | Owns this plan; checks the definition of done at sprint review; tracks open bugs |
| Sonam Zangmo (Team Lead) | TDD for routing; review of all `proto/` changes |
| Nyendrak | TDD and gopter tests for SLA policy, SLA tracking and status flow |
| Sonam Choki | CI pipeline and quality gates (SCRUM-101) |
| Every service owner | Unit tests for their own services; reviews for their assigned partner |
| Whole team | Exploratory charter (one member per sprint, rotating); pilot usability sessions in Week 7 |

Owners for Cypress, JMeter, axe-core, Lighthouse and SonarQube setup are assigned at Sprint 2 planning.

## 12. Reporting

Test results, coverage and open bugs are reported at each sprint review and recorded in `docs/sprint-reviews/`. Evidence (test output, screenshots) is saved in `docs/evidence/<name>/`.
