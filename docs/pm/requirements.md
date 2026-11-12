# Requirements: Jurwa

Source: project proposal (7 October 2026), project charter (`docs/pm/charter.md`) and the team's Sprint 1 decisions.

Priorities: **Must** (needed for the final demo), **Should** (planned, cut only if the schedule slips), **Could** (only if time allows).

## 1. Report form rules

| Field | Rule |
|---|---|
| Category | Required. One of: Hostel, Academic Block, Others (campus-wide). |
| Location | Required. Hostel, block, wing, area, chosen from the location list. |
| Description | Required. Free text, at most about 500 characters. Empty or whitespace-only is rejected. |
| Photo | Optional. JPG or PNG, at most about 5 MB. |

Routing uses only category and location, never the description.

## 2. Routing rules

| Category | Routed to (Category Head) | Can reassign to |
|---|---|---|
| Hostel | DSA (Dean of Student Affairs) | SSOs (male and female), Block Councillor, Maintenance Councillor, Deputy Chief Councillor (Mess), Chief Councillor |
| Academic Block | DAA (Dean of Academic Affairs) | One of the 6 departmental HoDs (5 HoDs; Mechanical's Programme Leader acts as HoD) |
| Others (campus-wide) | SSO (Student Service Officer) | Relevant estate or administrative staff |

## 3. Status flow

Open, then Assigned, then In Progress, then Resolved. Escalated is set automatically when a report that is not yet Resolved passes its SLA window. The full list of valid and invalid transitions is owned by the status-lifecycle service.

## 4. Functional requirements

### Student

| Id | Requirement | Priority |
|---|---|---|
| FR-01 | A student signs in with an institutional email address and password and receives a JWT. | Must |
| FR-02 | A student submits a report with category, location, description and an optional photo. | Must |
| FR-03 | Category is required and must be one of Hostel, Academic Block, Others. Any other value is rejected. | Must |
| FR-04 | Location is required (hostel, block, wing, area) and is chosen from the location list. | Must |
| FR-05 | Description is required, free text, at most about 500 characters. Empty or whitespace-only descriptions are rejected. | Must |
| FR-06 | A photo is optional. If given, it must be JPG or PNG and at most about 5 MB, otherwise it is rejected. | Must |
| FR-07 | A submitted report gets a unique id and the status Open, and the student sees a confirmation. | Must |
| FR-08 | A student views their own reports with the current status and the history of status changes. | Must |
| FR-09 | A student is notified when the status of their report changes. | Should |
| FR-10 | A student rates the resolution after the report is Resolved. | Should |

### Routing and SLA (automatic)

| Id | Requirement | Priority |
|---|---|---|
| FR-11 | The system routes every new report using only category and location, never the description. | Must |
| FR-12 | Hostel reports go to the DSA, Academic Block reports to the DAA, Others to the SSO. The report becomes Assigned. | Must |
| FR-13 | Each category has an SLA window, stored as policy. | Must |
| FR-14 | An SLA timer starts when the report is routed and stops when it is Resolved. | Must |
| FR-15 | A scheduled serverless function finds reports past their SLA window and sets them to Escalated, publishing an SLA-lapsed event. | Must |
| FR-16 | If a later step of the escalation flow fails, the earlier step is undone by a compensating action (Saga), for example an escalation reverts to the previous status. | Must |
| FR-17 | Every status change, routing decision, reassignment and escalation is written to the audit log. | Must |

### Category Head (DSA, DAA, SSO)

| Id | Requirement | Priority |
|---|---|---|
| FR-18 | A Category Head sees a queue of reports assigned to them, with status badges, and filters it by status. | Must |
| FR-19 | A Category Head reassigns a report to a staff user. The DAA can reassign to one of the 6 departmental HoDs (5 HoDs; Mechanical's Programme Leader acts as HoD). | Must |
| FR-20 | A Category Head moves a report to In Progress or Resolved. Invalid transitions are rejected. | Must |
| FR-21 | A Category Head adds internal notes to a report. Internal notes are not visible to the student. | Should |

### Staff (councillors, HoDs, estate staff)

| Id | Requirement | Priority |
|---|---|---|
| FR-22 | Staff users sign in and see only the reports assigned to them. | Must |
| FR-23 | Staff users update the status of their reports and add internal notes. | Must |
| FR-24 | Staff without an institutional email (cooks, cleaners) are represented by their supervisor or HoD, who acts for them with a full audit trail (risk R5). | Could |

### System Administrator

| Id | Requirement | Priority |
|---|---|---|
| FR-25 | The administrator manages the routing rules (category and location to Category Head). | Must |
| FR-26 | The administrator manages users, roles and the department directory (Category Heads, HoDs, staff). | Must |
| FR-27 | The administrator manages the category and location lists. | Should |
| FR-28 | The administrator manages SLA windows per category. | Should |

### College President

| Id | Requirement | Priority |
|---|---|---|
| FR-29 | The College President sees a read-only dashboard: report counts by category and status, resolution rate and the list of escalated reports. | Must |
| FR-30 | The College President cannot change any report, rule or user. | Must |
| FR-31 | Category Heads and the College President receive a periodic digest email of open and escalated reports. | Could |

## 5. Non-functional requirements

| Id | Category | Requirement | Priority |
|---|---|---|---|
| NFR-01 | Performance | A student can submit a report with a photo in under two minutes, verified by usability testing (proposal objective 2). | Must |
| NFR-02 | Performance | Report submission is load-tested with Apache JMeter. Target response time is agreed by the team before the load test (to be confirmed). | Should |
| NFR-03 | Availability | Every service implements the gRPC health check, and Kubernetes liveness and readiness probes restart failed pods. All services are healthy for the live demo. | Must |
| NFR-04 | Security | Every request except sign-in carries a JWT, checked at the api-gateway. Roles are checked at the gateway and inside each service. | Must |
| NFR-05 | Security | Passwords are stored hashed. Secrets are Kubernetes Secrets and are never committed to Git. The gateway rate-limits requests. | Must |
| NFR-06 | Accessibility | The PWA meets WCAG basics, checked with axe-core and Lighthouse. | Must |
| NFR-07 | Usability | The PWA is installable on a phone, mobile-first, and works in any desktop browser. | Must |
| NFR-08 | Observability | Services write structured JSON logs. Metrics, dashboards and traces are available through Prometheus, Grafana and Jaeger. | Should |
| NFR-09 | Maintainability | Every service follows the shared hexagonal template, has its own database and talks to other services only through versioned gRPC contracts. | Must |
| NFR-10 | Quality | At least 70% test coverage on routing, SLA and status-flow code. Unit tests must pass in CI before merge. | Must |
| NFR-11 | Cost | Only free or open-source tools; zero budget. | Must |

## 6. Traceability

| Requirement | Implemented by |
|---|---|
| FR-01 | identity, api-gateway |
| FR-02 | report-submission, api-gateway |
| FR-03 | report-submission, category-taxonomy |
| FR-04 | report-submission, location-taxonomy |
| FR-05 | report-submission |
| FR-06 | photo-media |
| FR-07 | report-submission, status-lifecycle |
| FR-08 | report-history |
| FR-09 | notification |
| FR-10 | feedback |
| FR-11 | routing-engine |
| FR-12 | routing-engine, routing-rules, department-directory, status-lifecycle |
| FR-13 | sla-policy |
| FR-14 | sla-tracking |
| FR-15 | sla-escalation-checker (serverless), sla-tracking, status-lifecycle |
| FR-16 | sla-tracking, reassignment, status-lifecycle |
| FR-17 | audit-log |
| FR-18 | report-history, api-gateway |
| FR-19 | reassignment, department-directory |
| FR-20 | status-lifecycle |
| FR-21 | internal-notes |
| FR-22 | authorization, report-history |
| FR-23 | status-lifecycle, internal-notes |
| FR-24 | identity, authorization, audit-log |
| FR-25 | routing-rules |
| FR-26 | identity, authorization, department-directory |
| FR-27 | category-taxonomy, location-taxonomy |
| FR-28 | sla-policy |
| FR-29 | analytics |
| FR-30 | authorization |
| FR-31 | digest-dispatcher (serverless), notification |
| NFR-01 | report-submission, photo-media, api-gateway |
| NFR-02 | report-submission, api-gateway |
| NFR-03 | all services |
| NFR-04 | api-gateway, authorization |
| NFR-05 | identity, api-gateway |
| NFR-06 | PWA (`web/`) |
| NFR-07 | PWA (`web/`) |
| NFR-08 | all services |
| NFR-09 | all services |
| NFR-10 | routing-engine, routing-rules, sla-policy, sla-tracking, status-lifecycle |
| NFR-11 | whole project |

### Coverage check: every service appears at least once

| Service | Requirements |
|---|---|
| report-submission | FR-02 to FR-05, FR-07 |
| photo-media | FR-06 |
| report-history | FR-08, FR-18, FR-22 |
| location-taxonomy | FR-04, FR-27 |
| internal-notes | FR-21, FR-23 |
| routing-rules | FR-12, FR-25 |
| routing-engine | FR-11, FR-12 |
| department-directory | FR-12, FR-19, FR-26 |
| category-taxonomy | FR-03, FR-27 |
| api-gateway | FR-01, FR-02, FR-18, NFR-04 |
| status-lifecycle | FR-07, FR-12, FR-15, FR-16, FR-20, FR-23 |
| sla-tracking | FR-14, FR-15, FR-16 |
| sla-policy | FR-13, FR-28 |
| reassignment | FR-16, FR-19 |
| audit-log | FR-17, FR-24 |
| identity | FR-01, FR-24, FR-26 |
| authorization | FR-22, FR-26, FR-30 |
| notification | FR-09, FR-31 |
| feedback | FR-10 |
| analytics | FR-29 |
| sla-escalation-checker (serverless) | FR-15 |
| digest-dispatcher (serverless) | FR-31 |
