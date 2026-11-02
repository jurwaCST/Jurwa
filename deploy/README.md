# deploy

Owner: Sonam Choki. Local Kubernetes for Jurwa: a k3d cluster, the `jurwa` namespace and a reusable manifest pattern for every gRPC service.

```
deploy/
  k3d/create-cluster.sh        creates cluster "jurwa" (1 server, 2 agents, local registry, ports 80/443)
  k3d/delete-cluster.sh        deletes it
  k3d/health.proto             grpc.health.v1 contract, for grpcurl when a service has no reflection
  k8s/base/namespace.yaml      namespace "jurwa"
  k8s/services/_example/       Deployment + Service + kustomization to copy for each service
```

## Prerequisites

Docker Desktop running ("Engine running"), plus `k3d` and `kubectl`:

```bash
docker info --format '{{.ServerVersion}}'
k3d version
kubectl version --client
```

Run every command below from the repo root.

## 1. Create the cluster

```bash
./deploy/k3d/create-cluster.sh
kubectl get nodes
```

Expected: three nodes (`k3d-jurwa-server-0`, `k3d-jurwa-agent-0`, `k3d-jurwa-agent-1`), all `Ready`. Traefik (the ingress controller) is left enabled and listens on ports 80 and 443 of your laptop.

The script also starts a local image registry:

| Where | Image name to use |
|---|---|
| From your laptop (`docker push`) | `localhost:5050/<image>:<tag>` |
| Inside the cluster (manifests) | `jurwa-registry:5050/<image>:<tag>` |

Port 5050 is used because macOS AirPlay already takes 5000.

## 2. Create the namespace

```bash
kubectl apply -f deploy/k8s/base/namespace.yaml
kubectl get ns
```

Expected: `jurwa` is listed as `Active`.

## 3. Deploy the example

The example uses a public test image (`agnhost grpc-health-checking`) that serves only the standard gRPC health check. It stands in until the real services have images. Push it to the local registry, then apply:

```bash
docker pull registry.k8s.io/e2e-test-images/agnhost:2.53
docker tag registry.k8s.io/e2e-test-images/agnhost:2.53 localhost:5050/agnhost:2.53
docker push localhost:5050/agnhost:2.53

kubectl apply -k deploy/k8s/services/_example
kubectl -n jurwa rollout status deploy/example
kubectl -n jurwa get pods
```

Expected: one `example-...` pod, `1/1 Running`. `1/1` means the gRPC readiness probe passed.

Call the health check yourself (the placeholder has no reflection, so grpcurl is given the proto file):

```bash
kubectl -n jurwa port-forward svc/example 19001:9001 &
grpcurl -plaintext -import-path deploy/k3d -proto health.proto localhost:19001 grpc.health.v1.Health/Check
kill %1
```

Expected: `{ "status": "SERVING" }`.

## What the example shows

- **Deployment**: one replica; image from the local registry; config from a ConfigMap as environment variables (`envFrom`); liveness and readiness probes using the built-in Kubernetes gRPC probe (`grpc.health.v1`, which every Jurwa service implements); CPU and memory requests and limits.
- **Service**: a `ClusterIP` on the service's gRPC port, so other services reach it at `<name>.jurwa.svc.cluster.local:<port>`.
- **kustomization.yaml**: puts everything in namespace `jurwa` and generates the ConfigMap. The generated name gets a hash suffix, so changing a value restarts the pods.

## Adding a real service

1. Copy `deploy/k8s/services/_example/` to `deploy/k8s/services/<service>/`.
2. Replace `example` with the service name everywhere, and set the service's gRPC port from the team port list (9001 report-submission to 9020 analytics) in `containerPort`, both probes, the Service and `PORT`.
3. Set `image:` to `jurwa-registry:5050/<service>:<tag>` and remove the `args:` line.
4. Add the service's environment variables to `configMapGenerator`. Passwords and connection strings with secrets go in a Kubernetes Secret, never in Git.
5. `kubectl apply -k deploy/k8s/services/<service>`

Service owners: ask Sonam Choki through a pull request for changes under `deploy/`.

## Build all images

`scripts/docker-build-all.sh` runs `make docker-build SERVICE=<name>` for every folder in `services/` except `_template`, prints OK or FAILED per service and exits non-zero if any failed. Docker Desktop must be running.

```bash
bash scripts/docker-build-all.sh
```

Expected: one `OK` line per service, then `Built N image(s), 0 failed.` A failed build prints the path of its log. Send that log to the service owner; if every service fails, the template Dockerfile is the problem (Sonam Zangmo).

Check that one image really runs (report-submission, port 9001):

```bash
docker run --rm -p 9001:9001 jurwa/report-submission
grpcurl -plaintext localhost:9001 grpc.health.v1.Health/Check   # in a second terminal
```

Expected: `"status": "SERVING"`. Stop the container with Ctrl+C.

To use an image in the cluster, tag and push it to the local registry:

```bash
docker tag jurwa/report-submission localhost:5050/report-submission:dev
docker push localhost:5050/report-submission:dev
```

## Delete the cluster

```bash
./deploy/k3d/delete-cluster.sh
```

Each person runs their own cluster. Do not share clusters (Gayley's learning-spike cluster is named `spike`).

## Troubleshooting

| Problem | Fix |
|---|---|
| `failed to connect to the docker API` | Start Docker Desktop and wait for "Engine running". |
| `port is already allocated` for 80, 443 or 5050 | Find the program with `lsof -nP -iTCP:<port> -sTCP:LISTEN` and stop it, then delete and recreate the cluster. |
| Pod stuck in `ImagePullBackOff` | The image is not in the registry. Check with `curl -s localhost:5050/v2/_catalog` and push it again. |
| Pod `0/1 Running` | The readiness probe fails. Check the port matches in the Deployment, the probes and `PORT`; see `kubectl -n jurwa describe pod <name>`. |
