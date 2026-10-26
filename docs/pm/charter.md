# Project Charter: Jurwa

| Field | Detail |
|---|---|
| Project name | Jurwa: A Progressive Web App for Campus Facility Issue Reporting |
| Organisation | College of Science and Technology (CST), Royal University of Bhutan |
| Programme | BE Software Engineering, Year III (Group 6) |
| Modules | SWE303 Software Project Management, WEB303 Microservices and Serverless Applications, SWE302 Software Testing and Quality Assurance |
| Sponsor | Module Tutors: Mr. Pema Namgay (WEB303, SWE302), Mr. Nima Yoezer (SWE303) |
| Team Lead | Sonam Zangmo |
| Scrum Master | Gayley Choden |
| Budget | Zero (free and open-source tools only) |
| PM tooling | Jira Free plan (backlog, sprints, hours) and GitHub (source control, CI/CD) |

## 1. Purpose

Replace informal group-chat reporting of campus facility issues with a trackable, automatically routed and automatically escalating system.

## 2. Background and problem

Campus life at CST depends on shared facilities: hostel plumbing and electrics, academic-block equipment, Wi-Fi and general estate infrastructure. Today the only way to report a fault is an informal group chat with whichever staff member is reachable.

Problem statement: campus facility issues are reported through informal, unrecorded group chats, so there is no automatic routing, no status tracking and no accountability when a report is ignored.

Problems identified:

- No record: a report is lost once the chat message scrolls past.
- No routing: the student must already know who to contact.
- No status visibility: the reporter cannot tell if the issue is being worked on.
- No escalation: an ignored report stays ignored.
- No accountability: nothing shows who was responsible for a delay.

## 3. Objectives

1. Build a microservices backend (20 services) that routes a report to the correct Category Head with 100% accuracy on the test rule-set.
2. Build an installable PWA in which a student can submit a report with a photo in under two minutes, verified by usability testing.
3. Implement automatic SLA escalation using a scheduled serverless function, shown in the final demo.
4. Build a test pyramid (unit, integration, end-to-end, performance, accessibility) tracked in SonarQube Community Build, with at least 70% coverage on routing, SLA and status-flow code.
5. Deploy the system as containers on Kubernetes (k3d) through a GitOps pipeline (GitHub Actions and Argo CD), shown in a live demo.
6. Manage the project within the schedule and a zero budget, tracking progress weekly with Earned Value Management (EVM).

## 4. Scope

### In scope

- Report submission (category, location, description, optional photo), automatic routing by category and location, reassignment and status tracking (Open, Assigned, In Progress, Resolved).
- Automatic escalation (status Escalated) when a report passes its SLA window.
- Category Head and staff queues, a System Administrator view and a read-only College President dashboard.
- Delivery as a Progressive Web App (installable on a phone, works in any desktop browser).
- 20 Go microservices with gRPC contracts, plus 2 serverless functions on Knative:

| Domain | Services |
|---|---|
| Identity and Access | identity, authorization |
| Reporting | report-submission, photo-media, report-history |
| Routing and Configuration | routing-rules, routing-engine, department-directory, category-taxonomy, location-taxonomy |
| Lifecycle and SLA | status-lifecycle, sla-tracking, sla-policy, reassignment, audit-log |
| Comms and Engagement | notification, feedback, analytics |
| Platform | api-gateway, internal-notes |
| Serverless (Knative) | sla-escalation-checker, digest-dispatcher |

### Out of scope

- Descoped from the proposal's 25 services: export, holiday-calendar, a separate digest service, a separate email-template service and a separate user-profile service. Digest emails are handled by the `digest-dispatcher` function.
- A native mobile app, payment features, SMS notifications and ML-based classification.
- Integration with existing institutional systems.

### Scope control

Scope is locked by this charter. Any scope change needs sign-off from the full team (risk R1) and is recorded in Jira.

## 5. Deliverables

- Project management documents: charter, stakeholder register, requirements, WBS, schedule, risk register, quality and communications plans, weekly EVM reports, closure report.
- Design: bounded contexts, gRPC/Protocol Buffers contracts, database schema, PWA wireframes.
- Working, unit-tested services and both serverless functions; installable PWA.
- Kubernetes deployment through the GitOps pipeline, with observability (Prometheus, Grafana, Jaeger).
- Test plan, test-case documentation and evidence; pilot-user feedback.
- Final presentation and live demo of the full report lifecycle, including escalation.

