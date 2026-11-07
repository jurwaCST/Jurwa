# Test plan and test cases (SCRUM-112)

Owner: Gayley Choden (Scrum Master). Version 1.0, 8 November 2026, end of Sprint 3. Module: SWE302.

Every number here comes from a file in `docs/evidence/`. Where a test was not run, this plan says so. The quality objectives and rules behind it are in `docs/pm/quality-plan.md`; the requirement ids are in `docs/pm/requirements.md`.

## 1. Scope and objectives

**In scope:** the 20 Go services in `services/`, the shared packages in `pkg/`, the two serverless functions in `functions/`, the web app in `web/` and the report flow through the api-gateway.

**Out of scope:** the Kubernetes cluster itself, third-party tools (Traefik, cert-manager, Prometheus, Grafana) and the email provider.

**Objectives** (quality plan section 1):

| # | Objective | Result in Sprint 3 |
|---|---|---|
| Q1 | Reports reach the correct Category Head | Met: every category and location routes correctly (`tests/spec`, S-08) |
| Q2 | At least 70% coverage on routing, SLA and status-flow packages | Met: 10 of 10 critical packages at 98.2% or more (section 5.1) |
| Q3 | No report is silently forgotten | Met for the domain: every status pair tested (S-10), 9 gopter properties hold (P-01 to P-09), Saga tests pass (SG-01 to SG-05). The full flow in the browser is not yet tested (section 7) |
| Q4 | A student submits a report with a photo in under two minutes | Not measured here: pilot usability sessions (Sonam Choki) |
| Q5 | The PWA is accessible | **Not run yet** (axe-core and Lighthouse, section 7) |
| Q6 | Report submission holds up under load | Met: p95 252 ms, 0 errors at 50 users (L-01) |
| Q7 | No untested code reaches `main` | Met: CI runs unit tests, the coverage gate and image builds on every pull request; branch protection blocks the merge when a check fails |

## 2. Test pyramid

| Level | What it checks | Tool | Owner | Where | Card |
|---|---|---|---|---|---|
| Unit | Domain and application logic of every service, with fakes and in-memory adapters | Go `testing`, Testify, table-driven tests | Each service owner | `services/*/**/*_test.go`, `pkg/`, `functions/` | every service card |
| Interaction (mocks) | A port is called with the right arguments, the right number of times | gomock | Sonam Zangmo | `pkg/resilience`, `services/sla-tracking` | SCRUM-121, SCRUM-92 |
| Integration | Adapters against a real PostgreSQL, MongoDB and Redis | Testcontainers | Each adapter owner; CI by Gayley | `*_integration_test.go` (build tag `integration`) | SCRUM-106 |
| Specification-based | Equivalence classes and boundary values written from the requirements | Go table-driven tests | Sonam Zangmo | `tests/spec/` | SCRUM-93 |
| State transition | Every pair of statuses, by a person and by the system | Go table-driven tests | Sonam Zangmo | `tests/spec/state_transition_test.go` | SCRUM-93 |
| Structural | Statement coverage, gate of 70% on critical packages | `go test -cover`, `scripts/check-coverage.sh` | Sonam Choki | `.github/workflows/quality-gate.yml` | SCRUM-111 |
| Property-based | SLA time-window rules hold for random times and windows | gopter | Nyendrak | `tests/property/` | SCRUM-107 |
| Saga | Escalation and compensation across services on the event bus | Go tests with in-memory bus | Nyendrak | `tests/saga/` | SCRUM-89 |
| BDD | The report lifecycle in Gherkin, mapped to the unit tests that check it | Gherkin (not executed yet) | Sonam Choki | `tests/bdd/` | SCRUM-94 |
| End-to-end | A real browser on the web app, through the gateway, to the services | Cypress 15.21.1 | Gayley | `web/cypress/` | SCRUM-108 |
| Web unit | Pages, role guards and the API client | Vitest, Testing Library | Gayley, Nyendrak | `web/src/**/*.test.tsx` | SCRUM-118, SCRUM-35 |
| Performance | `POST /reports` under load, and the rate limit | Apache JMeter | Sonam Zangmo | `tests/load/` | SCRUM-109 |
| Accessibility | WCAG basics on the main screens | axe-core, Lighthouse | Gayley | not written yet | Step 7 |
| Exploratory | The running app, off script, with a charter | Charter | Gayley (Sprint 3) | `docs/testing/exploratory-charter-sprint3.md` | SCRUM-112 |
| Code quality | Bugs, vulnerabilities, smells, duplication | SonarQube Community Build | Sonam Choki | `sonar-project.properties` | SCRUM-111 |

