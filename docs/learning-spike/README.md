# Learning spike: Kubernetes, gRPC, Knative (SCRUM-71)

Author: Gayley Choden. Date: 26 October 2026.

## What I built

A small Go gRPC service (`hello/`) with one method, `Greeter.SayHello`, plus the standard gRPC health check. I generated the Go code from `hello.proto` with protoc, built a Docker image, loaded it into a local k3d cluster, deployed it with `k8s.yaml` and called it with grpcurl.

| File | Purpose |
|---|---|
| `hello/hello.proto` | The contract: the `Greeter` service and its two messages |
| `hello/gen/hellov1/` | Go code generated from the proto by protoc (do not edit) |
| `hello/main.go` | The server: SayHello, health check, reflection, JSON logs, port from `PORT` |
| `hello/go.mod` | A separate Go module so the spike never touches the main Jurwa module |
| `hello/Dockerfile` | Two-stage build: compile with Go, then copy only the program into a small Alpine image |
| `hello/k8s.yaml` | A Deployment (keeps one pod running) and a Service (stable name and port) |

Evidence (real command output) is in `docs/evidence/gayley/spike-1-nodes.txt`, `spike-2-pods.txt` and `spike-3-grpcurl.txt`.

## How to run it (Git Bash, Docker Desktop running)

```bash
cd docs/learning-spike/hello
protoc --go_out=. --go_opt=module=hello --go-grpc_out=. --go-grpc_opt=module=hello hello.proto
k3d cluster create spike
kubectl get nodes
docker build -t hello:1 .
k3d image import hello:1 -c spike
kubectl apply -f k8s.yaml
kubectl get pods
kubectl port-forward svc/hello 50051:50051      # leave running, use a second terminal below
grpcurl -plaintext -d '{"name":"Gayley"}' localhost:50051 hello.v1.Greeter/SayHello
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
k3d cluster delete spike
```

Expected: the pod shows `1/1 Running`, SayHello returns `"Hello, Gayley!"` and the health check returns `SERVING`.

## Kubernetes basics

- **Pod:** the smallest thing Kubernetes runs, one or more containers sharing a network address. Pods are temporary; a replaced pod gets a new IP.
- **Deployment:** the desired state ("keep 1 copy of `hello:1` running"). If the pod crashes, the Deployment starts a new one.
- **Service:** a stable name and port (`hello:50051`) that forwards to every pod whose label matches its selector (`app: hello`), so callers never need pod IPs.
- **Readiness probe:** Kubernetes calls the gRPC health check and only sends traffic once the pod answers `SERVING`. This is why every Jurwa service implements `grpc.health.v1`.

## Why gRPC with Protocol Buffers between services

- The `.proto` file is a typed contract that both sides generate code from, so a mismatch fails at compile time instead of at runtime.
- Messages are binary and small, so calls are faster and lighter than JSON over HTTP.
- HTTP/2 carries many calls over one connection.
- Standard extras come for free: the health service for Kubernetes and reflection for tools like grpcurl.

## What Knative adds

- **Scale to zero:** a Knative Service runs no pods while idle and starts one when a request or event arrives. Our two functions only run a few times a day, so they cost nothing between runs.
- **PingSource:** sends an event on a cron schedule. In Jurwa it triggers `sla-escalation-checker` (find reports past their SLA) and `digest-dispatcher` (send summary digests).
- **Fallback (risk R3):** if Knative is unstable on k3d, the same code can run as a Kubernetes CronJob.

## Problems found

1. **kubectl could not reach the cluster on Windows.** k3d wrote the API address as `host.docker.internal`, which on the campus network resolved to a LAN IP (10.2.38.102) and timed out. Fix: point kubectl at localhost, using the port from `kubectl config view`:

   ```bash
   kubectl config set-cluster k3d-spike --server=https://127.0.0.1:<api-port>
   ```

   Creating the cluster with `k3d cluster create spike --api-port 127.0.0.1:6550` avoids the problem from the start. Sonam Choki should use this in the SCRUM-100 k3d config.
2. **The first cluster creation took about 7 minutes** while it downloaded images. Create the demo cluster well before the demo (risk R3).
3. **`go mod tidy` raised the Go version** in `go.mod` to 1.25 because the current gRPC library needs it. The Dockerfile base image must match (`golang:1.25-alpine`).

## What I would do differently in Jurwa

- Start every service from the shared template so layout, logging and health checks are the same everywhere.
- Use the agreed port per service (identity is 9016) instead of 50051.
- Keep one Go module for the whole repo and commit generated code in `proto/gen/`.
- Read secrets such as `JWT_SECRET` from Kubernetes Secrets, never from code.
- Push images to GitHub Container Registry and let Argo CD deploy them, instead of `k3d image import` and `kubectl apply` by hand.
