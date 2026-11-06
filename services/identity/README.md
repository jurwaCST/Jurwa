# identity

Owner: Gayley Choden. Port: 9016. Card: SCRUM-13 (Sprint 1 thin slice).

## What it does

Holds user accounts, checks passwords and issues JSON Web Tokens (JWTs). Users sign in with their institutional email and a password held only by this service; there is no link to college systems. It never decides what a user may do; that is the authorization service.

- Passwords are stored only as bcrypt hashes and are never logged or returned.
- New accounts get the `STUDENT` role.
- Tokens are signed with HS256 using `JWT_SECRET` and expire after `JWT_TTL` (15 minutes by default).
- A wrong email and a wrong password give the same error, so callers cannot tell which accounts exist.

Storage: when `DATABASE_URL` is set, users are kept in PostgreSQL (`adapters/out/postgres`, database `identity`, table `users`, which also holds each user's roles). On start the service runs `docs/design/schema/identity.sql` (embedded as `adapters/out/postgres/schema.sql`, a test keeps the two equal), which creates the table if needed. Without `DATABASE_URL` users are kept in memory and lost when the server stops. Both adapters pass the same checks (`adapters/out/repotest`).

## RPCs

From `proto/identity/v1/identity.proto`:

| RPC | What it does | Errors |
|---|---|---|
| `Register` | Creates a Student account | `INVALID_ARGUMENT` bad email, blank name, password under 8 characters or over 72 bytes; `ALREADY_EXISTS` email used |
| `Login` | Checks email and password, returns a JWT, its expiry and the user | `UNAUTHENTICATED` wrong email or password |
| `ValidateToken` | Checks a JWT and returns the user and roles (called by api-gateway) | `UNAUTHENTICATED` invalid, tampered or expired token, or the user no longer exists |

Plus the standard `grpc.health.v1.Health` service.

## Layout

```
domain/                 User, email and password rules, bcrypt, JWT signing and checking
application/            Register, Login, ValidateToken use cases
ports/                  UserRepository interface
adapters/out/memory/    in-memory UserRepository, used when DATABASE_URL is not set
adapters/out/postgres/  PostgreSQL UserRepository, used when DATABASE_URL is set
adapters/out/repotest/  checks both repositories must pass
adapters/in/grpc/       gRPC handler, maps domain errors to gRPC codes
cmd/server/             starts the server and wires everything together
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9016` | gRPC port |
| `SERVICE_NAME` | `identity` | Name written in every log line |
| `JWT_SECRET` | none, required | Signing secret, at least 32 bytes. The server will not start without it. Never commit it; in Kubernetes it comes from a Secret |
| `JWT_TTL` | `15m` | Token lifetime as a Go duration, for example `15m` or `1h` |
| `DATABASE_URL` | none | For example `postgres://jurwa:<password>@localhost:5432/identity?sslmode=disable`. Unset: users in memory. A wrong URL stops the service instead of falling back. Never commit the password; in Kubernetes it comes from a Secret |

## Run

Git Bash, from the repo root. The secret below is random and made only for your local run:

```bash
export JWT_SECRET=$(openssl rand -hex 32)
make run SERVICE=identity
```

In a second terminal:

```bash
grpcurl -plaintext localhost:9016 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{"email":"student@rub.edu.bt","full_name":"Test Student","password":"secret123"}' localhost:9016 jurwa.identity.v1.IdentityService/Register
grpcurl -plaintext -d '{"email":"student@rub.edu.bt","password":"secret123"}' localhost:9016 jurwa.identity.v1.IdentityService/Login
grpcurl -plaintext -d '{"access_token":"<accessToken from Login>"}' localhost:9016 jurwa.identity.v1.IdentityService/ValidateToken
```

Docker:

```bash
make docker-build SERVICE=identity
docker run --rm -p 9016:9016 -e JWT_SECRET=$(openssl rand -hex 32) jurwa/identity
```

## Test

```bash
make test
```

Table-driven tests cover: duplicate email, invalid email, short and long password, blank name, wrong password, unknown email, correct login returns a token, expired token, tampered payload, tampered signature, token signed with another secret, `alg: none` token, and the gRPC error codes. Coverage: domain 94.5%, application 87.5%, gRPC adapter 92.0%.

Integration tests on a real PostgreSQL 16 (Docker Desktop must be running):

```bash
go test -tags=integration ./services/identity/adapters/out/postgres/... -v
```

Evidence of a real run: `docs/evidence/gayley/identity-grpcurl.txt`. PostgreSQL: `docs/evidence/gayley/sprint3/db-adapters-integration.txt` and `db-adapters-live.txt`.
