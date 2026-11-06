# identity

Owner: Gayley Choden. Port: 9016. Card: SCRUM-13 (Sprint 1 thin slice).

## What it does

Holds user accounts, checks passwords and issues JSON Web Tokens (JWTs). Users sign in with their institutional email and a password held only by this service; there is no link to college systems. It never decides what a user may do; that is the authorization service.

- Passwords are stored only as bcrypt hashes and are never logged or returned.
- New accounts get the `STUDENT` role.
- Tokens are signed with HS256 using `JWT_SECRET` and expire after `JWT_TTL` (15 minutes by default).
- A wrong email and a wrong password give the same error, so callers cannot tell which accounts exist.

Sprint 1 stores users in memory, so they are lost when the server stops. PostgreSQL replaces this in Sprint 2.

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
adapters/out/memory/    in-memory UserRepository (Sprint 1)
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
| `SEED_DEMO_USERS` | off | `true` creates the demo accounts below at startup. **For the demo only, never enable it in production** |
| `SEED_DEMO_PASSWORD` | none, required when seeding | The one password of all demo accounts, at least 8 characters. There is no default. Never commit it |

## Demo accounts (SCRUM-126)

`Register` only ever makes Student accounts (the request has no role field), so the other roles come from a seed. With `SEED_DEMO_USERS=true` the service creates these accounts when it starts, through the same `UserRepository` port as Register (the password is bcrypt-hashed and never logged). Running it again creates nothing new; an account that already exists is left as it is. If `SEED_DEMO_PASSWORD` is missing or too short, the server stops with an error. If a demo email is already used by another account, it also stops.

| Email | Role | User id |
|---|---|---|
| `dsa.demo@jurwa.example` | Category Head | `00000000-0000-4000-8000-000000000001` |
| `daa.demo@jurwa.example` | Category Head | `00000000-0000-4000-8000-000000000002` |
| `sso.demo@jurwa.example` | Category Head | `00000000-0000-4000-8000-000000000003` |
| `staff.demo@jurwa.example` | Staff | `00000000-0000-4000-8000-000000000010` |
| `admin.demo@jurwa.example` | System Administrator | `00000000-0000-4000-8000-000000000020` |
| `president.demo@jurwa.example` | College President | `00000000-0000-4000-8000-000000000030` |

The three Category Head ids are the holders of DSA, DAA and SSO in department-directory, so a routed report goes to a user who can log in. `.example` is a reserved domain, so these are never real mailboxes. There is no RPC that creates a privileged user.

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

With the demo accounts (pick your own password):

```bash
SEED_DEMO_USERS=true SEED_DEMO_PASSWORD='<your demo password>' make run SERVICE=identity
grpcurl -plaintext -d '{"email":"dsa.demo@jurwa.example","password":"<your demo password>"}' localhost:9016 jurwa.identity.v1.IdentityService/Login
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

Demo seed tests (SCRUM-126): six accounts with the right roles and ids, a second run creates no duplicates, flag off creates nothing, missing or short password fails, a demo email held by another account fails, a seeded user logs in with the right role and id in the token, a wrong password fails, and Register still gives only Student.

Evidence of a real run: `docs/evidence/gayley/identity-grpcurl.txt`, demo seed: `docs/evidence/gayley/sprint3/identity-seed.txt`.
