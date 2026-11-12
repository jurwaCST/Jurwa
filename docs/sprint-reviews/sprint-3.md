# Sprint 3 review and retrospective

Sprint 3: 4 to 10 November 2026. Scrum Master: Gayley Choden.
Sources: Jira project SCRUM (read on 10 November 2026, after all hours were logged), the Sprint 3 EVM log `docs/pm/evm/week3.md` (Nyendrak, #123), merged pull requests on `main`, and each member's written retro.

## 1. Sprint goal

Deploy to k3d with Argo CD, real databases and Redis Streams events, the React PWA, the full test suite, and the final demo.

**Result: mostly met.**

- **Met:** the cluster runs all 20 services, the data stores and the web app, deployed from `main` by Argo CD. Events go through Redis Streams, and the database adapters are merged. The escalation function runs on Knative with a PingSource. The test pyramid is complete (Go, web, Cypress, accessibility, property, JMeter), and the student pilot met the two-minute goal.
- **Not met:** the app is not installable yet (SCRUM-98), and the rehearsal and final demo moved to Sprint 4.

## 2. What got done

44 cards Done, 182.5 planned hours. 19 of them (70.5 h) were created during the sprint, for pieces the pages and the demo needed that were not on a card at planning (`week3.md` section 3).

| Owner | Card | PR |
|---|---|---|
| Gayley Choden | SCRUM-118 web: React app setup | #67 |
| | SCRUM-126 identity: demo accounts | #70 |
| | SCRUM-127 database adapters | #74 |
| | SCRUM-106 integration tests (Testcontainers) | #75 |
| | SCRUM-105 observability | #76 |
| | SCRUM-108 end-to-end tests (Cypress) | #95 |
| | SCRUM-112 test plan | #98 |
| | SCRUM-137 authorization: report.list_assigned | #101 |
| | SCRUM-138 web: visual design and sign-up page | #111 |
| | SCRUM-141 web: shared components, 403 and 404 pages | #115 |
| | SCRUM-110 accessibility tests | #118 |
| | SCRUM-99 President dashboard (stretch) | #117 |
| Nyendrak Yoezer Zangmo | SCRUM-35 student report form | #81, #119 |
| | SCRUM-89 escalation Saga | #85, #93, #94, #102 |
| | SCRUM-107 property-based tests (gopter) | #89 |
| | SCRUM-104 Knative and PingSource | #103, #116 |
| | SCRUM-36 staff view | #121 |
| | SCRUM-37 admin dashboard (counts) | #122 |
| Sonam Choki | SCRUM-116 internal-notes permissions | #52 |
| | SCRUM-122 pkg/eventbus (Redis Streams) | #68 |
| | SCRUM-123 Kubernetes manifests | #71 |
| | SCRUM-125 database adapters | #77 |
| | SCRUM-103 Traefik and cert-manager TLS | #78 |
| | SCRUM-39 health checks | #79 |
| | SCRUM-131 report-history: ListReportsByAssignee | #83 |
| | SCRUM-111 SonarQube | #88 |
| | SCRUM-135 k8s gateway addresses | #90 |
| | SCRUM-130 report-history event consumers | #92 |
| | SCRUM-113 pilot usability test | #97, #110 |
| | SCRUM-139 k8s: deploy the web app | #109, #114 |
| Sonam Zangmo | SCRUM-121 resilience: gomock tests | #65 |
| | SCRUM-117 gateway: notes routes | #66 |
| | SCRUM-124 database adapters | #69 |
| | SCRUM-119 gateway: history routes | #73 |
| | SCRUM-120 gateway: photo routes | #80 |
| | SCRUM-109 performance tests (JMeter) | #82 |
| | SCRUM-128 gateway: status routes | #84 |
| | SCRUM-133 routing-engine: report.routed | #86 |
| | SCRUM-129 gateway: dashboard routes | #87 |
| | SCRUM-136 gateway: assigned reports route | #91, #108 |
| | SCRUM-102 Argo CD and Argo Rollouts | #96, #104, #106 |
| | SCRUM-132 resilience: no retry after a timeout | #99 |
| | SCRUM-84 demo script and presentation | #100, #107, #113 |
| | SCRUM-142 gateway: reassign and positions routes | #120 |

Also merged during Sprint 3, but not among the 44 Sprint 3 cards in `week3.md`: the web Dockerfile SCRUM-140 (#112, Gayley), the risk register R3 update SCRUM-83 (#105, Sonam Choki) and the Sprint 3 EVM log SCRUM-82 (#123, Nyendrak).

## 3. Carried over to Sprint 4

| Item | Owner | Why it is open |
|---|---|---|
| SCRUM-98 PWA installability (7 h) | Nyendrak | Waited for the shared design; no hours logged |
| SCRUM-85 closure report (1.5 h left of 4 h) | Gayley | End-of-project work; it needs the final demo, Sprint 4 EVM and sign-off |
| New card: edit routing rules and SLA hours on the admin page (4 h) | Nyendrak | The gateway has no admin routes yet |
| Rehearsal and final demo | Zangmo, whole team | Moved to Sprint 4 |
| A real SLA escalation tested end to end | Sonam Choki, Nyendrak | Needs a one-hour window; planned for the rehearsal |
| New card: real locations (wing names, room and hall numbers, the 3 canteens, Wi-Fi and Estate) | Sonam Choki | Pilot problems 1, 2 and 5 |
| New card: check the location exists when a report is submitted | Sonam Choki | The gateway accepted building "x" |
| New card: first submit after a restart times out ("service busy") | Sonam Choki, with Zangmo for the gateway timeout | Seen in the pilot setup |
| Pilot tasks 3 and 4 with staff | Sonam Choki (SCRUM-113 follow-up) | The staff view landed on 10 November; "My reports" does not exist yet |
| "My reports" and notifications pages | No owner yet | No card at planning |
| Knative fallback doc: delete the PingSource too | Nyendrak | Found by Sonam Choki while testing |
| Accessibility re-run on the staff and admin pages; exploratory session; screen reader check | Gayley | The pages were merged on the last day |
| D-08: report-submission saves before it publishes the event | Not assigned | Open defect, `docs/testing/test-plan.md` section 6 |
## 4. Hours and EVM (from `week3.md`)

| Member | Cards Done | Estimate of Done cards | Hours logged |
|---|---|---|---|
| Gayley Choden | 12 | 56 h | 59.5 h |
| Nyendrak Yoezer Zangmo | 6 | 30 h | 36 h |
| Sonam Choki | 12 | 47.5 h | 50.5 h |
| Sonam Zangmo | 14 | 49 h | 53 h |
| **Team** | **44** | **182.5 h** | **199 h** |

| | Sprint 3 | Cumulative (Sprints 1 to 3) |
|---|---|---|
| EV | 182.5 h | 455.5 h |
| AC | 199 h | 424 h |
| CPI | 0.92 | 1.07 |
| SPI against the team's plan | 1.00 | |
| SPI against the baseline (93.3 h per sprint) | 1.96 | 1.63 |

- **CPI is below 1 for the first time:** the Done work took about 9% more hours than planned. Most of the overrun is integration work done twice or that others waited on (the Saga, Kubernetes manifests, Argo CD, the database adapters, the visual design), and the three UI pages, which were rebuilt once the shared design arrived.
- **SPI 1.00 against the team's plan:** every card kept in Sprint 3 is Done.
- **The baseline is still too low:** at 93.3 h per sprint, it is about half of what the team plans and does, so SPI against it means little (action A1 below).
- AC includes the 2.5 h logged on SCRUM-85, which carries over, and SCRUM-98, which has no hours logged.

## 5. Sprint review

No review meeting was held; each member sent a written retro. Highlights:

- **The whole system runs in the cluster.** All 20 services, the data stores and the web app run behind one HTTPS address (the app on `/`, the API on `/api`), deployed by Argo CD from `main`. Argo CD heals a change made by hand: api-gateway was scaled to 0, and Argo CD set it back to 1 (Sonam Choki, Zangmo).
- **Escalation runs on Knative.** A PingSource starts the escalation function every 2 minutes, and it scales 0 to 1 to 0. The escalation Saga runs for real on Redis, and the CronJob fallback was tested in both directions (Nyendrak, Sonam Choki).
- **The pilot met the two-minute goal.** 5 of 5 students reported a problem with a photo in under two minutes: average 62 s, slowest 90 s, no help needed (Sonam Choki, SCRUM-113).
- **Tests found real bugs before users did.** JMeter found that a retried Register returned 409, fixed in SCRUM-132 (Zangmo). Re-running the accessibility tests on the new design found D-16, fixed before the merge (Gayley).
- **The web app has a full design and every role has a page:** the campus design and sign-up page, shared components, the report form, the staff view, the admin dashboard and the President dashboard. There are 112 web tests, plus 8 end-to-end Cypress tests and 12 accessibility tests (Gayley, Nyendrak).
- **The gateway covers every screen the web app needs,** including reassign, added the same day it was asked for (Zangmo, SCRUM-142).

## 6. Retrospective

### What went well

- **Sonam Choki:**
  - The cluster runs every service, deployed and kept in sync by Argo CD.
  - The web app is in the cluster, written tests first: 26 checks, also run in CI.
  - The pilot met charter success criterion 2.
  - Knative is installed, and the CronJob fallback was tested in both directions.
  - Reviewed and tested teammates' work on the cluster: SCRUM-104, SCRUM-89, #120, and the web image with SCRUM-36 and SCRUM-37.
- **Sonam Zangmo:**
  - Every card merged with a tests-first commit, a live check against the real services and evidence.
  - Live checks found gaps before the demo did: no timer start for escalation, missing gateway addresses, no student or staff pages, the web app not in the cluster. Each one got an owner and a card.
  - Argo CD shows Synced and Healthy, and self-heal was shown.
- **Nyendrak Yoezer Zangmo:**
  - All six cards kept in Sprint 3 merged with tests first and evidence: the Saga, Knative, the gopter tests and the three UI pages on the shared design.
  - Agreeing details in the group before building meant the pieces fitted first time: the CronJob fallback, the Argo CD Knative app and the positions route. Reassign was done the same night Zangmo added the routes.
- **Gayley Choden:**
  - Built the web base: the campus design and sign-up page, shared components, the 403 and 404 pages, the President dashboard and the web Dockerfile, all tests first.
  - The test plan tracks 16 defects. Cypress passes 8 of 8 end to end. Accessibility: 0 axe-core violations, Lighthouse 100, and the keyboard check passes.
  - When the free GitHub Actions minutes ran out and blocked every merge, the Git history was checked for secrets and the repository was made public, so checks ran again the same day at zero cost (issue I10).

### What did not go well

- **Missing pieces were found late.**
  - The UI pages and the gateway routes they needed (reassign, admin) were not on any card at planning: 19 cards were added during the sprint.
  - The missing student and staff pages were only found in the pilot, so the demo is not ready this sprint (Zangmo, Nyendrak).
  - Some features were dropped late: President escalation routing (Zangmo).
- **Pages were built before the shared design.** The report form was rebuilt twice, and each UI page took about an hour more than planned (Nyendrak). The shared components came in the middle of the sprint (Gayley).
- **Hours and cards slipped.**
  - All of Nyendrak's Sprint 3 hours were logged on one day, and three UI cards were In Progress at once (Nyendrak).
  - SCRUM-113 was moved to Done before its report was merged and had to be moved back (Sonam Choki).
  - SCRUM-140's hours were logged late (Gayley).
- **Hardware and network.**
  - Zangmo's laptop was too small for the full cluster: Docker hung at 4 GB and the load test overloaded it, so the demo now runs on Sonam Choki's laptop.
  - GitHub downloads timed out on the campus network (504) during the Knative install (Sonam Choki).
- **Setup surprises cost time.**
  - The organisation blocks deploy keys (switched to a read-only token). `:latest` images mean Argo CD does not redeploy new code by itself, and `kubectl rollout restart` fights Argo CD self-heal, so pods were deleted instead (Zangmo, Sonam Choki).
  - Cypress hung on TypeScript 7 and Node 26, and the dev server gave random blank pages under load (Gayley).
- **A fallback doc was incomplete.** Deleting the Argo CD app leaves the PingSource running, so the PingSource and the CronJob both escalated. Sonam Choki found it while testing (Nyendrak).
- **The demo password appeared in some screenshots.** It was kept out of the evidence (Sonam Choki).
- **The free Actions limit was not watched** until it blocked every merge two days before the demo (Gayley).

### Sprint 2 actions

| # | Action | Result |
|---|---|---|
| A1 | An estimate on every card before In Progress | Met: all 46 cards had planned hours |
| A2 | Log hours every day | Partly met: Nyendrak's hours were all logged on 10 November |
| A3 | Every shared piece gets an owner and a card | Met: the event bus (SCRUM-122, Sonam Choki), the web app and its components (SCRUM-118, 138, 141, Gayley) |
| A4 | Open a PR when a card is done; no card waits for the last day | Partly met: many cards merged during the sprint, but the staff view, admin dashboard and EVM merged on 10 November |
| A5 | Merge only after the green Approved | No merge before approval was reported |
| A6 | Check `api-routes.md` and the policy before a role rule | No new role rule mismatch was reported |
| A7 | Settle open design questions at the start | Not met: the gateway routes for the pages were settled mid-sprint |
| A8 | One card In Progress per person | Not met for Nyendrak (three UI cards at once) |

### What to change in Sprint 4

| # | Action | Owner |
|---|---|---|
| A1 | Rebaseline the EVM from Jira: PV per sprint is the team's planned hours | Nyendrak |
| A2 | Run the full demo flow end to end at the start of the sprint, so missing pieces show up early | Zangmo, whole team |
| A3 | A page card starts only when its components and gateway routes are merged; list the routes on the card at planning | Everyone; Gayley checks at planning |
| A4 | Move a card to Done only after its PR is merged and its evidence is on `main` | Everyone |
| A5 | Log hours the same day the work is done | Everyone |
| A6 | Give images a version tag, not `:latest`, so a merge redeploys through Argo CD | Zangmo |
| A7 | Test every fallback step on the cluster before it goes into a doc or the demo script | Whoever writes the step |
| A8 | Download installers early and keep a local copy | Sonam Choki |
| A9 | Clear the terminal after any command that prints a secret, before a screenshot | Everyone |
| A10 | Check the free limits of every tool (Actions minutes and others) at the start of each sprint | Gayley |
| A11 | Agree the demo dates and who presents what before building more | Zangmo |
| A12 | Re-run the browser and accessibility tests after every page merge, on the built app | Gayley |

## 7. Sprint 4

Sprint 4: 11 to 17 November 2026, the last sprint. Move the cards in section 3 into it in Jira. Goal (to confirm at planning): the installable PWA, the carried-over cards, the rehearsal with a real SLA escalation, the final demo, and the closure report (SCRUM-85).
