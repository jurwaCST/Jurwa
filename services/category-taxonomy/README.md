# category-taxonomy

## What it does

Holds the three report categories shown on the PWA form, in a fixed order, with their display name, help text and Category Head position:

| Order | Category | Name | Category Head |
|---|---|---|---|
| 1 | `HOSTEL` | Hostel | DSA |
| 2 | `ACADEMIC_BLOCK` | Academic Block | DAA |
| 3 | `OTHERS` | Others | SSO |

A System Administrator can change the display name and help text (`UpsertCategory`). The category codes, their order and their Category Head never change, and none of the three can be removed, because routing-rules and the report form depend on them. Who holds each head position is kept in department-directory, not here.

Storage: when `DATABASE_URL` is set the categories are kept in PostgreSQL (`adapters/out/postgres`, database `category_taxonomy`). On start the service runs the table design from `docs/design/schema/category-taxonomy.sql` (embedded as `adapters/out/postgres/schema.sql`, a test keeps the two equal), which creates the table if needed, adds the `head_position` column to an older table, and seeds the three categories once. Without `DATABASE_URL` it keeps them in memory (`adapters/out/memory`), and a rename is lost on restart. Both adapters pass the same checks (`adapters/out/repotest`).

## RPCs

Contract: `proto/category-taxonomy/v1/category-taxonomy.proto`, service `jurwa.categorytaxonomy.v1.CategoryTaxonomyService`.

| RPC | What it does | Errors |
|---|---|---|
| `ListCategories` | The three categories in display order | none |
| `GetCategory` | One category with its name and Category Head | INVALID_ARGUMENT for `CATEGORY_UNSPECIFIED`, NOT_FOUND for a value not in the list |
| `UpsertCategory` | Change the display name (and help text) of a category | INVALID_ARGUMENT for an unspecified category or a blank name, PERMISSION_DENIED unless `actor_role` is `SYSTEM_ADMIN`, NOT_FOUND for a new category |

`actor_role` is filled in by api-gateway from the caller's JWT.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9009` | gRPC port |
| `SERVICE_NAME` | `category-taxonomy` | Name written in every log line |

`DATABASE_URL` (for example `postgres://jurwa:<password>@localhost:5432/category_taxonomy?sslmode=disable`): use PostgreSQL. Unset: keep the categories in memory. A wrong URL stops the service instead of falling back to memory. The password comes from the environment (a Kubernetes Secret), never from the repository.

Integration tests on a real PostgreSQL 16 (Docker Desktop must be running):

```bash
go test -tags=integration ./services/category-taxonomy/adapters/out/postgres/... -v
```

## Run

```bash
make run SERVICE=category-taxonomy
grpcurl -plaintext localhost:9009 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9009 list
grpcurl -plaintext localhost:9009 jurwa.categorytaxonomy.v1.CategoryTaxonomyService/ListCategories
grpcurl -plaintext -d '{"category":"HOSTEL"}' localhost:9009 jurwa.categorytaxonomy.v1.CategoryTaxonomyService/GetCategory
grpcurl -plaintext -d '{"category":"HOSTEL","display_name":"Hostels","actor_role":"SYSTEM_ADMIN"}' localhost:9009 jurwa.categorytaxonomy.v1.CategoryTaxonomyService/UpsertCategory
```

Docker:

```bash
make docker-build SERVICE=category-taxonomy
docker run --rm -p 9009:9009 jurwa/category-taxonomy
```

## Test

```bash
make test
go test ./services/category-taxonomy/... -cover
```

The domain (`domain/`) is written test-first (commit "SCRUM-24 tests first") and has 100% statement coverage.
