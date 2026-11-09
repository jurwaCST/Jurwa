# Risk Register (SCRUM-44)

Project: Jurwa. Owner: Gayley Choden (Scrum Master). Version 1.0, 27 October 2026.

R1 to R5 come from the proposal (section 7) and the charter. R6 to R9 are new risks found during Sprint 1. Probability and impact are rated Low, Medium or High. Owners for R1 to R5 are proposed by the Scrum Master and confirmed at the Sprint 1 review.

## 1. Register

| ID | Risk | Probability | Impact | Strategy | Response | Trigger (early warning) | Owner | Status |
|---|---|---|---|---|---|---|---|---|
| R1 | Scope grows beyond the agreed architecture | Medium | High | Avoid | Scope is locked in the charter; any change needs full-team sign-off, recorded as a Jira `decision` card (see `comms-plan.md`) | A card or pull request adds a service or feature not in the charter | Sonam Zangmo (Team Lead) | Open |
| R2 | Learning curve of new tools (Kubernetes, gRPC) delays the start | High | Medium | Mitigate | Learning spike before full implementation (SCRUM-71, done 26 Oct); shared service template (SCRUM-73) | A member cannot build or run the template | Gayley Choden | Reduced: spike merged and demoed, template works on members' laptops |
| R3 | Local Kubernetes cluster or Knative setup is unstable near the demo date | Medium | High | Mitigate | Rehearse deployment early; keep a recorded demo as fallback; if Knative fails, run the same function code as Kubernetes CronJobs | Cluster or Knative fails to start in a rehearsal | Sonam Choki (deployment) | Reduced: cluster deployed by Argo CD, CronJob fallback runs (section 2.3) |
| R4 | Schedule slips once the build is under way | Medium | High | Mitigate | Weekly EVM shows drift early; lowest-priority work is cut first; testing and documentation are never cut | SPI or CPI below 0.9 (see `schedule-evm.md`) | Nyendrak (EVM baseline) | Open |
| R5 | Non-teaching staff without institutional email (cooks, cleaners) cannot log in | Low | Medium | Mitigate | Their supervisor or HoD holds the account and acts for them, with a full audit trail (requirement FR-24) | Pilot users report staff who cannot be reached | Gayley Choden (identity) | Open |
| R6 | The compressed schedule (8 weeks reduced to three 5-day sprints, change C2) leaves too little time to build, test and deploy 20 services | High | High | Mitigate | Services already cut from 25 to 20 (change C1); one thin slice per owner in Sprint 1; push daily and open draft pull requests early; carry over consciously at sprint review rather than miss silently | Sprint 1 cards carried over; SPI below 0.9 at the end of Sprint 1 | Gayley Choden (Scrum Master) | Open |
| R7 | Shared work owned by one person blocks everyone: the template (SCRUM-73) and contracts (SCRUM-74) gate every service slice | High | High | Mitigate | Template and contracts are first in the sprint order; reviewed by all three others within a day; the Scrum Master chases daily; blocked members do no-code cards meanwhile | SCRUM-74 still not merged on the last day of Sprint 1 (observed 27 Oct) | Sonam Zangmo (owner of SCRUM-73 and 74) | Open |
| R8 | Jira has no original estimates and few logged hours, so EV and AC cannot be calculated for the EVM report | High | Medium | Mitigate | Every owner adds an estimate to each card before the sprint and logs hours by sprint end; the Scrum Master checks on day 4 of each sprint | Cards without estimates at sprint planning (observed 26 Oct, `schedule-evm.md` section 6) | Gayley Choden (Scrum Master) | Open |
| R9 | k3d on Windows writes `host.docker.internal` as the cluster address, which resolves to a LAN IP on the campus network, so `kubectl` cannot connect | Medium | Medium | Mitigate | Create clusters with `--api-port 127.0.0.1:<port>` or point kubectl at `127.0.0.1` (documented in `docs/learning-spike/README.md`) | `kubectl` timeouts after `k3d cluster create` | Sonam Choki (k3d setup) | Open: found in the learning spike, 26 Oct |

## 2. Sprint 2 review (SCRUM-83)

Reviewed by Sonam Choki on 2 November 2026 (end of Sprint 2). Ratings marked "proposed" are confirmed or changed by the team at the Sprint 2 review. Real problems that already happened are tracked in `issue-log.md`.

### 2.1 Existing risks

