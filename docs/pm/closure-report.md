# Closure report: Jurwa (SCRUM-85)

Owner: Gayley Choden (Scrum Master). Draft, 9 November 2026. Module: SWE303.

Every number below has a source file in the repository, named next to it. Parts marked **Waiting for** need a number that does not exist yet; they are filled in before this report is merged.

## 1. Objectives (charter section 3)

| # | Objective | Result | Evidence |
|---|---|---|---|
| 1 | 20 microservices that route a report to the correct Category Head with 100% accuracy on the test rule-set | **Met.** 20 services in `services/`. Every category and location routes correctly (Hostel to DSA, Academic Block to DAA, Others to SSO), and the description never changes the route | `tests/spec/routing_rules_test.go`; `docs/testing/test-plan.md` S-08, S-09 |
| 2 | An installable PWA in which a student submits a report with a photo in under two minutes, verified by usability testing | **Partly met.** The report form works with a photo, end to end (Cypress 8 of 8) and is accessible (axe-core 0 violations, Lighthouse 100). The PWA is not installable yet: the manifest and service worker (SCRUM-98) are not merged. **Waiting for** the pilot result: the average and slowest time against 120 seconds (Sonam Choki, `docs/testing/pilot-usability-report.md`) | `docs/evidence/gayley/sprint3/cypress-run.txt`, `docs/testing/accessibility.md` |
| 3 | Automatic SLA escalation by a scheduled serverless function, shown in the final demo | **Partly met.** The escalation Saga and its compensation pass their tests (SG-01 to SG-05). The checker runs as a Kubernetes CronJob every minute instead of a Knative PingSource (risk R3 fallback, SCRUM-104). **Waiting for** the demo: whether a real escalation is shown live | `tests/saga/`, `deploy/knative/cronjob-fallback.yaml`, `docs/evidence/nyendrak/sprint3/cronjob-run-once.txt` |
| 4 | A test pyramid tracked in SonarQube, at least 70% coverage on routing, SLA and status-flow code | **Met.** 10 of 10 critical packages at 98.2% or more; 696 Go test functions, 47 web tests, 8 end-to-end, 9 gopter properties, JMeter p95 252 ms with 0 errors. SonarQube: coverage 61.7%, Security 3, Reliability 4, Maintainability 5 issues, duplications 7.8%, no issue in a Go service | `docs/testing/test-plan.md` section 5, `docs/testing/sonarqube.md` |
| 5 | Containers on Kubernetes (k3d) through GitOps (GitHub Actions and Argo CD), shown in a live demo | **Met for the pipeline.** CI builds and tests every pull request; Argo CD syncs the cluster from `main` and heals a change made by hand (api-gateway scaled to 0, back to 1 in 2 seconds). **Waiting for** the live demo | `.github/workflows/`, `deploy/argocd/`, `docs/evidence/sonam-zangmo/sprint3/argocd.txt` |
| 6 | Within the schedule and a zero budget, tracked weekly with EVM | **Budget met:** only free and open-source tools. **Schedule:** see section 3 | `docs/pm/evm/` |

## 2. Scope

**Built: 20 services and 2 serverless functions, not the 25 services of the first proposal.**

| Change | From | To | Why | Source |
|---|---|---|---|---|
| C1 | 25 services | 20 services: export, holiday-calendar, digest, email-template and user-profile removed; digest emails are sent by the `digest-dispatcher` function | Less time (C2). Risk R4: the lowest-priority services are cut first; testing and documentation are never cut | `docs/pm/schedule-evm.md` section 5 |
| C2 | 8 weeks | 3 five-day sprints, presentation about 9 November | The presentation date moved | `docs/pm/schedule-evm.md` section 5 |
| C3 | 6 or 7 services per member | 5 services per member, plus one function for two members | An even load | `docs/pm/schedule-evm.md` section 5 |

Holidays and weekends are therefore not counted in SLA windows (the holiday-calendar service was one of the five removed).

**Moved to Sprint 4 (not built):** the staff ticket view (SCRUM-36) and the admin dashboard (SCRUM-37), the other database adapters, Jaeger tracing, the godog runner for the BDD scenarios, and the President dashboard page (SCRUM-99, Stretch). The routes behind the staff and dashboard pages exist and are tested.

