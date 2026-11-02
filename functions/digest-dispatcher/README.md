# digest-dispatcher (serverless function)

Owner: Gayley Choden. HTTP on `PORT` (default 8080). Requirement FR-31.

## What it does

Sends each Category Head a short in-app digest of the reports in their category that are not resolved yet. Every request to `/` does one run and returns how many digests went out:

```json
{"digests":3,"errors":0}
```

It keeps no data of its own. On Knative it scales to zero between runs, and a PingSource calls it on a schedule (Sprint 3).

## One run

1. Read the counts per category from analytics `GetCategoryBreakdown`. If analytics cannot be reached, stop: nothing is sent, the answer is `503` with `{"digests":0,"errors":1}`.
2. For each category with unresolved reports (total minus resolved), find the Category Head:

   | Category | Position |
   |---|---|
   | Hostel | DSA |
   | Academic Block | DAA |
   | Others | SSO |

   The user holding the position comes from department-directory `GetStaff`.
3. Send one notification (`IN_APP`, `event_type` `digest.daily`) with notification `SendNotification`, for example:

   > Daily digest: Hostel
   > Hostel: 4 unresolved (2 open, 1 assigned, 1 in progress), 1 escalated. Please check Jurwa.

| Case | Result |
|---|---|
| No unresolved reports anywhere | nothing sent, `{"digests":0,"errors":0}` |
| A head with nothing unresolved | that head gets no digest |
| Directory lookup or send fails for one head, or nobody holds the position | counted in `errors`; the other heads still get theirs |
| Message length | at most 200 characters (notification's limit), proven for the largest possible numbers |

The College President digest in FR-31 is not built yet: department-directory has no President position. Sprint 3.

## Layout

| Path | Content |
|---|---|
| `dispatcher/` | The digest rules, behind three interfaces (`CountsSource`, `HeadDirectory`, `Sender`), tested with fakes (100%) |
| `clients/` | gRPC adapters for analytics, department-directory and notification (100%) |
| `main.go` | HTTP handler (`GET` or `POST /`), connections with `pkg/resilience` |
| `Dockerfile` | Multi-stage image, built from the repo root |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | HTTP port (Knative sets it) |
| `ANALYTICS_ADDR` | `localhost:9020` | analytics |
| `DEPARTMENT_DIRECTORY_ADDR` | `localhost:9008` | department-directory |
| `NOTIFICATION_ADDR` | `localhost:9018` | notification |

## Run (PowerShell)

Start the three services, each in its own terminal:

```powershell
$env:SEED_DEMO_DATA = "true"; go run ./services/analytics/cmd/server
go run ./services/department-directory/cmd/server
go run ./services/notification/cmd/server
```

Then the function, and call it:

```powershell
$env:PORT = "8082"; go run ./functions/digest-dispatcher
curl.exe -i http://localhost:8082/
grpcurl -plaintext -d '{\"recipient_user_id\":\"00000000-0000-4000-8000-000000000001\"}' localhost:9018 jurwa.notification.v1.NotificationService/ListNotifications
```

Docker (from the repo root):

```powershell
docker build -f functions/digest-dispatcher/Dockerfile -t jurwa/digest-dispatcher .
```

## Knative (Sprint 3)

Deployed as a Knative Service with a PingSource that posts to it once a day. Nyendrak's Knative card (SCRUM-104) has the YAML pattern.

## Test

```powershell
go test ./functions/digest-dispatcher/... -cover
```
