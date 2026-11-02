# department-directory

## What it does

Knows the 5 academic departments and every position a report can be routed or reassigned to, and who holds each one:

| Kind | Positions |
|---|---|
| `CATEGORY_HEAD` | `DSA` (Hostel), `DAA` (Academic Block), `SSO` (Others) |
| `HOD` | `HOD_DEPT_A` to `HOD_DEPT_E`, one per department |
| `STAFF` | `SSO_MALE`, `SSO_FEMALE` |
| `COUNCILLOR` | `BLOCK_COUNCILLOR`, `MAINTENANCE_COUNCILLOR`, `DEPUTY_CHIEF_COUNCILLOR_MESS`, `CHIEF_COUNCILLOR` |

routing-engine calls `GetStaff` to find who holds the Category Head position of a rule. reassignment uses `GetHoD` and `GetStaff` to find the person a report is reassigned to. Position codes are case-insensitive (`dsa` finds `DSA`).

**Department names:** the real names of the 5 departments are not written in `docs/pm/requirements.md` or `docs/design/database-schema.md`, so they are seeded as `DEPT_A` to `DEPT_E` ("Department A" to "Department E"). Replace them when the names are confirmed.

**Demo holders:** identity has no seeded users yet, so DSA, DAA and SSO are given fixed demo user ids (`00000000-0000-4000-8000-00000000000{1,2,3}`) to let a report be routed end to end. Every other position has no holder. Real holders are entered by the System Administrator in Sprint 3.

Sprint 2 keeps everything in memory (`adapters/out/memory`). The PostgreSQL adapter from `docs/design/schema/department-directory.sql` comes in Sprint 3.

## RPCs

Contract: `proto/department-directory/v1/department-directory.proto`, service `jurwa.departmentdirectory.v1.DepartmentDirectoryService`.

| RPC | What it does | Errors |
|---|---|---|
| `ListDepartments` | The 5 departments | none |
| `GetHoD` | The HoD position of one department | INVALID_ARGUMENT for an empty `department_code`, NOT_FOUND for an unknown department |
| `GetStaff` | One position and its holder | INVALID_ARGUMENT for an empty `position_code`, NOT_FOUND for an unknown code |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9008` | gRPC port |
| `SERVICE_NAME` | `department-directory` | Name written in every log line |

`DATABASE_URL` is added with the PostgreSQL adapter in Sprint 3.

## Run

```bash
make run SERVICE=department-directory
grpcurl -plaintext localhost:9008 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9008 list
grpcurl -plaintext localhost:9008 jurwa.departmentdirectory.v1.DepartmentDirectoryService/ListDepartments
grpcurl -plaintext -d '{"department_code":"DEPT_A"}' localhost:9008 jurwa.departmentdirectory.v1.DepartmentDirectoryService/GetHoD
grpcurl -plaintext -d '{"position_code":"DSA"}' localhost:9008 jurwa.departmentdirectory.v1.DepartmentDirectoryService/GetStaff
```

Docker:

```bash
make docker-build SERVICE=department-directory
docker run --rm -p 9008:9008 jurwa/department-directory
```

## Test

```bash
make test
go test ./services/department-directory/... -cover
```

The domain (`domain/`) is written test-first (commit "SCRUM-32 tests first") and has 100% statement coverage.
