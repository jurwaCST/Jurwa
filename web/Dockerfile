# Jurwa web app (SCRUM-140, for the Deployment in SCRUM-139).
# Build from the repo root with web/ as the context:
#   docker build -f web/Dockerfile web
#
# Stage 1 builds the static files; stage 2 serves them with nginx on port 8080
# as a non-root user. Unknown paths fall back to index.html, so /login or
# /report/new work after a page reload. The app calls the API at /api, which
# Traefik routes to the api-gateway (deploy/k8s/ingress).

# Debian, not Alpine: TypeScript 7's compiler is a native binary built for glibc.
FROM node:24-bookworm-slim AS build
WORKDIR /app
COPY package.json package-lock.json ./
# The Cypress test browser is not needed to build the app.
ENV CYPRESS_INSTALL_BINARY=0
RUN npm ci --no-audit --no-fund
COPY . .
ARG VITE_API_URL=/api
ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

# nginx-unprivileged listens on 8080 and runs as user nginx (uid 101).
FROM nginxinc/nginx-unprivileged:1.29-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
USER 101
