# Integration tests with Testcontainers

Card: SCRUM-106. Owner: Gayley Choden. Workflow: `.github/workflows/integration.yml`. Run with `make test-integration`.

Unit tests (`make test`) use the in-memory adapters and need nothing else. Integration tests check the real adapters (PostgreSQL, MongoDB, Redis) against a real database that Testcontainers starts in Docker and removes afterwards.

## The pattern

1. **Build tag.** The file is `repo_integration_test.go` (or `<store>_integration_test.go`) and its first line is `//go:build integration`. `make test` does not build it, so unit tests stay fast and need no Docker.
2. **Testcontainers.** The test starts the container with `testcontainers-go` (for PostgreSQL `postgres.Run(ctx, "postgres:16-alpine", ...)`), waits until it is ready, connects, runs the adapter's `Migrate`, and removes the container when the test ends (`testcontainers.CleanupContainer(t, c)` or `t.Cleanup`).
3. **One contract for both adapters.** The checks live once in `adapters/out/repotest/contract.go` (`repotest.Run(t, newRepo)`, where `newRepo` gives a fresh empty store). The in-memory adapter's test and the PostgreSQL adapter's test both call it, so the two stores behave the same. A test only the real database can do (for example "data survives a restart": open a second pool on the same container) goes next to it.
4. **Schema check.** `schema_test.go` (no tag, runs in `make test`) checks that the adapter's embedded `schema.sql` is the same as `docs/design/schema/<service>.sql`.

Example: `services/routing-rules/adapters/out/postgres/`.

## What is covered

On `main`, 6 November 2026:

| Store | Package | Owner |
|---|---|---|
| PostgreSQL | `services/category-taxonomy/adapters/out/postgres` | Sonam Zangmo |
| PostgreSQL | `services/routing-rules/adapters/out/postgres` | Sonam Zangmo |
| Redis | `pkg/eventbus` (Redis Streams event bus) | Sonam Choki |

In review (SCRUM-127, adds 5 PostgreSQL packages): identity, authorization, notification, feedback, analytics (Gayley Choden). A trial run of `main` plus SCRUM-127 passed locally: 126 packages ok.

**Not covered yet:**

- MongoDB: no integration test yet (report-submission and the other MongoDB services, Sonam Choki).
- Redis `sla:*` timers: no integration test yet (sla-tracking, Nyendrak Yoezer Zangmo).

When a store's adapter is merged, its integration test is picked up automatically: the workflow runs `./...` with the tag, so nothing needs to be added here except this table.

## Run locally

Docker Desktop must be running. From the repo root in Git Bash:

```bash
make test-integration
```

This runs every unit test and every integration test, with `-count=1` so nothing comes from the cache. To run one package: `go test -tags=integration ./services/<service>/adapters/out/postgres/... -v`.

The first run downloads the images (`postgres:16-alpine`, `redis:7-alpine`, `testcontainers/ryuk`), which can take a few minutes.

**Windows tip.** If a test panics with `rootless Docker is not supported on Windows`, Testcontainers did not find Docker Desktop in time (seen once in two runs while all packages ran at the same time). Tell it where Docker is, then run again:

```bash
export DOCKER_HOST=npipe:////./pipe/docker_engine
make test-integration
```

With this set, 3 runs out of 3 passed. The CI runner is Linux and does not need it.

## In CI

`.github/workflows/integration.yml` runs on every pull request and on every push to `main`, on `ubuntu-latest` (Docker is installed there). Steps: check out, set up Go from `go.mod`, `go vet -tags=integration ./...` (the normal `go vet` skips tagged files), then `make test-integration`. The job stops after 20 minutes.

Time locally: about 45 seconds for the whole run with the images already downloaded. Time in CI: 1m 51s in total (the job itself 1m 47s), first run on pull request #75, 6 November 2026, images not cached.

Evidence: `docs/evidence/gayley/sprint3/integration-all.txt` (local run) and `docs/evidence/gayley/sprint3/integration-ci.png` (the green workflow on #75).

## When a test fails in CI

Open the job log, find the first `--- FAIL` line and tell the test's owner in the group chat with that line. Do not disable or skip the test.
