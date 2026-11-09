# department-directory

## What it does

Knows the 5 academic departments and every position a report can be routed or reassigned to, and who holds each one:

| Kind | Positions |
|---|---|
| `CATEGORY_HEAD` | `DSA` (Hostel), `DAA` (Academic Block), `SSO` (Others) |
| `HOD` | `HOD_ARCH`, `HOD_CIVIL`, `HOD_CT`, `HOD_EEE`, `HOD_MECH`, `HOD_SH`, one per department |
| `STAFF` | `SSO_MALE`, `SSO_FEMALE` |
| `COUNCILLOR` | `BLOCK_COUNCILLOR`, `MAINTENANCE_COUNCILLOR`, `DEPUTY_CHIEF_COUNCILLOR_MESS`, `CHIEF_COUNCILLOR` |

routing-engine calls `GetStaff` to find who holds the Category Head position of a rule. reassignment uses `GetHoD` and `GetStaff` to find the person a report is reassigned to. Position codes are case-insensitive (`dsa` finds `DSA`).

**Departments:** the 6 academic departments of CST, named as on https://cst.edu.bt/department/:

| Code | Name | HoD on the staff directory (2 Nov 2026) |
|---|---|---|
| `ARCH` | Architecture | yes |
| `CIVIL` | Civil & Environmental Engineering | yes |
| `CT` | Computing Technologies | yes |
| `EEE` | Electrical & Electronics Engineering | yes |
| `MECH` | Mechanical Engineering | none listed: the Programme Leader acts as HoD |
| `SH` | Science & Humanities | yes |

The proposal mentions 5 departmental HoDs, and https://cst.edu.bt/staff-directory/ lists exactly 5. Mechanical Engineering has none listed, so its Programme Leader acts as HoD: the position keeps the code `HOD_MECH` (reassignment works the same way) and its title is "Head of Department, Mechanical Engineering (Programme Leader, acting)". Staff names are not stored here: holders are user ids from identity, assigned by the System Administrator.

**Demo holders:** the positions are held by the demo accounts that identity seeds (`SEED_DEMO_USERS`): DSA, DAA and SSO by the three Category Heads (`00000000-0000-4000-8000-00000000000{1,2,3}`), so a report is routed end to end, and `SSO_MALE` by the staff account (`...000000000010`), so a Category Head can reassign a report to a staff member who can log in. Every other position has no holder.

Sprint 2 keeps everything in memory (`adapters/out/memory`). The PostgreSQL adapter from `docs/design/schema/department-directory.sql` comes in Sprint 3.

## RPCs

Contract: `proto/department-directory/v1/department-directory.proto`, service `jurwa.departmentdirectory.v1.DepartmentDirectoryService`.

| RPC | What it does | Errors |
|---|---|---|
| `ListDepartments` | The 6 academic departments | none |
| `GetHoD` | The HoD position of one department | INVALID_ARGUMENT for an empty `department_code`, NOT_FOUND for an unknown department |
| `GetStaff` | One position and its holder | INVALID_ARGUMENT for an empty `position_code`, NOT_FOUND for an unknown code |
| `ListPositions` | Every position ordered by code; with `held_only` only the held ones (the reassign picker in the PWA, through api-gateway `GET /positions`) | none |

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
grpcurl -plaintext -d '{"department_code":"CT"}' localhost:9008 jurwa.departmentdirectory.v1.DepartmentDirectoryService/GetHoD
grpcurl -plaintext -d '{"position_code":"DSA"}' localhost:9008 jurwa.departmentdirectory.v1.DepartmentDirectoryService/GetStaff
grpcurl -plaintext -d '{"held_only":true}' localhost:9008 jurwa.departmentdirectory.v1.DepartmentDirectoryService/ListPositions
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
