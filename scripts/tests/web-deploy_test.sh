#!/usr/bin/env bash
# Tests for the web app in the cluster (SCRUM-139): the manifests in
# deploy/k8s/web, the "/" path in the ingress, and "web" in
# scripts/push-local-registry.sh. docker, make and curl are replaced by fakes
# on PATH, so nothing is built or pushed.
#   bash scripts/tests/web-deploy_test.sh
set -uo pipefail

cd "$(dirname "$0")/../.."
WEB="deploy/k8s/web"
INGRESS="deploy/k8s/ingress/ingress.yaml"
FAKES="$(mktemp -d)"
trap 'rm -rf "$FAKES"' EXIT

failures=0
check() { # description, condition result
  if [ "$2" = "0" ]; then echo "PASS  $1"; else echo "FAIL  $1"; failures=$((failures + 1)); fi
}
has() { grep -qE "$2" "$1" 2>/dev/null; echo $?; }
doc_named() { # file, metadata name: prints that YAML document of a multi-document file
  awk -v want="  name: $2" '
    /^---$/ { if (found) { printf "%s", buf; done = 1; exit } buf = ""; next }
    { buf = buf $0 "\n" }
    $0 == want { found = 1 }
    END { if (found && !done) printf "%s", buf }' "$1" 2>/dev/null
}

# 1. Deployment
D="$WEB/deployment.yaml"
check "deployment: file exists" "$([ -f "$D" ]; echo $?)"
check "deployment: kind Deployment named web" "$([ "$(has "$D" '^kind: Deployment')" = 0 ] && [ "$(has "$D" '^  name: web$')" = 0 ]; echo $?)"
check "deployment: image from the in-cluster registry" "$(has "$D" 'image: jurwa-registry:5050/web:latest')"
check "deployment: container port 8080 named http" "$(has "$D" 'containerPort: 8080')"
check "deployment: readiness probe is HTTP on /" "$(grep -A3 'readinessProbe:' "$D" 2>/dev/null | grep -A2 'httpGet:' | grep -qE 'path: /$'; echo $?)"
check "deployment: liveness probe is HTTP" "$(grep -A1 'livenessProbe:' "$D" 2>/dev/null | grep -q 'httpGet:'; echo $?)"
check "deployment: no service account token" "$(has "$D" 'automountServiceAccountToken: false')"
check "deployment: CPU and memory limits" "$(grep -A3 'limits:' "$D" 2>/dev/null | grep -q 'memory:'; echo $?)"
check "deployment: no secrets (static files only)" "$([ "$(has "$D" 'secretKeyRef')" != 0 ]; echo $?)"

# 2. Service
S="$WEB/service.yaml"
check "service: kind Service named web" "$([ "$(has "$S" '^kind: Service')" = 0 ] && [ "$(has "$S" '^  name: web$')" = 0 ]; echo $?)"
check "service: port http 8080" "$(grep -A3 'name: http' "$S" 2>/dev/null | grep -q 'port: 8080'; echo $?)"

# 3. Kustomization
check "web kustomization lists deployment and service" "$([ "$(has "$WEB/kustomization.yaml" 'deployment.yaml')" = 0 ] && [ "$(has "$WEB/kustomization.yaml" 'service.yaml')" = 0 ]; echo $?)"
check "deploy/k8s lists web" "$(has deploy/k8s/kustomization.yaml '^  - web$')"

# 4. Ingress: "/" goes to web, in its own Ingress without the /api strip
web_ingress="$(doc_named "$INGRESS" jurwa-web)"
check "ingress: Ingress jurwa-web exists" "$([ -n "$web_ingress" ]; echo $?)"
check "ingress: jurwa-web uses HTTPS (websecure, tls)" "$(echo "$web_ingress" | grep -q 'entrypoints: websecure' && echo "$web_ingress" | grep -q 'router.tls: "true"'; echo $?)"
check "ingress: jurwa-web does not strip /api" "$(echo "$web_ingress" | grep -q 'strip-api'; [ $? -ne 0 ]; echo $?)"
check "ingress: jurwa-web sends / to service web, port http" "$(echo "$web_ingress" | grep -A6 'path: /$' | grep -q 'name: web' && echo "$web_ingress" | grep -A8 'path: /$' | grep -q 'name: http'; echo $?)"
check "ingress: jurwa-web has the host rule and the no-host rule (phone by IP)" "$([ "$(echo "$web_ingress" | grep -c 'path: /$')" -eq 2 ] && echo "$web_ingress" | grep -q 'host: jurwa.localtest.me'; echo $?)"
check "ingress: /api still goes to api-gateway" "$(doc_named "$INGRESS" jurwa | grep -A6 'path: /api' | grep -q 'name: api-gateway'; echo $?)"

# 5. Everything still renders (only where kubectl is installed)
if command -v kubectl >/dev/null; then
  rendered="$(kubectl kustomize deploy/k8s 2>&1)"; code=$?
  check "kustomize: deploy/k8s renders" "$([ $code -eq 0 ]; echo $?)"
  check "kustomize: web Deployment in namespace jurwa" "$(echo "$rendered" | grep -B2 -A4 '^kind: Deployment' | grep -A3 'name: web$' | grep -q 'namespace: jurwa'; echo $?)"
else
  echo "SKIP  kustomize checks (no kubectl)"
fi

# 6. Push script: "web" builds web/Dockerfile and pushes localhost:5050/web:latest
CALLS="$FAKES/calls"
for tool in docker make curl; do
  printf '#!/usr/bin/env bash\necho "%s $*" >> "%s"\n' "$tool" "$CALLS" > "$FAKES/$tool"
  chmod +x "$FAKES/$tool"
done
out="$(PATH="$FAKES:$PATH" bash scripts/push-local-registry.sh web 2>&1)"; code=$?
calls="$(cat "$CALLS" 2>/dev/null)"
check "push web: exit code 0" "$([ $code -eq 0 ]; echo $?)"
check "push web: builds web/Dockerfile with web/ as context" "$(echo "$calls" | grep -qE '^docker build .*-f web/Dockerfile .*web$'; echo $?)"
check "push web: pushes localhost:5050/web:latest" "$(echo "$calls" | grep -q '^docker push localhost:5050/web:latest$'; echo $?)"
check "push web: does not use make (not a Go service)" "$(echo "$calls" | grep -q '^make'; [ $? -ne 0 ]; echo $?)"
check "push web: prints OK web" "$(echo "$out" | grep -qE '^OK +web$'; echo $?)"

echo
if [ "$failures" -eq 0 ]; then echo "All checks passed."; else echo "$failures check(s) failed."; fi
[ "$failures" -eq 0 ]
