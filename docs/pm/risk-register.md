# Risk Register (SCRUM-44)

Project: Jurwa. Owner: Gayley Choden (Scrum Master). Version 1.0, 27 October 2026.

R1 to R5 come from the proposal (section 7) and the charter. R6 to R9 are new risks found during Sprint 1. Probability and impact are rated Low, Medium or High. Owners for R1 to R5 are proposed by the Scrum Master and confirmed at the Sprint 1 review.

## 1. Register

| ID | Risk | Probability | Impact | Strategy | Response | Trigger (early warning) | Owner | Status |
|---|---|---|---|---|---|---|---|---|
| R1 | Scope grows beyond the agreed architecture | Medium | High | Avoid | Scope is locked in the charter; any change needs full-team sign-off, recorded as a Jira `decision` card (see `comms-plan.md`) | A card or pull request adds a service or feature not in the charter | Sonam Zangmo (Team Lead) | Open |
| R2 | Learning curve of new tools (Kubernetes, gRPC) delays the start | High | Medium | Mitigate | Learning spike before full implementation (SCRUM-71, done 26 Oct); shared service template (SCRUM-73) | A member cannot build or run the template | Gayley Choden | Reduced: spike merged and demoed, template works on members' laptops |
| R3 | Local Kubernetes cluster or Knative setup is unstable near the demo date | Medium | High | Mitigate | Rehearse deployment early; keep a recorded demo as fallback; if Knative fails, run the same function code as Kubernetes CronJobs | Cluster or Knative fails to start in a rehearsal | Sonam Choki (deployment) | Open |
| R4 | Schedule slips once the build is under way | Medium | High | Mitigate | Weekly EVM shows drift early; lowest-priority work is cut first; testing and documentation are never cut | SPI or CPI below 0.9 (see `schedule-evm.md`) | Nyendrak (EVM baseline) | Open |
| R5 | Non-teaching staff without institutional email (cooks, cleaners) cannot log in | Low | Medium | Mitigate | Their supervisor or HoD holds the account and acts for them, with a full audit trail (requirement FR-24) | Pilot users report staff who cannot be reached | Gayley Choden (identity) | Open |
| R6 | The compressed schedule (8 weeks reduced to three 5-day sprints, change C2) leaves too little time to build, test and deploy 20 services | High | High | Mitigate | Services already cut from 25 to 20 (change C1); one thin slice per owner in Sprint 1; push daily and open draft pull requests early; carry over consciously at sprint review rather than miss silently | Sprint 1 cards carried over; SPI below 0.9 at the end of Sprint 1 | Gayley Choden (Scrum Master) | Open |
| R7 | Shared work owned by one person blocks everyone: the template (SCRUM-73) and contracts (SCRUM-74) gate every service slice | High | High | Mitigate | Template and contracts are first in the sprint order; reviewed by all three others within a day; the Scrum Master chases daily; blocked members do no-code cards meanwhile | SCRUM-74 still not merged on the last day of Sprint 1 (observed 27 Oct) | Sonam Zangmo (owner of SCRUM-73 and 74) | Open |
| R8 | Jira has no original estimates and few logged hours, so EV and AC cannot be calculated for the EVM report | High | Medium | Mitigate | Every owner adds an estimate to each card before the sprint and logs hours by sprint end; the Scrum Master checks on day 4 of each sprint | Cards without estimates at sprint planning (observed 26 Oct, `schedule-evm.md` section 6) | Gayley Choden (Scrum Master) | Open |
| R9 | k3d on Windows writes `host.docker.internal` as the cluster address, which resolves to a LAN IP on the campus network, so `kubectl` cannot connect | Medium | Medium | Mitigate | Create clusters with `--api-port 127.0.0.1:<port>` or point kubectl at `127.0.0.1` (documented in `docs/learning-spike/README.md`) | `kubectl` timeouts after `k3d cluster create` | Sonam Choki (k3d setup) | Open: found in the learning spike, 26 Oct |

## 2. Probability and impact matrix

| | Low impact | Medium impact | High impact |
|---|---|---|---|
| **High probability** | | R2, R8 | R6, R7 |
| **Medium probability** | | R9 | R1, R3, R4 |
| **Low probability** | | R5 | |

Risks in the top-right cells (R6, R7) are reviewed at every stand-up.

## 3. Review

The Scrum Master reviews this register at every sprint review: update status, add new risks, and close risks that can no longer happen. Changes are made through a pull request.
