# Final demo: outline, script and fallback

Card: SCRUM-84 (work package 1.8, epic SCRUM-70). Owner: Sonam Zangmo. Presentation: about 9 November 2026 (charter, section 7).

The demo shows the full life of one report, in the order a real report lives it, and proves each success criterion of the charter (section 10) on screen. It runs on Sonam Choki's 16 GB laptop with everything from `deploy/k8s`, deployed by Argo CD (`deploy/argocd/application.yaml`). Sonam Zangmo's 8 GB laptop with the small demo set (`deploy/argocd/application-demo.yaml`) is the second machine.

## 1. Outline

The time limit for the presentation is not given yet. This outline assumes 15 minutes as a working draft; section 1b shows what to keep when the slot is shorter.

| # | Time | Step | Presenter (proposed) | Shows | Charter criterion |
|---|---|---|---|---|---|
| 1 | 1:00 | Problem and architecture | Sonam Zangmo | One slide: 20 services, gateway, event bus, k3d | Objective 1 |
| 2 | 1:30 | GitOps | Sonam Zangmo | Argo CD UI: Synced and Healthy; scale the gateway to 0 and watch Argo CD restore it | "Deployed to k3d through GitHub Actions and Argo CD" |
| 3 | 2:00 | Student submits a report with a photo | Nyendrak Yoezer Zangmo | PWA on a phone: sign in, form, photo, submit; a timer runs on screen | "Submit with a photo in under two minutes" |
| 4 | 1:30 | Routing | Sonam Choki | Student's "My reports": the report is routed to DSA; timeline "Report submitted", "Routed to DSA" | "100% correct routing" |
| 5 | 2:00 | Category Head works the report | Sonam Zangmo | DSA signs in: "Assigned to me", adds an internal note, moves it to In Progress | Gateway roles, notes, status flow |
| 6 | 2:00 | SLA escalation | Nyendrak Yoezer Zangmo | A report submitted before the demo has passed its SLA: it is ESCALATED, the timeline shows it, the Saga hands it back | "SLA escalation works automatically" |
| 7 | 1:30 | Dashboard | Gayley Choden | System Administrator and President dashboards (counts per status and category) | Analytics |
| 8 | 1:30 | Quality | Gayley Choden | SonarQube coverage, Cypress run, JMeter numbers, accessibility | "70% coverage, tests passing in CI" |
| 9 | 1:00 | Observability | Sonam Choki | Grafana dashboard and a Jaeger trace of the submit in step 3 | NFR-08 |
| 10 | 1:00 | Project | Sonam Zangmo | EVM chart, risks closed, what we would do next | "Within budget, weekly EVM" |

Presenters are a proposal: each person shows what they built. Change it at the first rehearsal.

## 1b. If the slot is shorter

Steps are kept in this order, so the demo can be cut from the bottom without losing the story.

| Slot | Keep | Drop or move to one closing slide |
|---|---|---|
| 15 min or more | All 10 steps | Nothing |
| 10 min | 1, 3, 4, 5, 6, 2 | 7, 8, 9, 10 become one slide with the numbers (coverage, p95, EVM) |
| 5 min | 3, 4, 5, 6 (the report from submit to escalation), with one sentence on Argo CD | Everything else on the closing slide |

Steps 3 to 6 are the core: they prove routing and escalation, two of the charter's success criteria, live.

## 2. Script

Demo accounts come from identity's seed (`SEED_DEMO_USERS=true`, SCRUM-126). The password is shared in the group chat on the day, never written in the repo or on a slide.

| # | Who clicks | What to do | What the audience must see | Route or service |
|---|---|---|---|---|
| 2 | Zangmo | Open Argo CD (`kubectl -n argocd port-forward svc/argocd-server 8443:443`). Then `kubectl -n jurwa scale deploy/api-gateway --replicas=0` | App `jurwa` Synced and Healthy; the gateway is back to 1 within seconds | Argo CD selfHeal |
| 3 | Nyendrak | On the phone: sign in with the prepared student account (the PWA has no sign-up page; the account is made with `POST /auth/register` before the demo, like the pilot accounts), open "New report", pick Hostel, Hostel B, write "Tap leaking", attach a photo, submit | Confirmation with the report id, in under 2 minutes | `POST /photos`, `POST /reports` |
| 4 | Choki | Same phone: "My reports", open the report | Status ASSIGNED to DSA; timeline with "Report submitted" and "Routed to DSA" | `GET /me/reports`, `GET /reports/{id}/timeline`; routing-engine `report.routed` |
| 5 | Zangmo | Laptop: sign in as `dsa.demo@jurwa.example`; "Assigned to me"; open the report; add the note "Plumber booked for 2 pm"; set status In Progress | The report is in the list; the note is saved; status IN_PROGRESS; the student's timeline updates | `GET /me/assigned-reports`, `POST /reports/{id}/notes`, `POST /reports/{id}/status` |
| 6 | Nyendrak | Open the report prepared before the demo (section 3) | Status ESCALATED, timeline "SLA lapsed", reassigned by the Saga | sla-tracking, sla-escalation-checker, status-lifecycle, reassignment |
| 7 | Gayley | Sign in as `admin.demo`, then `president.demo` | Counts per status and per category, including the reports from this demo | `GET /dashboard/summary`, `GET /dashboard/categories` |
| 8 | Gayley | Slides from the evidence folders | SonarQube summary, Cypress video, JMeter table (p95 252 ms, 0 errors) | `docs/evidence/*/sprint3/` |
| 9 | Choki | Grafana; Jaeger, search for the submit from step 3 | Request rate and latency; one trace across gateway and report-submission | Prometheus, Grafana, Jaeger |