## 3. Test cases per level

Ids are used in the traceability matrix (section 4). Each row is a real test in the repo.

### 3.1 Unit (selected; 677 unit test functions in total, section 5.2)

| Id | What it checks | Input | Expected | Where |
|---|---|---|---|---|
| U-01 | Sign-in issues and checks a JWT | a user, a secret | token parses back to the same user and role; bad tokens refused | `services/identity/domain` `TestIssueAndParse`, `TestParseRejectsBadTokens` |
| U-02 | Passwords are hashed and checked | a password | right password accepted, wrong refused, weak refused | `services/identity/domain` `TestCheckPassword`, `TestValidatePassword` |
| U-03 | Every role against every action | 5 roles and no role, x all actions | allowed only where the policy says; unknown is denied | `services/authorization/domain` `TestEveryRoleAgainstEveryAction`, `TestDefaultDeny` |
| U-04 | The gateway rate limit | a burst over the limit | the extra requests are blocked, then the bucket refills | `services/api-gateway/domain` `TestLimiterAllowsBurstThenBlocks`, `TestLimiterRefills` |
| U-05 | A student reads only their own report | two student ids | owner allowed, other student refused | `services/api-gateway/domain` `TestCanReadReport` |
| U-06 | Location must come from the list | known and unknown locations | known accepted, unknown refused | `services/location-taxonomy/domain` `TestCheckLocation` |
| U-07 | The timeline is oldest first and a student sees only their reports | events out of order; two reporters | ordered timeline; own reports only, newest first | `services/report-history/application` `TestTimelineIsOldestFirstEvenWhenRecordedOutOfOrder`, `TestListReportsByReporterNewestFirstAndOnlyOwn` |
| U-08 | Routing ignores the description and routes by category and location | a report event | the right Category Head; dry run publishes nothing | `services/routing-engine/application` `TestRouteReport`, `TestRouteReportDryRunPublishesNothing` |
| U-09 | SLA window per category | each category | a window of at least 1 hour | `services/sla-policy/domain` `TestNewPolicy`, `TestPolicy_Window` |
| U-10 | Reassignment rules | a head, a staff user, a HoD | allowed reassignments pass, others refused | `services/reassignment/domain` `TestReassign`, `TestCheckCanReassign` |
| U-11 | Internal notes are not for students | each role | Category Head and Staff add; students refused | `services/internal-notes/domain` `TestWhoMayUseNotes` |
| U-12 | Every audit event is accepted and kept | each action | stored with its details | `services/audit-log/domain` `TestAllActionsAccepted`, `TestNewEvent_Boundaries` |
| U-13 | A notification is sent once | a status change | the sender is called once and stored as sent | `services/notification/application` `TestSendCallsSenderOnceAndStoresAsSent` |
| U-14 | Feedback rating and comment limits | ratings -1 to 6, long comments | 1 to 5 accepted, -1, 0 and 6 refused | `services/feedback/domain` `TestRatingBoundaries`, `TestCommentBoundaries` |
| U-15 | Dashboard counts | a sequence of events | totals per status and category, average resolution time, period filter | `services/analytics/domain` `TestSummaryAfterASequenceOfEvents`, `TestCategoryBreakdown`, `TestPeriodFilter` |
| U-16 | The dashboard is for the System Administrator and the President only | each role's token | 200 for those two, 403 for others, 401 without a token | `services/api-gateway/adapters/in/http` `TestDashboardRoles`, `TestDashboardRoutesNeedAToken` |
| U-17 | Assigned reports use the id from the JWT | a Category Head token | only that head's reports; other roles 403 | `services/api-gateway/adapters/in/http` `TestAssignedReportsUseTheIDFromTheJWT`, `TestAssignedReportsRoles` |
| U-18 | The escalation checker escalates each lapsed report | lapsed and not lapsed timers | each lapsed one escalated once, stops when sla-tracking is down | `functions/sla-escalation-checker/checker` `TestRun_EscalatesEachLapsedReport`, `TestRun_SlaTrackingDown` |
| U-19 | The digest sends nothing when there is nothing open | no open or escalated reports | no email | `functions/digest-dispatcher` `TestNoOpenOrEscalatedReportsSendsNothing` |
| U-20 | Metrics and JSON logs for every gRPC call | one call | one counter, one duration, one log line with a correlation id | `pkg/observability` `TestUnaryCountsOneRequestWithItsLabels`, `TestUnaryWritesOneJSONLogLine` |

