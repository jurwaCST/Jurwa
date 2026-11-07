# SonarQube and the coverage quality gate (SCRUM-111)

Owner: Sonam Choki. Proposal objective 4: at least 70% coverage on critical code, tracked in SonarQube Community Build.

There are two parts, because the CI runners on GitHub cannot reach a SonarQube that runs on a laptop:

| Part | Where | What it does |
|---|---|---|
| Coverage gate | CI, every pull request (`.github/workflows/quality-gate.yml`) | `scripts/check-coverage.sh` fails the pull request when a critical package is under 70% or its tests fail |
| SonarQube | A laptop, before a review or a sprint review | Full scan: bugs, vulnerabilities, code smells, duplication and coverage for the whole repository |

## 1. The coverage gate

Critical packages are the `domain` and `application` packages of the routing, SLA and status-flow services: routing-rules, routing-engine, sla-policy, sla-tracking, status-lifecycle (10 packages).

```bash
bash scripts/check-coverage.sh                  # 70% by default
MIN_COVERAGE=80 bash scripts/check-coverage.sh  # stricter
bash scripts/tests/check-coverage_test.sh       # tests of the script itself (go is faked)
```

Each package is OK, LOW (under the minimum), FAIL (its tests fail) or NO TESTS. The exit code is 1 when any package is not OK.

Result on 6 Nov 2026 (`docs/evidence/sonam-choki/sprint3/coverage-gate.txt`):

| Package | Coverage |
|---|---|
| routing-rules domain, application | 100%, 100% |
| routing-engine domain, application | 100%, 100% |
| sla-policy domain, application | 100%, 100% |
| sla-tracking domain, application | 100%, 98.1% |
| status-lifecycle domain, application | 100%, 96.9% |

10 of 10 at or above 70%. The workflow also prints the total for the whole repository (`go tool cover -func`), 35.1% today, and keeps `coverage.out` as an artifact. That total is low because it also counts `cmd/server/main.go` files, adapters without unit tests and the generated `proto/gen` code.

## 2. SonarQube on a laptop

Start it (first start takes a few minutes):

```bash
docker run -d --name sonarqube -p 9000:9000 -v sonarqube_data:/opt/sonarqube/data sonarqube:26.11.0.134175-community
curl -s localhost:9000/api/system/status      # wait for "status":"UP"
```

Open http://localhost:9000, sign in as `admin` / `admin` and set a new password. Create a local project with key `jurwa` (new code: follow the instance default), choose "Locally" and generate a token. Keep the token out of Git and out of chat.

Scan from the repo root:

```bash
go test ./... -coverprofile=coverage.out
export SONAR_TOKEN=<your token>
docker run --rm \
  -e SONAR_HOST_URL=http://host.docker.internal:9000 \
  -e SONAR_TOKEN="$SONAR_TOKEN" \
  -v "$PWD:/usr/src" \
  sonarsource/sonar-scanner-cli:12.3
```

`host.docker.internal` is how a container on Docker Desktop (macOS, Windows) reaches the laptop. On Linux use `--network host` and `http://localhost:9000`. The settings are in `sonar-project.properties`: generated protobuf code, mocks, tests and `web/node_modules` are excluded from the sources, and coverage comes from `coverage.out`. Warnings like `File '.../proto/gen/...' is not included in the project, ignoring coverage` are expected for the excluded files.

Next time: start the container again with `docker start sonarqube`.

## 3. Results (scan of 6 Nov 2026, Community Build 26.11.0.134175)

| Measure | Result |
|---|---|
| Security | 3 open issues, rating C |
| Reliability | 4 open issues, rating C |
| Maintainability | 5 open issues, rating A |
| Coverage | 61.7% on 5.3k lines to cover |
| Duplications | 7.8% on 19k lines |
| Security Hotspots | none to review |

SonarQube's 61.7% is higher than the `go tool cover` total because generated code and mocks are excluded from the SonarQube sources.

**No issue is in the Go services or in a critical package.** All 7 issues:

| # | File | Issue | Severity | Decision |
|---|---|---|---|---|
| 1 | `docs/learning-spike/hello/k8s.yaml` | Service account token is mounted automatically | Security, Medium | Noted: learning-spike file, never deployed. **The same rule is applied to our real manifests:** every generated Deployment now has `automountServiceAccountToken: false` (no Jurwa service calls the Kubernetes API). All 20 services redeployed and UP |
| 2 to 5 | `docs/learning-spike/hello/k8s.yaml` | No memory limit; no CPU, memory or storage request | Security and Reliability, Medium | Noted: learning-spike file. Our generated Deployments already set CPU and memory requests and limits |
| 6 | `web/src/api/client.ts` | `/\/+$/` can be slow on a long run of slashes (super-linear regular expression) | Reliability, Medium | Given to the owner (Gayley). Low risk: the input is the configured API address, not user input |
| 7 | `docs/learning-spike/hello/Dockerfile` | `alpine` image runs as root | Security, Low | Noted: learning-spike file. Our service images use `distroless/static:nonroot` and run as a non-root user |

Screenshots: `docs/evidence/sonam-choki/sprint3/sonarqube-*.png`.
