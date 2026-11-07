# Property-based tests: SLA time windows

Card: SCRUM-107. Owner: Nyendrak Yoezer Zangmo. Tool: gopter (`github.com/leanovate/gopter` v0.2.11). Test file: `tests/property/sla_window_test.go`.

A property-based test states a rule that must hold for every input, then checks it against many random inputs. Here the inputs are start times anywhere in 2026 and 2027 (to the second), the three categories (HOSTEL, ACADEMIC_BLOCK, OTHERS), SLA windows from 1 hour to 30 days, and "now" from 3 days before to 3 days after the deadline. Every property runs on at least 100 random cases (`MinSuccessfulTests = 100`).

The tests use the real domain code of sla-policy (`NewPolicy`, `Policy.Window`) and sla-tracking (`NewTimer`, `IsDue`, `Lapse`, `Stop`). No services, network or database.

## Results

Run on 7 November 2026 (`docs/evidence/nyendrak/sprint3/gopter.txt`):

| # | Property | Why it matters | Result |
|---|---|---|---|
| 1 | The deadline is never before the start time | A report cannot be late before it is reported | Passed, 100 cases |
| 2 | Deadline minus start equals the sla-policy window for the category | The SLA set by the System Administrator is the SLA used | Passed, 100 cases |
| 3 | A timer is due exactly when now is at or after the deadline, never before | No report is escalated early, none is missed | Passed, 100 cases |
| 3b | `Lapse` succeeds exactly when the timer is due, and only once | A report is escalated once per lapse (the Saga relies on it) | Passed, 100 cases |
| 4 | A longer window never gives an earlier deadline | Raising a window can only give more time | Passed, 100 cases |
| 5 | Starting later never gives an earlier deadline | A report routed later is not due sooner | Passed, 100 cases |
| 6 | A stopped timer is never due, however late it is | A resolved report is never escalated | Passed, 100 cases |
| 7 | sla-policy accepts every window of 1 hour or more (up to 2^30 hours) | No valid window is refused | Passed, 100 cases |
| 8 | sla-policy refuses every window of 0 hours or less | A zero or negative window cannot be saved | Passed, 100 cases |

## Counter-examples

gopter found no counter-example, so the domain code was not changed.

## Holidays and weekends

Not tested, because the domain does not support them. The holiday-calendar service was one of the five services removed when the team moved from 25 to 20 services, so SLA windows count calendar hours, weekends and holidays included. A property such as "the deadline never falls on a holiday" would test a rule the system does not have.

## Run

```bash
go test ./tests/property/... -v
```