### 3.2 Specification-based and state transition (`tests/spec/`, 15 tests, details in `docs/testing/spec-tests.md`)

| Id | What it checks | Input | Expected | Where |
|---|---|---|---|---|
| S-01 | Description boundaries | 0, spaces, 1, 499, 500, 501 characters; 500 and 501 Dzongkha characters | 1 to 500 valid; 0, spaces and 501 invalid; counted in characters, not bytes | `TestDescriptionBoundaries` |
| S-02 | Category classes | empty, HOSTEL, ACADEMIC_BLOCK, OTHERS, LIBRARY, hostel | the three valid; others invalid | `TestCategoryClasses` |
| S-03 | Location classes | missing, spaces, no building, building only, full | building required, wing and area optional | `TestLocationClasses` |
| S-04 | Photo id classes | none, UUID, empty, blank | none or UUID valid | `TestPhotoClasses` |
| S-05 | Reporter id | missing, not a UUID, UUID | UUID only | `TestReporterClasses` |
| S-06 | Photo size boundaries (each run for JPG and PNG) | 0 bytes, 1 byte, the signature only, 5 MB - 1, exactly 5 MB, 5 MB + 1 | from the signature up to exactly 5 MB valid; empty, 1 byte and over 5 MB invalid | `TestPhotoSizeBoundaries` |
| S-07 | Photo type classes | JPG, PNG, ` IMAGE/PNG `, GIF, PDF, no type, a PNG sent as JPG, a PDF sent as PNG | JPG and PNG only, and the bytes must match the type | `TestPhotoTypeClasses` |
| S-08 | Routing for every category and location | all categories and locations | Hostel to DSA, Academic Block to DAA, Others to SSO | `TestRoutingEveryCategoryAndLocation`, `TestRoutingUnspecifiedCategory` |
| S-09 | The description never changes routing | same report, different descriptions | same Category Head | `TestDescriptionNeverChangesRouting` |
| S-10 | Every status pair, by a person | all from-to pairs | only OPEN, ASSIGNED, IN_PROGRESS, RESOLVED in order; no skipping or going back | `TestTransitionTableByUser` |
| S-11 | Every status pair, by the system | all from-to pairs | the system alone sets ESCALATED; compensation only back to the previous status | `TestTransitionTableBySystem`, `TestCompensationOnlyToPrevious` |
| S-12 | Invalid status values and the full path | unknown statuses; OPEN to RESOLVED | unknown refused; the full path works | `TestInvalidStatusValues`, `TestFullLifecyclePath` |

### 3.3 Property-based (`tests/property/`, gopter, 100 random cases each, details in `docs/testing/property-tests.md`)

| Id | Property | Where |
|---|---|---|
| P-01 | The deadline is never before the start time | `sla_window_test.go` |
| P-02 | Deadline minus start equals the window for the category | `sla_window_test.go` |
| P-03 | A timer is due exactly when now is at or after the deadline | `sla_window_test.go` |
| P-04 | `Lapse` succeeds exactly when due, and only once | `sla_window_test.go` |
| P-05 | A longer window never gives an earlier deadline | `sla_window_test.go` |
| P-06 | Starting later never gives an earlier deadline | `sla_window_test.go` |
| P-07 | A stopped timer is never due | `sla_window_test.go` |
| P-08 | Every window of 1 hour or more is accepted | `sla_window_test.go` |
| P-09 | Every window of 0 hours or less is refused | `sla_window_test.go` |

### 3.4 Saga (`tests/saga/`)

| Id | What it checks | Expected | Where |
|---|---|---|---|
| SG-01 | An SLA lapse escalates the report and hands it up | status ESCALATED, new assignee | `TestSaga_LapseEscalatesReportAndHandsItUp` |
| SG-02 | A report resolved meanwhile | the escalation is undone | `TestSaga_ResolvedReport_EscalationIsUndone` |
| SG-03 | The status step fails | the previous assignee is back (compensation) | `TestSaga_StatusStepFails_PreviousAssigneeIsBack` |
| SG-04 | The same event twice | escalated once | `TestSaga_RepeatedEvent_EscalatesOnce` |
| SG-05 | One run | one correlation id through every step | `TestSaga_OneCorrelationIDThroughTheRun` |

### 3.5 Integration (Testcontainers, 19 tests in 9 files)