## 3. Before the demo

| When | Who | Check |
|---|---|---|
| Day before | Choki | Full cluster on the 16 GB laptop through Argo CD: all pods 1/1 Running, app Synced and Healthy |
| Day before | Zangmo | Second machine: the small demo set synced, as backup for steps 2 to 5 |
| Day before | All | One full rehearsal with a stopwatch, recorded (section 4) |
| T minus 90 min | Nyendrak | SLA window for HOSTEL set to 1 hour (the smallest allowed: `window_hours` is a whole number of hours) and the escalation report submitted, so it is overdue during step 6 |
| Before T minus 90 min | Choki (or Nyendrak with access) | Escalation pieces on the demo cluster, done once: rebuild and push status-lifecycle, sla-tracking and reassignment (`bash scripts/push-local-registry.sh status-lifecycle sla-tracking reassignment`) and restart them (`kubectl -n jurwa rollout restart deploy/<name>`): the images use the `latest` tag, so Argo CD does not notice new code by itself. Then build and push the checker image and `kubectl apply -f deploy/knative/cronjob-fallback.yaml`. For the demo the escalation runs as this CronJob only; do not apply `deploy/argocd/application-knative.yaml`, so the two never run together |
| T minus 90 min to the demo | All | **Do not restart sla-tracking or sla-policy** (no redeploy, no `rollout restart`, no Argo CD sync of it): its timers, and the 1-hour window in sla-policy, are in memory, so a restart loses the escalation report's timer or window. Check with `kubectl -n jurwa get pods` that its RESTARTS stays the same |
| T minus 30 min | Choki | `kubectl -n jurwa get pods` all Running; `curl /health` through Traefik |
| T minus 15 min | Zangmo | Create the demo student with `POST /auth/register` if it does not exist yet; sign in once with every demo account; phone and laptop on the same Wi-Fi; screen sharing tested |
| T minus 5 min | All | Close other apps; notifications off; browser zoom 125% |

## 4. Fallback (risk R3)

Each step has a fallback, so one failure never stops the demo. Switch without apologising and keep the timing.

| If this fails | Do this |
|---|---|
| The 16 GB cluster | Switch to the 8 GB laptop (steps 2 to 5 run there); show steps 6 to 9 from the recording |
| Argo CD UI | Show `kubectl -n argocd get applications` and `docs/evidence/sonam-zangmo/sprint3/argocd-ui-synced-healthy.png` |
| The phone or Wi-Fi | Run the PWA in the laptop browser in a phone-sized window |
| A route returns an error | Play that step from the recording; the evidence file of the route has the same call with its answer |
| Escalation not visible in time | Run the checker once by hand (`curl` its `/`), or play step 6 from the recording |
| Everything | Play the full recording (about 12 minutes) and present over it |

The recording is made at the full rehearsal (screen and voice, 1080p), saved in the team Drive, and linked from `docs/evidence/` (not committed: too large).

## 5. Open items before the rehearsal

These block parts of the script. Each owner says when it is merged.

| Item | Blocks step | Owner | Card |
|---|---|---|---|
| Nothing calls sla-tracking `StartTimer`, so no report gets an SLA timer in a real run. Nyendrak is adding the consumers: report.submitted registers OPEN in status-lifecycle; report.routed sets ASSIGNED, records the first assignee in reassignment and starts the timer; status.changed to RESOLVED stops it | 6 | Nyendrak | in progress |
| The escalation checker is not deployed. Nyendrak adds a run-once mode and a CronJob every minute in `deploy/knative/` (fallback while Knative is not done); Choki tests it on the cluster | 6 | Nyendrak, Choki | SCRUM-104 |
| sla-tracking keeps timers in memory until its database adapter lands (see the T minus 90 rule) | 6 | Nyendrak | |
| `report.list_assigned` in the authorization policy | 5 | Gayley | SCRUM-137 |
| Student "My reports" page with the timeline, and the staff page ("Assigned to me", note, status); the pilot (SCRUM-113) found both missing. Being built now; if not ready, show steps 4 and 5 from the API in a terminal | 4, 5 | Nyendrak, Gayley | SCRUM-36 |
| Web app deployed in the cluster on the same HTTPS address as `/api` | 3 | Choki (manifests), Nyendrak or Gayley (Dockerfile) | k8s: deploy the web app |
| Gateway has no feedback, notifications, reassignment or admin (SLA policy) routes; SLA policy is set with grpcurl before the demo | 6, and the feedback step that is left out | Zangmo | after the demo, if wanted |

## 6. Rehearsal log

| Date | Who was there | Total time | What went wrong | Fixed by |
|---|---|---|---|---|
| | | | | |