## 3. Schedule and cost (EVM)

| Sprint | PV | EV | AC | CPI | SPI | Source |
|---|---|---|---|---|---|---|
| 1 (22 to 27 Oct) | 93.3 h | 138 h | 96 h | 1.44 | 1.48 | `docs/pm/evm/week1.md` |
| 2 (28 Oct to 3 Nov) | 93.3 h | 135 h | 129 h | 1.05 | 1.45 | `docs/pm/evm/week2.md` |
| 3 (4 to 8 Nov) | **Waiting for** the Sprint 3 EVM log (Nyendrak) | | | | | `docs/pm/evm/week3.md` |
| Cumulative after Sprint 2 | 186.7 h | 273 h | 225 h | 1.21 | 1.46 | `docs/pm/evm/week2.md` |

Read honestly:
- **Estimates came late.** In Sprint 1, estimates and hours were entered only at the end of the sprint (issue I4), so its CPI of 1.44 is optimistic. Sprint 2 was much closer to 1 (1.05), once three of the four members entered planned hours before starting.
- **SPI is high mainly because the baseline is too low.** The baseline gives each sprint 93.3 h, while the team planned about 135 h per sprint. The action to revise the per-sprint PV (A3) was still not done after Sprint 2.

## 4. Risks and issues at the end

| Risk | What happened | Status at the end | Source |
|---|---|---|---|
| R1 scope grows | Scope went down, not up: 25 to 20 services (C1) | Closed: did not happen | `risk-register.md` 2.2 |
| R2 learning curve | Local setup cost time in Sprint 1. Sprint 2 services were built from the template without new tool problems; Sprint 3 had new Windows problems with Testcontainers and Cypress, solved and documented | Reduced | `risk-register.md` 2.2 |
| R3 cluster or Knative unstable near the demo | Knative was not used; the CronJob fallback runs. Argo CD deploys the cluster. **Waiting for** Sonam Choki's R3 update (SCRUM-83) to reach `main` and the rehearsal | Reduced | `risk-register.md`; SCRUM-83 |
| R4 schedule slips | Work again landed on the last day of each sprint; Must cards were finished and Stretch cards cut, in the planned order | Happened, handled by cutting Stretch work | `risk-register.md` 2.2 |
| R5 staff without institutional email | Not built (FR-24 is a Could) | Open, carried to Sprint 4 | `test-plan.md` section 4 |

**Issue log at the end** (`docs/pm/issue-log.md`): I1, I3 and I9 closed; I2, I4 and I8 in progress; I5, I6 and I7 open. I6 (old Jira cards from the 25-service design) is closed by closing those cards as "Won't do".

**Defects:** 15 defects found in Sprints 2 and 3 are listed with how each was found and who fixed it in `docs/testing/test-plan.md` section 6. Two are open and carried to Sprint 4: retries after a timeout can create a duplicate (D-07, SCRUM-132) and report-submission saves before it publishes (D-08).

## 5. What was delivered

| Area | Where |
|---|---|
| Services and functions | `services/`, `functions/`, one README each |
| Contracts | `proto/` |
| Web app | `web/` |
| Deployment | `deploy/k8s/`, `deploy/argocd/`, `deploy/knative/`, `deploy/observability/` |
| CI | `.github/workflows/` |
| Tests | `tests/`, `web/cypress/`, `docs/testing/` |
| Project management | `docs/pm/`, `docs/sprint-reviews/` |
| Evidence per member | `docs/evidence/gayley/`, `docs/evidence/nyendrak/`, `docs/evidence/sonam-choki/`, `docs/evidence/sonam-zangmo/` |

## 6. Acceptance

**Waiting for** the final presentation and live demo (about 9 November) and the tutors' sign-off. The team accepts the delivery when the demo script (`docs/pm/demo-script.md`) runs end to end, or its recorded fallback is shown.

| Name | Role | Accepted |
|---|---|---|
| Sonam Zangmo | Team Lead | |
| Gayley Choden | Scrum Master | |
| Nyendrak | Developer | |
| Sonam Choki | Developer | |