| Id | What it checks | Where |
|---|---|---|
| I-01 | Each repository passes the same contract on a real database as in memory | `*_integration_test.go` in analytics, authorization, category-taxonomy, feedback, identity, notification, report-submission (MongoDB), routing-rules |
| I-02 | Publish and consume on a real Redis Streams | `pkg/eventbus/redis_integration_test.go` |

### 3.6 End-to-end (`web/cypress/e2e/report-flow.cy.ts`)

| Id | What it checks | Expected |
|---|---|---|
| E-01 | A student signs in through the UI | the header shows the signed-in name |
| E-02 | A wrong password | an error on the page, still on `/login` |
| E-03 | A Hostel report is submitted and read back | 201, status OPEN, same category and location |
| E-04 | A 501 character description | 400 from the server |
| E-05 | A photo is uploaded, attached and viewed | 201 with a photo id; the report keeps it; `image/png` with `nosniff` |
| E-06 | A 5.5 MB photo | 400 "photo is larger than 5 MB" |
| E-07 | Another student views the photo | 403 |
| E-08 | Another student reads the report | 403 |

### 3.7 Performance (`tests/load/`)

| Id | What it checks | Input | Expected | Result |
|---|---|---|---|---|
| L-01 | `POST /reports` under load | 50 users, 2 minutes, 1 report a second each, rate limit raised to 1000 | p95 under 500 ms, under 1% errors | Met (section 5.5) |
| L-02 | The rate limit | 5 users at 10 a second from one IP, default limit 20 | about 20 a second accepted, the rest 429 | Met (section 5.5) |

### 3.8 Web unit (Vitest)

| Id | What it checks | Where |
|---|---|---|
| W-01 | Role guards: each page only for its roles, 403 page otherwise | `web/src/App.test.tsx` |
| W-02 | API client: token, base address, error bodies | `web/src/api/client.test.ts` |
| W-03 | The report form: fields, limits and messages | `web/src/pages/ReportForm.test.tsx` |

## 4. Traceability matrix

| Requirement | Tests |
|---|---|
| FR-01 sign in with email and password, JWT | U-01, U-02, E-01, E-02, I-01 (identity) |
| FR-02 submit a report | E-03, E-05, W-03, L-01 |
| FR-03 category one of three | S-02 |
| FR-04 location from the list | S-03, U-06 |
| FR-05 description at most 500 characters | S-01, E-04, W-03 |
| FR-06 photo JPG or PNG, at most 5 MB | S-06, S-07, E-05, E-06 |
| FR-07 unique id and status OPEN | E-03, S-12 |
| FR-08 own reports and status history | U-07, U-05, E-08 |
| FR-09 notified on a status change | U-13 (unit only) |
| FR-10 rate the resolution | U-14 (unit only) |
| FR-11 route by category and location only | S-09, U-08 |
| FR-12 Hostel to DSA, Academic Block to DAA, Others to SSO | S-08, I-01 (routing-rules) |
| FR-13 SLA window per category | U-09, P-02, P-08, P-09 |
| FR-14 SLA timer starts and stops | P-01, P-03, P-07 |
| FR-15 scheduled function escalates | U-18, P-04, SG-01, SG-04 |
| FR-16 compensation (Saga) | SG-02, SG-03, S-11 |
| FR-17 audit log | U-12 |
| FR-18 Category Head queue | U-17 |
| FR-19 reassign | U-10, SG-01 |
| FR-20 status changes, invalid ones refused | S-10, S-12 |
| FR-21 internal notes hidden from students | U-11 |
| FR-22 staff see only their reports | U-17 |
| FR-23 staff update status and add notes | S-10, U-11 |
| FR-25 manage routing rules | I-01 (routing-rules); unit tests in `services/routing-rules` |
| FR-26 manage users, roles, departments | U-03; unit tests in `services/identity` and `services/department-directory` |
| FR-27 manage categories and locations | unit tests in `services/category-taxonomy` and `services/location-taxonomy` |
| FR-28 manage SLA windows | U-09, P-08, P-09 |
| FR-29 read-only dashboard | U-15, U-16 |
| FR-30 the President changes nothing | U-03 |
| FR-31 periodic digest | U-19 (unit only) |
| NFR-02 load test | L-01, L-02 |
| NFR-03 health checks | `TestUnaryCountsHealthChecksWithoutALogLine` (U-20); probes are in the manifests |
| NFR-04 JWT and role checks | U-03, U-05, U-16, U-17, E-07, E-08 |
| NFR-05 hashed passwords, rate limit | U-02, U-04, L-02 |
| NFR-08 JSON logs and metrics | U-20 |
| NFR-10 70% coverage on critical code | coverage gate (section 5.1) |

