# routing-rules

## What it does

Stores the routing rules: which Category Head position receives a report of each category, optionally for one building. Seeded with Hostel to DSA, Academic Block to DAA and Others to SSO.

It never reads the description (routing uses only category and location, FR-11) and never knows who holds a position; department-directory does. routing-engine calls `GetRule` to route a new report.

Matching: a rule for the exact building beats the category rule (building `""`). Among equally specific rules the lowest priority number wins (default 100). Inactive rules are ignored.

Storage: when `DATABASE_URL` is set the rules are kept in PostgreSQL (`adapters/out/postgres`, database `routing_rules`). On start the service runs the table design from `docs/design/schema/routing-rules.sql` (embedded as `adapters/out/postgres/schema.sql`, a test keeps the two equal), which creates the table if needed and seeds the three default rules once. Without `DATABASE_URL` it keeps them in memory (`adapters/out/memory`), and changes are lost on restart. Both adapters pass the same checks (`adapters/out/repotest`).

## RPCs

Contract: `proto/routing-rules/v1/routing-rules.proto`, service `jurwa.routingrules.v1.RoutingRulesService`.

| RPC | What it does | Errors |
|---|---|---|
| `GetRule` | Rule for a category and building | INVALID_ARGUMENT for an unspecified category, NOT_FOUND when no active rule matches |
| `ListRules` | All rules, optionally with inactive ones | none |
| `UpsertRule` | Create or replace the rule for (category, building) | INVALID_ARGUMENT for a missing category or target position |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9006` | gRPC port |
| `SERVICE_NAME` | `routing-rules` | Name written in every log line |

`DATABASE_URL` (for example `postgres://jurwa:<password>@localhost:5432/routing_rules?sslmode=disable`): use PostgreSQL. Unset: keep the rules in memory. A wrong URL stops the service instead of falling back to memory. The password comes from the environment (a Kubernetes Secret), never from the repository.

Integration tests on a real PostgreSQL 16 (Docker Desktop must be running):

```bash
go test -tags=integration ./services/routing-rules/adapters/out/postgres/... -v
```

## Run

```bash
make run SERVICE=routing-rules
grpcurl -plaintext localhost:9006 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9006 list
grpcurl -plaintext -d '{"category":"HOSTEL","location":{"building":"Hostel B"}}' localhost:9006 jurwa.routingrules.v1.RoutingRulesService/GetRule
grpcurl -plaintext localhost:9006 jurwa.routingrules.v1.RoutingRulesService/ListRules
```

Docker:

```bash
make docker-build SERVICE=routing-rules
docker run --rm -p 9006:9006 jurwa/routing-rules
```

## Test

```bash
make test
go test ./services/routing-rules/... -cover
```

The domain (`domain/`) is written test-first and has 100% statement coverage. `TestQueryHasNoDescription` fails if a description field is ever added to the routing input.
