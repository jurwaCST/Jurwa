# sla-escalation-checker (serverless function)

Owner: Nyendrak Yoezer Zangmo. HTTP on `PORT` (default 8080). Requirement FR-15.

## What it does

Escalates reports that passed their SLA window. Every request to `/` does one run and returns what it did:

```json
{"lapsed":1,"escalated":1,"skipped":0,"errors":0}
```

It keeps no data of its own. On Knative it scales to zero between runs, and a PingSource calls it on a schedule (SCRUM-104).

## One run

1. Ask sla-tracking `ListDueTimers` (server time, limit 100). sla-tracking marks each due timer LAPSED, publishes `sla.lapsed` and returns it, so a report is never handed out twice.
2. For each lapsed timer, call status-lifecycle `ChangeStatus` with `to_status = ESCALATED`, an empty `actor_id` (the system) and the reason `SLA lapsed: deadline was <time>`. status-lifecycle saves `previous_status` for the Saga.
3. If a batch was full (100), ask again, at most 10 batches per run. Anything left is picked up by the next run.

| Case | Result |
|---|---|
| Nothing due | `200`, all counts 0 |
| Report already RESOLVED or ESCALATED (`FAILED_PRECONDITION`) | counted in `skipped` |
| status-lifecycle fails for one report | counted in `errors`; the other reports are still escalated |
| sla-tracking cannot be reached | `503`, with the counts of anything already done in this run and `errors` + 1 |

Known gap: sla-tracking marks the timer LAPSED before the function escalates the report. If status-lifecycle is down for that report even after the retries in `pkg/resilience`, the next run will not retry it. The escalation Saga (SCRUM-89), where status-lifecycle reacts to the `sla.lapsed` event, closes this gap. The error is logged with the `report_id`.

The ESCALATED assignment back to the Category Head (reassignment) is also part of SCRUM-89.

## Layout

| Path | Content |
|---|---|
| `checker/` | The run, behind two interfaces (`DueTimers`, `StatusChanger`), tested with fakes (100%) |
| `clients/` | gRPC adapters for sla-tracking and status-lifecycle (100%) |
| `main.go` | HTTP handler (`GET` or `POST /`), connections with `pkg/resilience` |
| `Dockerfile` | Multi-stage image, built from the repo root |

Same layout as `functions/digest-dispatcher`.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | HTTP port (Knative sets it) |
| `SLA_TRACKING_ADDR` | `localhost:9012` | sla-tracking |
| `STATUS_LIFECYCLE_ADDR` | `localhost:9011` | status-lifecycle |

## Run (PowerShell)

Start the services, each in its own terminal (sla-tracking needs sla-policy):

```powershell
go run ./services/sla-policy/cmd/server
go run ./services/sla-tracking/cmd/server
$env:SEED_REPORT_IDS = "11111111-1111-1111-1111-111111111111"; go run ./services/status-lifecycle/cmd/server
```

Then the function, and call it:

```powershell
$env:PORT = "8090"; go run ./functions/sla-escalation-checker
curl.exe -i -X POST http://localhost:8090/
```

The shortest SLA window is 1 hour, so a report only lapses an hour after its timer starts. `docs/evidence/nyendrak/scrum-27-lapse-run.txt` is a real run that waited for that.

Docker (from the repo root):

```powershell
docker build -f functions/sla-escalation-checker/Dockerfile -t jurwa/sla-escalation-checker .
```

## Test

```powershell
go test ./functions/sla-escalation-checker/... -cover
```
