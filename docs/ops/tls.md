# HTTPS with Traefik and cert-manager (SCRUM-103)

Owner: Sonam Choki. The api-gateway is reached over HTTPS at `https://jurwa.localtest.me/api/...`. `jurwa.localtest.me` is a public name that always points to 127.0.0.1, so no hosts file change is needed.

## How it works

```
browser --HTTPS 443--> Traefik (built into k3d) --/api stripped--> api-gateway:8080
                         |
                         certificate from Secret jurwa-tls, made by cert-manager
```

| File in `deploy/k8s/ingress/` | What it does |
|---|---|
| `cluster-issuer.yaml` | `ClusterIssuer` `jurwa-selfsigned`: cert-manager signs certificates itself |
| `certificate.yaml` | `Certificate` for `jurwa.localtest.me`, stored in Secret `jurwa-tls`, valid 90 days, renewed 15 days before the end |
| `middleware.yaml` | Traefik middlewares: `strip-api` (`/api/categories` reaches the gateway as `/categories`) and `redirect-https` (http to https) |
| `tls-store.yaml` | Makes `jurwa-tls` Traefik's default certificate, so a request by IP address gets it too |
| `ingress.yaml` | `/api` to `api-gateway` port `http` (8080), for the host and for any host (phone by IP). Port 80 only redirects to HTTPS |

The web app (path `/`) is not routed yet. When its Service exists, add a `/` path to `ingress.yaml` and build the app with `VITE_API_URL=/api`.

## Commands

```bash
bash scripts/install-cert-manager.sh          # once per cluster, cert-manager v1.21.3
bash scripts/create-secrets.sh
kubectl apply -k deploy/k8s                   # includes deploy/k8s/ingress
kubectl -n jurwa get certificate              # READY True
curl -k -i https://jurwa.localtest.me/api/health
curl -k -i https://jurwa.localtest.me/api/categories -H "Authorization: Bearer $TOKEN"
```

`-k` tells curl to accept the self-signed certificate.

## Browser warning (expected)

Chrome shows "Your connection is not private" (`NET::ERR_CERT_AUTHORITY_INVALID`) because no public certificate authority signed the certificate. The connection is still encrypted. Click Advanced, then Proceed. Screenshots: `docs/evidence/sonam-choki/sprint3/traefik-tls-1-browser-warning.png` and `traefik-tls-2-browser-health-200.png`.

## Phone test (6 Nov 2026, iPhone 13, Safari, same Wi-Fi, laptop IP 10.2.2.254)

| Step | Result |
|---|---|
| Open `https://10.2.2.254/api/health` | Safari: "This Connection Is Not Private" (expected, self-signed) |
| Show Details, Visit Website | Safari accepted the certificate and reached Traefik on the laptop: `404 page not found`, which is Traefik's answer for `/` (the web app is not deployed yet) |
| Open `https://10.2.2.254/api/health` again | Did not show `{"status":"ok"}` on the phone. The same address returns 200 from the laptop |

So the phone reaches the cluster over HTTPS, but the API path did not load on the phone, and the cause is not found yet. Screenshots: `traefik-tls-3-iphone-warning.png`, `traefik-tls-4-iphone-404.png`.

**PWA install:** browsers only install a PWA from a page with a trusted certificate (or from `http://localhost`). A self-signed certificate is not trusted, so installing from the phone over this setup is not expected to work. The Lighthouse PWA check (Nyendrak, Gayley) runs on the laptop against `http://localhost`, which Chrome treats as secure.

## Troubleshooting

| Problem | Fix |
|---|---|
| `no matches for kind "ClusterIssuer"` | cert-manager is not installed: `bash scripts/install-cert-manager.sh` |
| Certificate `READY False` | `kubectl -n jurwa describe certificate jurwa-tls` and `kubectl -n cert-manager logs deploy/cert-manager` |
| `404 page not found` | Traefik has no route for that path. Only `/api/...` is routed |
| Connection refused on 443 | The cluster was created without `--port 443:443@loadbalancer`. Recreate it with `deploy/k3d/create-cluster.sh` |
