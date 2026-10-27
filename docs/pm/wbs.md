# Work Breakdown Structure (SCRUM-44)

Project: Jurwa. Owner: Gayley Choden (Scrum Master). Version 1.0, 27 October 2026.

The eight work packages and their hours are copied from Nyendrak's baseline (`docs/pm/schedule-evm.md`, section 1), which takes them unchanged from the proposal (section 4D). Hours are person-hours. Hours are planned per work package only; tasks inside a package have no separate hour estimates in the baseline, so none are given here. Task-level estimates live on the Jira cards.

## 1. WBS

| WP | Work package | Tasks | Jira cards (Sprint 1) | Planned hours |
|---|---|---|---|---|
| 1.1 | Initiation | 1.1.1 Project charter; 1.1.2 Stakeholder register; 1.1.3 Requirements gathering | SCRUM-77 charter, SCRUM-44 stakeholder register, SCRUM-78 requirements | 20 |
| 1.2 | Planning | 1.2.1 WBS and schedule; 1.2.2 Risk register; 1.2.3 Quality management plan; 1.2.4 Communications plan | SCRUM-79 schedule and EVM, SCRUM-44 WBS and risk register, SCRUM-80 quality plan, SCRUM-81 communications plan | 25 |
| 1.3 | System design | 1.3.1 Learning spike (Kubernetes, gRPC, Knative); 1.3.2 Bounded contexts; 1.3.3 gRPC / .proto contracts; 1.3.4 Database schema; 1.3.5 UI/UX wireframes | SCRUM-71 learning spike, SCRUM-72 bounded contexts, SCRUM-74 contracts, SCRUM-75 database design, SCRUM-76 wireframes | 35 |
| 1.4 | Implementation | 1.4.1 Shared service template; 1.4.2 Service domains (Identity and Access, Reporting, Routing and Configuration, Lifecycle and SLA, Comms and Engagement, Platform); 1.4.3 Two serverless functions; 1.4.4 PWA front end | SCRUM-73 template; thin slices SCRUM-17, 18, 19, 13 | 160 |
| 1.5 | Testing | 1.5.1 Unit; 1.5.2 Integration; 1.5.3 End-to-end; 1.5.4 Performance; 1.5.5 Accessibility | From Sprint 2 (see `quality-plan.md`) | 70 |
| 1.6 | Deployment | 1.6.1 Dockerfiles and images; 1.6.2 Kubernetes manifests (k3d); 1.6.3 CI (GitHub Actions); 1.6.4 CD (Argo CD); 1.6.5 Observability | SCRUM-100 k3d and manifests, SCRUM-101 CI | 30 |
| 1.7 | Monitoring and control | 1.7.1 Status reports and EVM; 1.7.2 Issue log and risk review | SCRUM-82 first EVM log | 10 |
| 1.8 | Closure | 1.8.1 Final presentation; 1.8.2 Lessons learned; 1.8.3 Archive handover | Sprint 3 and presentation (about 9 Nov) | 10 |
| | **Total (BAC)** | | | **360** |

Check: 20 + 25 + 35 + 160 + 70 + 30 + 10 + 10 = 360, the same total as `schedule-evm.md`.

## 2. Hierarchy

```
1 Jurwa
├── 1.1 Initiation (20 h)
├── 1.2 Planning (25 h)
├── 1.3 System design (35 h)
├── 1.4 Implementation (160 h)
├── 1.5 Testing (70 h)
├── 1.6 Deployment (30 h)
├── 1.7 Monitoring and control (10 h)
└── 1.8 Closure (10 h)
```

## 3. Notes

- Scope follows the charter: 20 services and 2 serverless functions (change C1 in `schedule-evm.md`). The hours were not changed when services were cut.
- Whether the 80 h of WP 1.1 to 1.3 count as already spent (tracking BAC 280 h) or as planned inside Sprint 1 (BAC 360 h) is open question 1 in `schedule-evm.md`, to be decided by the Scrum Master and Team Lead at the Sprint 1 review.
- If the baseline in `schedule-evm.md` changes, this file is updated in the same sprint.
