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
| 9 | 1:00 | Observability | Sonam Choki | Grafana dashboard and the JSON log line of the submit in step 3 (no Jaeger: tracing was not built, `docs/ops/observability.md`) | NFR-08 |
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
| 3 | Nyendrak | On the phone: sign in with the prepared student account (made before the demo on the sign-up page or with `POST /auth/register`), open "Report", pick the Hostel tile, a hostel, its floor and area, write "Tap leaking", attach a photo, send | Confirmation with the report id, in under 2 minutes | `POST /photos`, `POST /reports` |
| 4 | Choki | Same phone: "My reports", open the report | Badge Assigned; timeline "Open" (Report submitted), then "Assigned"; the bell has a new message "We received your report" | `GET /me/reports`, `GET /reports/{id}/timeline`; routing-engine `report.routed` |
| 5 | Zangmo | Laptop: sign in as `dsa.demo@jurwa.example`; "Assigned to me"; open the report; add the note "Plumber booked for 2 pm"; set status In Progress | The report is in the list; the note is saved; status IN_PROGRESS; the student's timeline updates | `GET /me/assigned-reports`, `POST /reports/{id}/notes`, `POST /reports/{id}/status` |
| 6 | Nyendrak | Open the report prepared before the demo (section 3) | Status ESCALATED, timeline "SLA lapsed", reassigned by the Saga | sla-tracking, sla-escalation-checker, status-lifecycle, reassignment |
| 7 | Gayley | Sign in as `admin.demo`, then `president.demo` | Counts per status and per category, including the reports from this demo | `GET /dashboard/summary`, `GET /dashboard/categories` |
| 8 | Gayley | Slides from the evidence folders | SonarQube summary, Cypress video, JMeter table (p95 252 ms, 0 errors) | `docs/evidence/*/sprint3/` |
| 9 | Choki | Grafana (`kubectl -n jurwa port-forward svc/grafana 3000:3000`, http://localhost:3000, dashboard "Jurwa services"); then `kubectl -n jurwa logs deploy/report-submission --tail=5` | Request rate, errors and latency per service; the log line of the step 3 submit with its correlation id | Prometheus, Grafana, JSON logs |

## 3. Before the demo

| When | Who | Check |
|---|---|---|
| Day before | Choki | Full cluster on the 16 GB laptop through Argo CD: all pods 1/1 Running, app Synced and Healthy |
| Day before | Zangmo | Second machine: the small demo set synced, as backup for steps 2 to 5 |
| Day before | All | One full rehearsal with a stopwatch, recorded (section 4) |
| T minus 90 min | Nyendrak | SLA window for HOSTEL set to 1 hour (the smallest allowed: `window_hours` is a whole number of hours) and the escalation report submitted, so it is overdue during step 6 |
| Before T minus 90 min | Choki (or Nyendrak with access) | Latest code on the demo cluster, done once after the last merge: `git pull` on `main`, then rebuild and push every image, the web app included (`bash scripts/push-local-registry.sh` and `bash scripts/push-local-registry.sh web`), and delete the pods so they pull it (`kubectl -n jurwa delete pod --all`, then wait until `kubectl -n jurwa get pods` shows all Running). Never `rollout restart`, which Argo CD self-heal undoes: the images use the `latest` tag with `imagePullPolicy: Always`, so Argo CD does not notice new code by itself but a new pod pulls it. Then build and push the checker image and `kubectl apply -f deploy/knative/cronjob-fallback.yaml`. For the demo the escalation runs as this CronJob only; do not apply `deploy/argocd/application-knative.yaml`, so the two never run together |
| T minus 90 min to the demo | All | **Do not restart sla-tracking or sla-policy** (no redeploy, no `rollout restart`, no Argo CD sync of it): its timers, and the 1-hour window in sla-policy, are in memory, so a restart loses the escalation report's timer or window. Check with `kubectl -n jurwa get pods` that its RESTARTS stays the same |
| T minus 30 min | Choki | `kubectl -n jurwa get pods` all Running; `curl /health` through Traefik |
| T minus 15 min | Zangmo | Create the demo student with `POST /auth/register` if it does not exist yet; sign in once with every demo account; phone and laptop on the same Wi-Fi; screen sharing tested |
| T minus 5 min | All | Close other apps; notifications off; browser zoom 125% |

## 3b. Commands, in order

On the demo laptop, in the repo folder. Each block waits for the one before it.

**A. Latest code on the cluster (about 30 to 45 minutes)**

```bash
git checkout main && git pull
k3d cluster start jurwa                      # only if the cluster is stopped
kubectl config current-context               # must say k3d-jurwa
bash scripts/push-local-registry.sh          # all 20 services
bash scripts/push-local-registry.sh web      # the web app
docker build -f functions/sla-escalation-checker/Dockerfile -t localhost:5050/sla-escalation-checker:latest .
docker push localhost:5050/sla-escalation-checker:latest
kubectl -n jurwa delete pod --all            # new pods pull the new images
kubectl -n jurwa get pods -w                 # wait until all are Running, then Ctrl+C
```

**B. Escalation runner, Grafana, Argo CD**

```bash
kubectl apply -f deploy/knative/cronjob-fallback.yaml
kubectl get cronjob,pingsource -A            # only the CronJob may run the checker; delete a PingSource for it if one is listed
kubectl apply -k deploy/observability        # Prometheus and Grafana (not part of the Argo CD app)
kubectl -n argocd get applications           # jurwa: Synced, Healthy
```

**C. T minus 90: SLA window first, then the escalation report.** The timer reads the window when the report is routed, so the order matters.

```bash
kubectl -n jurwa port-forward svc/sla-policy 9013:9013      # keep open; then in a second terminal:
grpcurl -plaintext -d '{"category":"HOSTEL","window_hours":1}' localhost:9013 jurwa.slapolicy.v1.SlaPolicyService/UpsertPolicy
grpcurl -plaintext -d '{"category":"HOSTEL"}' localhost:9013 jurwa.slapolicy.v1.SlaPolicyService/GetPolicy
```

Then, in the app as the demo student, submit a HOSTEL report and a spare and note the time: step 6 is ready about 62 minutes later. From now on, no restarts.

**D. Before the demo**

```bash
# demo student, once; the password comes from the group chat, never from this file
curl -k -X POST https://localhost/api/auth/register -H "Content-Type: application/json" \
  -d '{"email":"<demo student email>","fullName":"<name>","password":"<password>"}'
kubectl -n jurwa get pods                                     # all Running, RESTARTS unchanged
kubectl -n argocd port-forward svc/argocd-server 8443:443     # Argo CD UI, user admin
kubectl -n jurwa port-forward svc/grafana 3000:3000           # Grafana
```

The Argo CD admin password is in the Secret `argocd-initial-admin-secret` of that cluster (`deploy/argocd/README.md`, section 5); type it in, never paste it in a chat.

**E. During the demo:** step 2 `kubectl -n jurwa scale deploy/api-gateway --replicas=0`; step 9 `kubectl -n jurwa logs deploy/report-submission --tail=5`; step 6 fallback in section 4.

## 4. Fallback (risk R3)

Each step has a fallback, so one failure never stops the demo. Switch without apologising and keep the timing.

| If this fails | Do this |
|---|---|
| The 16 GB cluster | Switch to the 8 GB laptop (steps 2 to 5 run there); show steps 6 to 9 from the recording |
| Argo CD UI | Show `kubectl -n argocd get applications` and `docs/evidence/sonam-zangmo/sprint3/argocd-ui-synced-healthy.png` |
| The phone or Wi-Fi | Run the PWA in the laptop browser in a phone-sized window |
| A route returns an error | Play that step from the recording; the evidence file of the route has the same call with its answer |
| Escalation not visible in time | Run the checker once by hand: `kubectl -n jurwa create job --from=cronjob/sla-escalation-checker check-now`, wait 20 seconds and reload; or play step 6 from the recording |
| Everything | Play the full recording (about 12 minutes) and present over it |

The recording is made at the full rehearsal (screen and voice, 1080p), saved in the team Drive, and linked from `docs/evidence/` (not committed: too large).

## 5. Open items before the rehearsal

These block parts of the script. Each owner says when it is merged.

| Item | Blocks step | Owner | Card |
|---|---|---|---|
| sla-tracking keeps timers in memory until its database adapter lands (see the T minus 90 rule) | 6 | Nyendrak | future work |
| Gateway has no admin (SLA policy) routes; the SLA policy is set with grpcurl before the demo | 6 | Zangmo | future work |

Done since Sprint 3 (checked on main, 18 November 2026, SCRUM-153):

| Item | Done by |
|---|---|
| Consumers that start the SLA timer (report.submitted, report.routed, status.changed) | sla-tracking, status-lifecycle and reassignment consumers |
| Escalation checker deployed (CronJob fallback, and Knative PingSource on Choki's cluster) | SCRUM-104 |
| `report.list_assigned` in the authorization policy | SCRUM-137 |
| Student "My reports" with the timeline, and the staff page ("Assigned to me", note, status, reassign) | SCRUM-143, SCRUM-36 |
| Web app deployed in the cluster on the same HTTPS address as `/api` | SCRUM-139 |
| Gateway routes for feedback, notifications and reassignment | SCRUM-147, SCRUM-142 |
| Feedback stars on a resolved report (can be shown after step 5: set the report Resolved, the student rates it) | SCRUM-145 |
| Installable app (PWA) with an offline banner, for the phone in step 3 | SCRUM-98 |
| Messages to the student: "We received your report", then In progress and Resolved; the bell shows a dot | SCRUM-155 (PR #134, merged 19 Nov) |

## 6. Rehearsal log

| Date | Who was there | Total time | What went wrong | Fixed by |
|---|---|---|---|---|
| | | | | |