| ID | Review date | Probability now | Impact now | Happened? | Notes |
|---|---|---|---|---|---|
| R1 | 2 Nov 2026 | Low (proposed) | High | No | Scope went down, not up: 25 services cut to 20 (change C1). No card adds a service outside the charter |
| R2 | 2 Nov 2026 | Low (proposed) | Medium | Partly, in Sprint 1 | Local setup cost time in Sprint 1 (protoc and PATH, Docker Desktop paused, k3d networking on Windows, CRLF line endings). Sprint 2 services were built from the template without new tool problems |
| R3 | 2 Nov 2026 | Medium | High | Not yet | No real service is deployed to the cluster yet; only images are built (SCRUM-40). Rehearsal still needed in Sprint 3 |
| R4 | 2 Nov 2026 | High (proposed) | High | Yes | Sprint 2 work was squeezed into the last days again (for example SCRUM-23 was redone on 29 Oct and the other Choki services landed on 2 Nov). Week 2 EVM is not logged yet |
| R5 | 2 Nov 2026 | Low | Medium | No | Not testable until the pilot |
| R6 | 2 Nov 2026 | High | High | Yes | Two cards carried over from Sprint 1 (SCRUM-90, SCRUM-38). 16 services in one sprint is the same pressure |
| R7 | 2 Nov 2026 | Medium (proposed) | High | Yes, in Sprint 1 | Template and contracts landed about a day late (issue I1). In Sprint 2 they were already merged, so services could start on day 1 |
| R8 | 2 Nov 2026 | High | Medium | Yes | Estimates and hours were entered only at the end of Sprint 1 (issue I4). Sprint 2 hours must be logged before the week 2 EVM |
| R9 | 2 Nov 2026 | Low (proposed) | Medium | Yes, in the spike | Workaround documented; no new reports in Sprint 2 |

### 2.2 New risks from Sprint 1

| ID | Risk | Probability | Impact | Strategy | Response | Trigger (early warning) | Owner | Status |
|---|---|---|---|---|---|---|---|---|
| R10 | The Team Lead owns the template, contracts, gateway and four services, so her workload blocks others when she is late | High | High | Mitigate | Blocking cards first at planning (retro action A1); another member may take a non-blocking card of hers with her agreement; the Scrum Master checks her load at every stand-up | A teammate is waiting for one of her pull requests for more than a day | Gayley Choden (Scrum Master) | Open |
| R11 | The proposal promises 25 services but the build is 20 (change C1), so assessors may mark the difference as missing scope | Medium | High | Mitigate | Explain change C1 and its reason in the closure report and the presentation; close old 25-service cards as "Won't do" (`schedule-evm.md` section 5) | A tutor asks for a removed service, or old cards still open at Sprint 3 planning | Sonam Zangmo (Team Lead) | Open |
| R12 | Jira statuses are moved in bulk at the end of a sprint, so the board does not show real progress and the tutors (who grade Jira) see a burst of changes | High | Medium | Mitigate | Move a card the day work starts and the day its pull request merges; one card In Progress per person (retro action A7) | Many cards change status on the last day of a sprint | Each card owner; Gayley Choden checks | Open |
| R13 | A rule needs data the proto contract does not carry, so services invent their own way to pass it (for example the caller's role for internal-notes travels in gRPC metadata `x-user-role`) | Medium | Medium | Mitigate | Record each case in the issue log and agree one convention with the contract owner before api-gateway is built | api-gateway and a service disagree on a header or field | Sonam Zangmo (contracts, api-gateway) | Reduced: role rule agreed on 2 Nov and gateway work planned for Sprint 3 (issue I8) |

### 2.3 Sprint 3 update: R3 (8 November 2026)

Updated by Sonam Choki on 8 November 2026, the day before the demo, on request of the Team Lead.

| ID | Review date | Probability now | Impact now | Happened? | Notes |
|---|---|---|---|---|---|
| R3 | 8 Nov 2026 | Low (proposed) | High | No | All 20 services and the data stores run on the 16 GB demo laptop, deployed and synced by Argo CD from `main` (SCRUM-102); self-heal tested (api-gateway scaled to 0, set back to 1 by Argo CD). A second machine runs the small demo set (`deploy/argocd/application-demo.yaml`). The escalation function runs as a Kubernetes CronJob instead of Knative (SCRUM-104): a Job every minute, all completed on the cluster. The demo script has a fallback for each step (`docs/pm/demo-script.md`). Still open: the full rehearsal and its recording (9 Nov), and a real escalation, which needs a 1 hour SLA window and is tested in the rehearsal |

## 3. Probability and impact matrix

| | Low impact | Medium impact | High impact |
|---|---|---|---|
| **High probability** | | R2, R8 | R6, R7 |
| **Medium probability** | | R9 | R1, R3, R4 |
| **Low probability** | | R5 | |

Risks in the top-right cells (R6, R7) are reviewed at every stand-up.

After the Sprint 2 review (proposed ratings, section 2):

| | Low impact | Medium impact | High impact |
|---|---|---|---|
| **High probability** | | R8, R12 | R4, R6, R10 |
| **Medium probability** | | R13 | R3, R7, R11 |
| **Low probability** | | R2, R5, R9 | R1 |

## 4. Review

The Scrum Master reviews this register at every sprint review: update status, add new risks, and close risks that can no longer happen. Changes are made through a pull request.
