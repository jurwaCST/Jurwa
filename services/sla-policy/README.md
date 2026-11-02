# sla-policy

Owner: Nyendrak Yoezer Zangmo. gRPC port 9013. Contract: `proto/sla-policy/v1/sla-policy.proto`.

## What it does

Stores the SLA window (in hours) for each category: how long a report may stay unresolved before it is escalated. sla-tracking reads it when a timer starts. It never tracks timers or escalates reports itself.

## Rules

| Rule | Result |
|---|---|
| Category must be HOSTEL, ACADEMIC_BLOCK or OTHERS | otherwise `INVALID_ARGUMENT` |
| `window_hours` must be greater than 0 | 0 or negative gives `INVALID_ARGUMENT` |
| `updated_by` must be a UUID | otherwise `INVALID_ARGUMENT` |
| One window per category; Upsert replaces it | |
| No window set yet | `GetPolicy` gives `NOT_FOUND` |

The window values are not decided yet (see `docs/design/database-schema.md`, open questions), so nothing is seeded. The System Administrator sets them with `UpsertPolicy`. Role checks happen in the api-gateway and authorization service.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `GetPolicy` | Returns the window for a category | `INVALID_ARGUMENT`, `NOT_FOUND` |
| `UpsertPolicy` | Creates or replaces the window for a category | `INVALID_ARGUMENT` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Category`, `Policy`, `NewPolicy` rules (100% tested) |
| `application/` | Use cases: `GetPolicy`, `UpsertPolicy` |
| `ports/` | `PolicyRepository` interface |
| `adapters/in/grpc/` | gRPC handler, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory store (PostgreSQL later, schema in `docs/design/schema/sla-policy.sql`) |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9013` | gRPC port |
| `SERVICE_NAME` | `sla-policy` | Name written in every log line |

## Run (PowerShell)

```powershell
go run ./services/sla-policy/cmd/server
```

In a second terminal:

```powershell
grpcurl -plaintext localhost:9013 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{\"category\":\"HOSTEL\",\"window_hours\":48,\"updated_by\":\"22222222-2222-2222-2222-222222222222\"}' localhost:9013 jurwa.slapolicy.v1.SlaPolicyService/UpsertPolicy
grpcurl -plaintext -d '{\"category\":\"HOSTEL\"}' localhost:9013 jurwa.slapolicy.v1.SlaPolicyService/GetPolicy
```

Docker (from the repo root):

```powershell
docker build -f services/sla-policy/Dockerfile --build-arg SERVICE=sla-policy -t jurwa/sla-policy .
docker run --rm -p 9013:9013 jurwa/sla-policy
```

## Test

```powershell
go test ./services/sla-policy/... -cover
```
