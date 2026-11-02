# BDD scenarios (SCRUM-94)

Owner: Sonam Choki. Behaviour-driven development (BDD) scenarios for the report lifecycle, written in Gherkin (SWE302).

## How to read the feature file

`features/report-lifecycle.feature` describes the behaviour in plain language that a tutor or a Category Head can check:

- **Feature**: the whole behaviour (the report lifecycle).
- **Rule**: one business rule; the scenarios under it are examples of that rule.
- **Scenario**: one example. **Given** is the starting state, **When** is the action, **Then** is the expected result. **And** continues the step before it.
- **Scenario Outline** with **Examples**: the same scenario run once per table row.
- **Background**: steps that come before every scenario.

Status names and rules are the ones agreed for the project: statuses `OPEN`, `ASSIGNED`, `IN_PROGRESS`, `RESOLVED`, `ESCALATED`; description at most 500 characters; photo JPEG or PNG, at most 5 MB; Hostel to DSA, Academic Block to DAA, Others to SSO.

## Which unit tests check each scenario

| Scenario | Checked today by |
|---|---|
| A student submits a valid report | `services/report-submission/domain/report_test.go` `TestNewReportFillsFields` |
| A report without a category is rejected | `services/report-submission/domain/report_test.go` `TestNewReportValidation` |
| A description of 501 / exactly 500 characters | `services/report-submission/domain/report_test.go` `TestNewReportValidation` |
| A photo over 5 MB / of exactly 5 MB | `services/photo-media/domain/photo_test.go` `TestNewPhotoSizeBoundaries` |
| A report is routed to the Category Head of its category | `services/routing-rules/domain/rule_test.go` `TestMatchDefaultRules` |
| A category head reassigns a report | reassignment service (Nyendrak), Sprint 2 |
| A report moves from Open to Resolved | `services/status-lifecycle/domain/transition_test.go` `TestCheckTransition_AllPairs`; timeline order in `services/report-history/application/service_test.go` `TestTimelineIsOldestFirstEvenWhenRecordedOutOfOrder` |
| A resolved report cannot go back to Open; a step cannot be skipped | `services/status-lifecycle/domain/transition_test.go` `TestCheckTransition_AllPairs` and `TestReportStatus_ChangeToInvalidLeavesItUnchanged` |
| A report whose SLA passes becomes Escalated; a resolved report is never escalated | `services/status-lifecycle/domain/transition_test.go` `TestCheckTransition_AllPairs` (system escalation); the timer is `sla-escalation-checker` (Nyendrak) |
| A student cannot read internal notes | `services/internal-notes/domain/note_test.go` `TestWhoMayUseNotes` |

## Step definitions

The scenarios are not executed yet. In Sprint 3 the `godog` library (Cucumber for Go) will run them: each Given, When and Then line is matched to a Go function that calls the domain packages, starting with the form and routing scenarios.
