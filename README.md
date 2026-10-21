# Jurwa

A Progressive Web App for campus facility issue reporting at CST. Students report a broken facility (category, location, description, optional photo). The system routes it to the right Category Head, tracks it through its status flow, and escalates it automatically when its SLA window passes.

Group 6, BE Software Engineering Year III. Modules: SWE303, WEB303, SWE302.

## Repository layout

| Folder | What goes here |
|---|---|
| `services/` | One folder per service, e.g. `services/report-submission/`. Built from the shared Go template |
| `proto/` | gRPC `.proto` contracts shared by all services |
| `deploy/` | Dockerfiles, Kubernetes manifests, k3d, Argo CD, Knative |
| `web/` | React PWA |
| `docs/` | Project documents, sprint review and retro notes |

## Services and owners (20 services + 2 serverless functions)

| Owner | Services | Serverless function |
|---|---|---|
| Sonam Choki | `report-submission`, `photo-media`, `report-history`, `location-taxonomy`, `internal-notes` | none |
| Sonam Zangmo | `routing-rules`, `routing-engine`, `department-directory`, `category-taxonomy`, `api-gateway` | none |
| Nyendrak | `status-lifecycle`, `sla-tracking`, `sla-policy`, `reassignment`, `audit-log` | `sla-escalation-checker` |
| Gayley | `identity`, `authorization`, `notification`, `feedback`, `analytics` | `digest-dispatcher` |

## Report form rules

- Category: required
- Location: required (hostel, block, wing, area)
- Description: required, free text, about 500 characters at most
- Photo: optional (JPG or PNG, about 5 MB at most)
- Routing uses only category and location, never the description

## Getting started

1. Install: Go, Docker Desktop, k3d, kubectl, protoc (with the Go gRPC plugins), Node.
2. Clone the repo: `git clone https://github.com/jurwaCST/Jurwa.git`
3. Read `CONTRIBUTING.md` before your first branch.
4. Find your cards on the Jira board and make one branch per card.

## Sprints

| Sprint | Dates | Focus |
|---|---|---|
| Sprint 1 | 19-25 Oct | Design, planning, first working slice, k3d and CI |
| Sprint 2 | 26 Oct-1 Nov | Build everything |
| Sprint 3 | 2-8 Nov | Deployment, testing, presentation |