**Requirements with no test:**

| Requirement | Why |
|---|---|
| FR-24 staff represented by their supervisor (Could) | Not built |
| NFR-01 submit with a photo in under two minutes | Measured by the pilot usability sessions, not by an automated test |
| NFR-06 WCAG basics | axe-core and Lighthouse not run yet (Step 7) |
| NFR-07 installable PWA | The PWA manifest and service worker (SCRUM-98) are not merged, so there is nothing to test |
| NFR-09 hexagonal template, own database, gRPC only | Checked in code review, not by a test |
| NFR-11 free tools only | A project rule, not testable |

**Tested only in parts:** FR-09, FR-10 and FR-31 have unit tests but no end-to-end test. FR-25 to FR-28 have domain tests, but the admin screens (SCRUM-37) were moved to Sprint 4. FR-08's timeline and FR-18 to FR-20 are not yet in the Cypress story (section 7). FR-15's schedule (Knative PingSource, SCRUM-104) is not tested.

## 5. Results summary

### 5.1 Coverage of the critical packages (`docs/evidence/gayley/sprint3/all-tests.txt`)

| Package | Coverage |
|---|---|
| routing-rules domain, application | 100.0%, 100.0% |
| routing-engine domain, application | 100.0%, 100.0% |
| sla-policy domain, application | 100.0%, 100.0% |
| sla-tracking domain, application | 100.0%, 98.2% |
| status-lifecycle domain, application | 100.0%, 100.0% |

10 of 10 at or above the 70% target (NFR-10). The same gate runs in CI on every pull request (`docs/evidence/sonam-choki/sprint3/coverage-gate.txt`). Every other service's `domain` and `application` package is at 82.7% or more (lowest: identity application).

### 5.2 Number of tests

| Kind | Count | Result | Evidence |
|---|---|---|---|
| Go unit test functions (subtests not counted) | 677 | 140 packages pass, 0 fail | `all-tests.txt` |
| Go integration test functions (tag `integration`) | 19 | Pass in CI | `docs/evidence/gayley/sprint3/integration-all.txt`, `integration-ci.png` |
| Specification-based (`tests/spec`) | 15 | Pass | `all-tests.txt` |
| Property-based (`tests/property`) | 9 properties x 100 cases | Pass, no counter-example | `docs/evidence/nyendrak/sprint3/gopter.txt` |
| Saga (`tests/saga`) | 5 | Pass | `all-tests.txt` |
| Web unit (Vitest) | 47 | Pass | `docs/evidence/gayley/sprint3/web-tests.txt` |
| End-to-end (Cypress) | 8 | 8 of 8 pass | `docs/evidence/gayley/sprint3/cypress-run.txt`, `cypress-run.mp4` |
| BDD scenarios (Gherkin) | 14 | Not executed; each mapped to unit tests | `tests/bdd/README.md` |

The test functions were counted with `go test -list '.*' ./...` (677) and with `-tags=integration` (696).

### 5.3 SonarQube (scan of 6 Nov 2026, `docs/testing/sonarqube.md`)

| Measure | Result |
|---|---|
| Security | 3 open issues, rating C |
| Reliability | 4 open issues, rating C |
| Maintainability | 5 open issues, rating A |
| Coverage | 61.7% on 5.3k lines to cover |
| Duplications | 7.8% on 19k lines |
| Security Hotspots | none |

No issue is in a Go service or a critical package. Six are in the learning-spike files (never deployed), one is in `web/src/api/client.ts` (D-10).

### 5.4 Property-based (gopter)

9 properties, 100 random cases each, all passed, no counter-example (`gopter.txt`). Holidays and weekends are not tested: the domain counts calendar hours (the holiday-calendar service was removed with change C1).

### 5.5 Performance (JMeter, `docs/evidence/sonam-zangmo/sprint3/jmeter-report-submission.txt`)

| Run | Requests | Errors | p95 | Throughput | Target |
|---|---|---|---|---|---|
| L-01 load: 50 users, 120 s, rate limit 1000 | 5,617 | 0 (0.00%) | 252 ms | 47.3 a second | p95 under 500 ms, under 1% errors: **met** |
| L-02 rate limit: 5 users, 15 s, default limit 20 | 716 | 408 x 429 | 8 ms | 49.2 a second | 308 accepted, about 20 a second: **the limit works** |

