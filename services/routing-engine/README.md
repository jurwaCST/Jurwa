# routing-engine

> Created from `services/_template`. Replace every TODO below, and delete this note.

## Start a new service (template users)

From the repo root, with your service name and port from the table below:

```bash
make new-service NAME=routing-rules PORT=9006
```

| Port | Service | Port | Service |
|---|---|---|---|
| 9001 | report-submission | 9011 | status-lifecycle |
| 9002 | photo-media | 9012 | sla-tracking |
| 9003 | report-history | 9013 | sla-policy |
| 9004 | location-taxonomy | 9014 | reassignment |
| 9005 | internal-notes | 9015 | audit-log |
| 9006 | routing-rules | 9016 | identity |
| 9007 | routing-engine | 9017 | authorization |
| 9008 | department-directory | 9018 | notification |
| 9009 | category-taxonomy | 9019 | feedback |
| 9010 | api-gateway (gRPC; HTTP is 8080) | 9020 | analytics |

This copies `services/_template` to `services/<name>` and fills in the name and port. Then:

1. Run `make proto` so your generated code in `proto/gen/<name>/v1/` is up to date.
2. Write domain tests first in `domain/`, then the domain code.
3. Add ports (interfaces) in `ports/`, adapters in `adapters/out/`, gRPC handlers in `adapters/in/grpc/`.
4. Register your gRPC service in `cmd/server/main.go` (see the TODO there).
5. Run `make test`, `make lint` and `make run SERVICE=<name>`.

Layout:

```
cmd/server/main.go     starts the gRPC server (health check, reflection, JSON logs, graceful shutdown)
domain/                business rules, plain Go, no framework imports
application/           use cases
ports/                 interfaces the core depends on
adapters/in/grpc/      gRPC handlers
adapters/out/          database, Redis and other adapters
Dockerfile             multi-stage image, built from the repo root
```

## What it does

TODO: one or two sentences on what this service owns and what it never does.

## RPCs

TODO: list the RPCs from `proto/routing-engine/v1/`, one line each.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9007` | gRPC port |
| `SERVICE_NAME` | `routing-engine` | Name written in every log line |

TODO: add `DATABASE_URL` and others when the service uses them. Never hard-code passwords.

## Run

```bash
make run SERVICE=routing-engine
grpcurl -plaintext localhost:9007 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9007 list
```

Docker:

```bash
make docker-build SERVICE=routing-engine
docker run --rm -p 9007:9007 jurwa/routing-engine
```

## Test

```bash
make test
```