## 6. Milestones

Delivery runs in three five-day sprints.

| Sprint | Dates (2026) | Milestone |
|---|---|---|
| Sprint 1 | 22 to 27 Oct | Charter, requirements, plans, wireframes, service template, gRPC contracts, k3d cluster, CI, one thin-slice service per owner |
| Sprint 2 | 28 Oct to 2 Nov | Core services implemented with unit tests, databases, PWA front end |
| Sprint 3 | 3 to 8 Nov | Serverless functions, full deployment through Argo CD, full test suite, observability |
| Presentation | around 9 Nov | Final presentation, live demo and closure report |

## 7. Stakeholders and roles

| Stakeholder | Role | Engagement |
|---|---|---|
| Students | Reporters | Usability testing; feedback after the pilot |
| DSA, DAA, SSO | Category Heads (Hostel, Academic Block, Others) | Requirements interviews; reassignment-flow walkthrough |
| Hostel and academic role-holders (councillors, HoDs, estate staff) | Concerned parties reached through reassignment; sign in as staff users | Consulted on the current informal process |
| System Administrator (dev team) | Full technical control | Defines the routing rules |
| College President | Read-only oversight dashboard | Sees the final presentation; no operational role |
| Module Tutors | Sponsors | Weekly progress updates; final presentation |

### Project team

| Member | Project role | Services owned |
|---|---|---|
| Sonam Zangmo | Team Lead | routing-rules, routing-engine, department-directory, category-taxonomy, api-gateway |
| Gayley Choden | Scrum Master | identity, authorization, notification, feedback, analytics; digest-dispatcher |
| Sonam Choki | Developer; owns deployment and CI | report-submission, photo-media, report-history, location-taxonomy, internal-notes |
| Nyendrak Yoezer Zangmo | Developer | status-lifecycle, sla-tracking, sla-policy, reassignment, audit-log; sla-escalation-checker |

Every pull request is peer-reviewed, and nobody merges their own.

## 8. High-level risks

| No. | Risk | Likelihood | Impact | Response |
|---|---|---|---|---|
| R1 | Scope grows beyond the agreed architecture | Medium | High | Avoid: scope locked in this charter; changes need full-team sign-off |
| R2 | Learning curve of new tools (Kubernetes, gRPC) delays the start | High | Medium | Mitigate: dedicated learning spike before full implementation |
| R3 | Local Kubernetes cluster or Knative is unstable near the demo | Medium | High | Mitigate: rehearse deployment early; keep a recorded demo; fall back to Kubernetes CronJobs if Knative fails |
| R4 | Schedule slips once the build is under way | Medium | High | Mitigate: weekly EVM shows drift early; testing and documentation are never cut |
| R5 | Non-teaching staff without institutional email (cooks, cleaners) cannot log in | Low | Medium | Mitigate: their supervisor or HoD holds the account and acts on their behalf, with a full audit trail |

## 9. Constraints and assumptions

- Zero budget: free or open-source tools only, running on the team's own laptops. No paid cloud.
- Student team of four, working alongside other modules.
- Effort baseline of about 360 person-hours, logged in Jira.
- Local, demo-scale deployment for academic assessment.
- Users sign in with institutional email and a password held by Jurwa's identity service; no integration with college systems.

## 10. Success criteria

- 100% correct routing on the test rule-set.
- A student can submit a report with a photo in under two minutes in usability testing.
- SLA escalation works automatically in the live demo.
- At least 70% test coverage on routing, SLA and status-flow code, with tests passing in CI.
- All services deployed to k3d through GitHub Actions and Argo CD in the live demo.
- Project delivered within budget (zero) with weekly EVM reports.

## 11. Sign-off

| Name | Role | Signature | Date |
|---|---|---|---|
| Sonam Zangmo | Team Lead | | |
| Gayley Choden | Scrum Master | | |
| Sonam Choki | Team member | | |
| Nyendrak Yoezer Zangmo | Team member | | |
| Mr. Pema Namgay | Tutor, WEB303 and SWE302 (Sponsor) | | |
| Mr. Nima Yoezer | Tutor, SWE303 (Sponsor) | | |