All 50 users come from one laptop, so the rate limit was raised for L-01; on campus each student has their own IP.

### 5.6 End-to-end (Cypress)

8 of 8 pass in Edge 157, 4 seconds (`cypress-run.txt`, video `cypress-run.mp4`). Not in CI: it needs the whole stack.

### 5.7 Accessibility (axe-core, Lighthouse)

**Not run yet.** No scores to report (section 7).

## 6. Defect log

Defects found in Sprints 2 and 3.

| # | Defect | How it was found | Fixed by | Card |
|---|---|---|---|---|
| D-01 | internal-notes let the System Administrator add notes, against `api-routes.md` and the authorization policy | Review after the merge (issue I8) | Sonam Choki | SCRUM-116 |
| D-02 | The internal-notes contract had no role, so students could not be denied; the gateway must send `x-user-role` | Design check (issue I8) | Sonam Zangmo | SCRUM-116, SCRUM-117 |
| D-03 | The feedback Dockerfile did not copy `pkg/`, so its image did not build | Image build in CI | Sonam Choki, Gayley | PR #55 |
| D-04 | The identity, authorization, notification and analytics Dockerfiles did not copy `pkg/` after they started using `pkg/observability` | Image build for SCRUM-105 | Gayley | SCRUM-105 |
| D-05 | The reassignment, sla-tracking and status-lifecycle Dockerfiles did not copy `pkg/` after the Saga, so every pull request had 3 red checks | The Images check on every pull request | Nyendrak | SCRUM-89 (PR #93) |
| D-06 | Passing health checks wrote a log line every few seconds and filled the logs | Live run of the observability work | Gayley (tests first) | SCRUM-105 |
| D-07 | The gateway retries calls that create something after a timeout, so a retry can make a duplicate or fail with 409 | JMeter first attempt | Open, Sonam Zangmo | SCRUM-132 |
| D-08 | report-submission saves the report, then publishes the event; if the publish fails the student gets an error but the report is saved, so a retry makes a duplicate | Code review of SCRUM-125 | Open, noted for Sprint 4 | SCRUM-125 |
| D-09 | The dashboard time example `+06:00` gives 400 when typed into a URL, because `+` is read as a space | Code review of SCRUM-129 | Note asked of Sonam Zangmo | SCRUM-129 |
| D-10 | The regular expression `/\/+$/` in the web API client can be slow on a long run of slashes | SonarQube | Open, Gayley (low risk: the input is the configured API address) | SCRUM-111 |
| D-11 | Cypress hung with no error with TypeScript 7 and Node 26 | Running the end-to-end tests | Gayley: esbuild compiler, plain JavaScript config | SCRUM-108 |
| D-12 | A new report was not in report-history, so its timeline answered 404 | End-to-end check with curl | Sonam Choki: report-history reads the events | SCRUM-130 |
| D-13 | Stopping the end-to-end stack by killing bash left the services running and holding ports | Ports taken on the next run | Gayley: documented, Ctrl+C stops all | SCRUM-108 |
| D-14 | The first location-taxonomy slice was not test-first | Review | Sonam Choki: reverted and redone | issue I9 |

## 7. Known gaps and risks

| Gap | Effect | Plan |
|---|---|---|
| Accessibility (axe-core, Lighthouse) not run | NFR-06 and Q5 not shown | Step 7, now that the report form (SCRUM-35) is merged |
| PWA manifest and service worker not merged | NFR-07 cannot be tested; Traefik's HTTPS works but a self-signed certificate blocks a phone install | SCRUM-98 (Nyendrak) |
| The Cypress story stops before the form, routing, status change and timeline steps | The full flow is tested by parts, not end to end in one browser test | Add them now that SCRUM-35, SCRUM-128, SCRUM-131 and SCRUM-136 are merged |
| BDD scenarios are not executed | They document behaviour; the unit tests check it | godog step definitions, Sprint 4 |
| Integration tests cover 8 services and the event bus, not all 20 | The other adapters are tested only in memory | The remaining database adapters are Sprint 4 |
| Knative schedule (SCRUM-104) not tested | The escalation checker's logic is tested, not its schedule | CronJob fallback agreed |
| Jaeger tracing not done | No traces; the correlation id in the logs is used instead | Stretch |
| Cypress and JMeter are not in CI | They depend on runs on a laptop | Evidence files are kept for each run |
| D-07 and D-08 open | Duplicate reports are possible when a service is slow | SCRUM-132, Sprint 4 |
