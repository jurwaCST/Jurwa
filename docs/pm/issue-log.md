# Issue Log (SCRUM-83)

Project: Jurwa. Created by Sonam Choki, 2 November 2026.

An issue is a problem that has already happened; a risk (in `risk-register.md`) is one that might. Every issue has an owner. Status is Open, In progress or Closed. Changes are made through a pull request.

| ID | Date raised | Description | Owner | Status | Resolution |
|---|---|---|---|---|---|
| I1 | 27 Oct 2026 | The service template (SCRUM-73) and the proto contracts (SCRUM-74) landed about a day late, so every Sprint 1 service slice was squeezed into the last day | Sonam Zangmo | Closed | Both merged on 27 Oct. For Sprint 2, blocking cards go first and are agreed at planning (retro action A1, risk R7) |
| I2 | 27 Oct 2026 | `go mod tidy` removed testify from the template's go.mod, found only during SCRUM-74, so new services could not compile their tests | Sonam Zangmo | In progress | Fixed in the template. `make template-check` builds and tests a fresh copy of the template; retro action A4 adds it to CI |
| I3 | 27 Oct 2026 | CI went red after the template merged: `go test` "matched no packages" because Go skips the `_template` folder | Sonam Choki | Closed | Fixed in PR #16 (SCRUM-101); green run saved in `docs/evidence/sonam-choki/ci-green-run.png` |
| I4 | 28 Oct 2026 | Original estimates and logged hours were entered only at the end of Sprint 1, so the first EVM log was late and CPI and SPI look better than they are | Gayley Choden | In progress | Estimates on day 1 and daily hour logging (retro action A2, risk R8); per-sprint PV to be revised (action A3) |
| I5 | 28 Oct 2026 | Checklists went stale: they no longer matched the real state of the work | Gayley Choden | Open | Reviewers check the Definition of Done in quality-plan.md on every review. |
| I6 | 28 Oct 2026 | Jira still has subtasks and cards from the old 25-service design (for example SCRUM-87 holiday-calendar, SCRUM-95 email-template, SCRUM-96 digest) | Each card owner | Open | Close them as "Won't do" to match change C1 (`schedule-evm.md` section 5); each owner closes only their own cards |
| I7 | 28 Oct 2026 | `make lint` failed on Windows because files had CRLF line endings | Gayley Choden | Open | Add a `.gitattributes` file so line endings are always LF (retro action A6) |
| I8 | 2 Nov 2026 | The internal-notes contract has no role field, but students must be denied. internal-notes reads the caller's role from gRPC metadata `x-user-role`; api-gateway must send it. The service also let System Admin add notes, while `api-routes.md` and the authorization policy allow only Category Head and Staff to add | Sonam Zangmo | In progress | Agreed on 2 Nov by Sonam Zangmo and Gayley Choden: keep the `api-routes.md` rule (add: Category Head, Staff; read: Category Head, Staff, System Admin). internal-notes changed to match (SCRUM-116). Sonam Zangmo adds the notes routes to api-gateway in a Sprint 3 card, setting `author_id` and `x-user-role` from the JWT. No change to `proto/` |
| I9 | 29 Oct 2026 | The first location-taxonomy slice (PR #29) did not follow the team script, so it was reverted (PR #30) and redone test-first, which cost about a day | Sonam Choki | Closed | Redone on branch `SCRUM-23-location-taxonomy` with a separate "tests first" commit |
