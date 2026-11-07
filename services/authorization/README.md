# authorization

Owner: Gayley Choden. gRPC port 9017. Contract: `proto/authorization/v1/authorization.proto`.

## What it does

Answers one question: may this role do this action? The api-gateway (and other services) call it after identity has checked the token. It never checks tokens, and it never checks who owns a report: "Student (own)" is checked by the caller, which has the report data.

## Rules

| Rule | Result |
|---|---|
| Default deny: an action not in the table is denied | `allowed: false` with a reason |
| Unknown role number | `allowed: false`, reason `unknown role` |
| `role` is `ROLE_UNSPECIFIED` | `INVALID_ARGUMENT` |
| `action` or `resource` is empty | `INVALID_ARGUMENT` |
| A denial is a normal answer, not a gRPC error | `allowed: false`, `reason` says why |

## Role table

From `docs/design/api-routes.md`. S = Student, CH = Category Head, SF = Staff, AD = System Administrator, PR = College President.

| Action | Route | S | CH | SF | AD | PR |
|---|---|---|---|---|---|---|
| `me.read` | `GET /me` | yes | yes | yes | yes | yes |
| `category.list` | `GET /categories` | yes | yes | yes | yes | yes |
| `location.list` | `GET /categories/{category}/locations` | yes | yes | yes | yes | yes |
| `notification.list` | `GET /me/notifications` | yes | yes | yes | yes | yes |
| `photo.upload` | `POST /photos` | yes | | | | |
| `photo.read` | `GET /photos/{photoId}` | yes | yes | yes | yes | |
| `report.submit` | `POST /reports` | yes | | | | |
| `report.read` | `GET /reports/{reportId}` | yes | yes | yes | yes | |
| `report.list_own` | `GET /me/reports` | yes | | | | |
| `report.timeline.read` | `GET /reports/{reportId}/timeline` | yes | yes | yes | yes | |
| `report.status.read` | `GET /reports/{reportId}/status` | yes | yes | yes | yes | |
| `report.status.change` | `POST /reports/{reportId}/status` | | yes | yes | | |
| `report.reassign` | `POST /reports/{reportId}/reassign` | | yes | yes | | |
| `report.assignments.read` | `GET /reports/{reportId}/assignments` | | yes | yes | yes | |
| `report.notes.add` | `POST /reports/{reportId}/notes` | | yes | yes | | |
| `report.notes.read` | `GET /reports/{reportId}/notes` | | yes | yes | yes | |
| `report.feedback.submit` | `POST /reports/{reportId}/feedback` | yes | | | | |
| `report.feedback.read` | `GET /reports/{reportId}/feedback` | yes | yes | | yes | |
| `report.audit.read` | `GET /reports/{reportId}/audit` | | | | yes | |
| `department.list` | `GET /departments` | | yes | | yes | |
| `routing_rule.read` | `GET /admin/routing-rules` | | | | yes | |
| `routing_rule.write` | `PUT /admin/routing-rules` | | | | yes | |
| `sla_policy.read` | `GET /admin/sla-policies/{category}` | | | | yes | |
| `sla_policy.write` | `PUT /admin/sla-policies/{category}` | | | | yes | |
| `dashboard.read` | `GET /dashboard/summary`, `GET /dashboard/categories` | | | | yes | yes |

The first five action names are the same as in `services/api-gateway/domain/auth.go`.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `CheckPermission` | Returns `allowed` and, when denied, a `reason` | `INVALID_ARGUMENT` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Role`, action names, `Policy` with the role table and `Check` (100% tested) |
| `application/` | Use case: `CheckPermission` |
| `ports/` | `PolicyRepository` interface |
| `adapters/in/grpc/` | gRPC handler, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory store with the default table, used when `DATABASE_URL` is not set |
| `adapters/out/postgres/` | PostgreSQL store, used when `DATABASE_URL` is set |
| `adapters/out/repotest/` | Checks both stores must pass |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9017` | gRPC port |
| `SERVICE_NAME` | `authorization` | Name written in every log line |
| `DATABASE_URL` | none | For example `postgres://jurwa:<password>@localhost:5432/authorization_db?sslmode=disable`. Unset: built-in default table. A wrong URL stops the service instead of falling back. The password comes from the environment (a Kubernetes Secret), never from the repository |

Storage: when `DATABASE_URL` is set the table is kept in PostgreSQL (`adapters/out/postgres`, database `authorization_db`, table `role_permissions`). On start the service runs `docs/design/schema/authorization.sql` (embedded as `adapters/out/postgres/schema.sql`, a test keeps the two equal), which creates the tables if needed and seeds the default table once. Without `DATABASE_URL` it uses the built-in default table. Both adapters pass the same checks (`adapters/out/repotest`), which prove the seeded SQL table is the same as `DefaultPolicy`.

## Run (PowerShell)

```powershell
go run ./services/authorization/cmd/server
```

In a second terminal:

```powershell
grpcurl -plaintext localhost:9017 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9017 list
grpcurl -plaintext -d '{\"role\":\"STUDENT\",\"action\":\"report.submit\",\"resource\":\"report\"}' localhost:9017 jurwa.authorization.v1.AuthorizationService/CheckPermission
grpcurl -plaintext -d '{\"role\":\"PRESIDENT\",\"action\":\"report.reassign\",\"resource\":\"report\"}' localhost:9017 jurwa.authorization.v1.AuthorizationService/CheckPermission
```

Docker (from the repo root):

```powershell
docker build -f services/authorization/Dockerfile --build-arg SERVICE=authorization -t jurwa/authorization .
docker run --rm -p 9017:9017 jurwa/authorization
```

## Test

```powershell
go test ./services/authorization/... -cover
```

Integration tests on a real PostgreSQL 16 (Docker Desktop must be running):

```powershell
go test -tags=integration ./services/authorization/adapters/out/postgres/... -v
```

Evidence: `docs/evidence/gayley/sprint3/db-adapters-integration.txt` and `db-adapters-live.txt`.
