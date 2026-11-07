# End-to-end tests with Cypress

Card: SCRUM-108. Owner: Gayley Choden. Code: `web/cypress/`. Config: `web/cypress.config.ts`.

These tests use Jurwa the way a person does: a real browser on the web app, which talks to the real api-gateway and services. They check that the parts work together, which the unit and integration tests cannot.

## The story

The full story from the guide, and what runs today:

| # | Step | How | Today |
|---|---|---|---|
| 1 | A student signs in | UI: `/login`, types email and password | **Runs** |
| 2 | The student submits a Hostel report with a photo and sees the confirmation | UI: `/report/new` | **Gap**: the form is a placeholder until SCRUM-35 (Nyendrak). Today the report is submitted through the gateway with `cy.request` (no photo) |
| 3 | The report is routed to DSA (Hostel rule) | `cy.request` as the DSA Category Head: list assigned reports | **Gap**: the gateway has no route that lists assigned reports yet |
| 4 | The Category Head sets the status In Progress, then Resolved | `cy.request` to the gateway status route | **Gap**: no status route in the gateway yet (Zangmo's stretch step) |
| 5 | The student sees Resolved on the report timeline | UI | **Gap**: the timeline route is on SCRUM-119 (Zangmo), not merged, and there is no timeline page yet |

When a gap closes, the step is added to `web/cypress/e2e/report-flow.cy.ts` and this table changes to "Runs".

## The tests (`web/cypress/e2e/report-flow.cy.ts`)

Each test registers a new student first (`cy.registerStudent()`), so it never depends on an earlier run or on test order.

| Test | Proves |
|---|---|
| a student signs in through the UI | The login page, the `/api` proxy, the gateway and identity work together, and the header shows the signed-in name |
| a wrong password shows an error and stays on the sign-in page | A 401 from identity reaches the page as a visible error, the user is not let in |
| a student submits a Hostel report and can read it back as OPEN | The gateway takes the reporter id from the token, report-submission stores the report with status OPEN, and it can be read back |
| a description of 501 characters is refused by the server | The 500 character limit (FR-05) holds at the server, not only in the form |
| another student cannot read the report | A student reads only their own reports (the gateway answers 403) |

Still to add once the form exists (guide step 5): the 501 character description refused **in the form**, and a 6 MB photo refused with the size message.

Waits use `cy.intercept(...).as('login')` and `cy.wait('@login')`, never a fixed `cy.wait(2000)`.

## Test data

- `web/cypress/fixtures/users.json`: the demo account emails from the identity seed (SCRUM-126) and the student password. Test accounts only, never real ones. The student is new on every run (`student-<time>@e2e.jurwa.example`).
- `web/cypress/fixtures/photo.png`: a real 8 x 8 PNG of 74 bytes, for the photo step.
- The demo accounts' password is not in the repo: pass it as `CYPRESS_DEMO_PASSWORD` when a test needs a demo account.

## Run it

Three Git Bash windows, from the repo root. Docker is not needed (everything runs in memory).

1. The services (gateway, identity, authorization, category-taxonomy, location-taxonomy, report-submission, internal-notes):

   ```bash
   export SEED_DEMO_PASSWORD='<any 8+ characters>'
   bash web/cypress/start-stack.sh
   ```

   Wait for `Ready: gateway on http://localhost:8080`. A port is taken? Move it, for example `GATEWAY_HTTP_PORT=18080 AUTHORIZATION_PORT=19017 CATEGORY_PORT=19009 bash web/cypress/start-stack.sh`.
2. The web app (with `VITE_API_URL=/api` in `web/.env.local`):

   ```bash
   cd web && npm run dev
   ```

   With another gateway port: `GATEWAY_URL=http://localhost:18080 npm run dev`.
3. The tests:

   ```bash
   cd web && npx cypress run
   ```

   With other ports: `CYPRESS_baseUrl=http://localhost:5174 CYPRESS_apiUrl=http://localhost:18080 npx cypress run`. To watch them in a window: `npm run e2e:open`.

Cypress writes a video to `web/cypress/videos/` and, when a test fails, a screenshot to `web/cypress/screenshots/` (both ignored by git).

**Windows tips.**

- In the VS Code terminal Cypress may stop with `bad option: --smoke-test`. VS Code sets `ELECTRON_RUN_AS_NODE=1`; run `unset ELECTRON_RUN_AS_NODE` first, or use a normal Git Bash window.
- If the built-in Electron browser crashes, use another one: `npx cypress run --browser edge` (or `chrome`).

## Not in CI

CI would need the whole stack (gateway, six services, the web dev server) in the runner, which is too heavy for this sprint. The run is shown locally instead: `docs/evidence/gayley/sprint3/cypress-run.txt` and the video.

## Known gaps

- Steps 2 to 5 of the story wait for the report form (SCRUM-35), the photo route, a route listing assigned reports, the status route and the timeline (SCRUM-119). See the table above.
- Cypress warns that `Cypress.env()` will be replaced by `cy.env()` and `Cypress.expose()` in a future version. It still works in 15.21; change it when upgrading.
