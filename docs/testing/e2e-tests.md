# End-to-end tests with Cypress

Card: SCRUM-108. Owner: Gayley Choden. Code: `web/cypress/`. Config: `web/cypress.config.ts`.

These tests use Jurwa the way a person does: a real browser on the web app, which talks to the real api-gateway and services. They check that the parts work together, which the unit and integration tests cannot.

## The story

The full story from the guide, and what runs today:

| # | Step | How | Today |
|---|---|---|---|
| 1 | A student signs in | UI: `/login`, types email and password | **Runs** |
| 2 | The student submits a Hostel report with a photo and sees the confirmation | UI: `/report/new` | **Runs** (SCRUM-152): the form is filled and sent in `student-pages.cy.ts` (category tile, the first real place, a description) and shows "Report sent" with the number. The photo is still attached with `cy.request` in `report-flow.cy.ts` |
| 3 | The report is routed to DSA (Hostel rule) | `cy.request`: the report's status | **Runs** (SCRUM-152): routing-engine sends `report.routed` and the status becomes Assigned; the test waits for it (needs Redis) |
| 4 | The Category Head sets the status In Progress, then Resolved | `cy.request` as `dsa.demo` to `POST /reports/{id}/status` | **Runs** (SCRUM-152), in the feedback test |
| 5 | The student sees Resolved on the report timeline | UI: `/reports/{id}` | **Runs** (SCRUM-152): the timeline shows Open then Assigned, and Resolved in the feedback test |

When a gap closes, the step is added to `web/cypress/e2e/report-flow.cy.ts` and this table changes to "Runs".

## The tests (`web/cypress/e2e/report-flow.cy.ts`)

Each test registers a new student first (`cy.registerStudent()`), so it never depends on an earlier run or on test order.

| Test | Proves |
|---|---|
| a student signs in through the UI | The login page, the `/api` proxy, the gateway and identity work together, and the header shows the signed-in name |
| a wrong password shows an error and stays on the sign-in page | A 401 from identity reaches the page as a visible error, the user is not let in |
| a student submits a Hostel report and can read it back as OPEN | The gateway takes the reporter id from the token, report-submission stores the report with status OPEN, and it can be read back |
| a description of 501 characters is refused by the server | The 500 character limit (FR-05) holds at the server, not only in the form |
| a student uploads a photo, attaches it to a report and can view it | `POST /photos` returns a photo id, the report keeps it, and `GET /photos/{id}` sends the image as `image/png` with `nosniff` |
| a photo over 5 MB is refused with the size message | A 5.5 MB PNG gets 400 "photo is larger than 5 MB" from photo-media (FR-05). 5.5 MB stays under the gateway's 6 MB body limit, so the answer comes from the 5 MB rule itself |
| another student cannot view the photo | A student sees only their own photos (403) |
| another student cannot read the report | A student reads only their own reports (the gateway answers 403) |

Photos are uploaded with the browser's own `fetch` and `FormData` through the `/api` proxy (`cy.uploadPhoto`), so the request is the same multipart body the form will send.

Still to add (guide step 5): the 501 character description and the over 5 MB photo refused **in the form**, with the message on the page. The form exists now (SCRUM-35); these two are not written yet.

## The student pages (`web/cypress/e2e/student-pages.cy.ts`, SCRUM-152)

Card: SCRUM-152. Owner: Nyendrak. The pages are Gayley's: My reports and the report page (SCRUM-143), notifications and the bell (SCRUM-144), feedback stars (SCRUM-145), with the gateway routes from SCRUM-147. Each test registers and signs in a new student. Helpers that set up a report and move it along are in `web/cypress/support/student.ts`.

| ID | Test | Steps | Expected result | Result (18 Nov 2026) |
|---|---|---|---|---|
| E2E-SP-01 | a report sent in the form shows in My reports and its page shows Open, then Assigned | Sign in; Report tab; choose the Hostel tile, the first building, floor and area; type a description; Send report; wait until the status is Assigned; My reports tab; open the card | "Report sent" with the report number; My reports has one card with the badge Assigned; the report page shows the description, and "What happened" shows Open, then Assigned | Pass |
| E2E-SP-02 | the bell opens Notifications, which shows the list or the empty state | Sign in as a new student (no reports yet); click the bell | The address is `/notifications`, the heading is Notifications, `GET /me/notifications` answers 200, and the page shows the list or "No notifications yet", never an error | Pass |
| E2E-SP-03 | on a resolved report the student gives four stars and sees the thank-you | Send a Hostel report (`cy.request`); as `dsa.demo`, wait for Assigned, then set In Progress and Resolved; open the report page; click the 4th star; type a comment; Send feedback | "What happened" shows Resolved; the choice reads "4 stars out of 5"; `POST /reports/{id}/feedback` answers 200 or 201; the page says "You rated this 4 out of 5 stars." | Pass |
| E2E-SP-04 | after a report is sent, the bell shows new and Notifications lists "We received your report" (added 19 Nov, after SCRUM-155) | Send a Hostel report (`cy.request`); wait until `GET /me/notifications` has the message; open Home; click the bell | Before the click the bell is "Notifications, new" (the dot); the page lists "We received your report" with "You can follow it in My reports."; on the page the dot is gone | Pass (19 Nov 2026) |

Evidence: `docs/evidence/nyendrak/sprint4/scrum-152-tests-first-fail.txt` (the tests run first against main from before the pages, commit `adbcb2a`: 3 of 3 fail because the pages are missing, while the form, routing and status changes already work) and `scrum-152-tests.txt` (the whole suite after: **26 of 26 pass**, report flow 8, accessibility 15, student pages 3). After the notification consumer was merged (SCRUM-155, PR #134), E2E-SP-04 was added: run first against main from before it (commit `a848418`) it fails with no message (`scrum-152-notifications-tests-first-fail.txt`), and after it the whole suite passes again (`scrum-152-notifications-tests.txt`). Screenshots of each page at phone (390 px) and laptop (1366 px) width, light and dark: `scrum-152-<page>-<size>-<mode>.png`.

The routing step happens over the event bus a moment after the report is sent, so `waitForStatus` asks for the status up to 20 times, half a second apart, instead of one fixed wait.

Waits use `cy.intercept(...).as('login')` and `cy.wait('@login')`, never a fixed `cy.wait(2000)`.

## Test data

- `web/cypress/fixtures/users.json`: the demo account emails from the identity seed (SCRUM-126) and the student password. Test accounts only, never real ones. The student is new on every run (`student-<time>@e2e.jurwa.example`).
- `web/cypress/fixtures/photo.png`: a real 8 x 8 PNG of 74 bytes, for the photo step.
- The demo accounts' password is not in the repo: pass it as `CYPRESS_demoPassword` (this exact case) when a test needs a demo account.

## Run it

Three Git Bash windows, from the repo root. Docker is not needed (everything runs in memory).

1. The services (gateway, identity, authorization, category-taxonomy, location-taxonomy, report-submission, internal-notes, photo-media, report-history):

   ```bash
   export SEED_DEMO_PASSWORD='<any 8+ characters>'
   bash web/cypress/start-stack.sh
   ```

   Wait for `Ready: gateway on http://localhost:8080`. A port is taken? Move it, for example `GATEWAY_HTTP_PORT=18080 GATEWAY_GRPC_PORT=19010 AUTHORIZATION_PORT=19017 CATEGORY_PORT=19009 bash web/cypress/start-stack.sh`. Stop it with Ctrl+C in that window: closing the window or killing bash leaves the services running. To find one still holding a port: `netstat -ano | findstr :9010`, then `taskkill /F /PID <pid>`.
2. The web app, **built** (with `VITE_API_URL=/api` in `web/.env.local`):

   ```bash
   cd web && npm run build && npx vite preview --port 4173 --strictPort
   ```

   With another gateway port: `GATEWAY_URL=http://localhost:18080 npx vite preview --port 4173 --strictPort`, and run the tests with `CYPRESS_baseUrl=http://localhost:4173`. Test the built app, not `npm run dev`: under the load of Cypress, Edge and nine services, the dev server sometimes still showed a blank page when a test looked, so tests failed at random (9 November 2026). The built app is also what is deployed.
3. The tests:

   ```bash
   cd web && npx cypress run
   ```

   With other ports: `CYPRESS_baseUrl=http://localhost:5174 CYPRESS_apiUrl=http://localhost:18080 npx cypress run`. To watch them in a window: `npm run e2e:open`.

Cypress writes a video to `web/cypress/videos/` and, when a test fails, a screenshot to `web/cypress/screenshots/` (both ignored by git).

**Windows tips.**

- In the VS Code terminal Cypress may stop with `bad option: --smoke-test`. VS Code sets `ELECTRON_RUN_AS_NODE=1`; run `unset ELECTRON_RUN_AS_NODE` first, or use a normal Git Bash window.
- If the built-in Electron browser crashes, use another one: `npx cypress run --browser edge` (or `chrome`).
- Why the config is `cypress.config.mjs` (plain JavaScript) and the tests are compiled with esbuild: with TypeScript 7 and Node 26, Cypress hung after the `allowCypressEnv` warning and printed nothing else. Its default compiler (webpack + ts-loader) needs the TypeScript JS API, which TypeScript 7 no longer has, and its helper process crashed (exit code `0xC0000409`). Its loader for a `.ts` config (tsx) also hung on Node 26. The tests themselves stay TypeScript.

## Not in CI

CI would need the whole stack (gateway, eight services, the web dev server) in the runner, which is too heavy for this sprint. The run is shown locally instead: `docs/evidence/gayley/sprint3/cypress-run.txt` (8 of 8 passing, 7 Nov 2026) and `cypress-run.mp4`.

## Known gaps

- The student pages tests need Redis (routing, status, notifications); `start-stack.sh` starts every service they use, including `notification` (SCRUM-155).
- (Closed by SCRUM-155) Notifications were not sent to students, so the bell only showed the empty state. The consumer is merged and E2E-SP-04 checks a real message.
- Feedback (SCRUM-145) is clicked through in E2E-SP-03 but not yet checked by axe.
- (Closed by SCRUM-152) Steps 2 to 5 of the story waited for the report form (SCRUM-35), a route listing assigned reports (`ListReportsByAssignee`, Choki, then `GET /me/assigned-reports`, Zangmo), the status routes (Zangmo) and report-history reading events (for the timeline). See the table above.
- Cypress warns that `Cypress.env()` will be replaced by `cy.env()` and `Cypress.expose()` in a future version. It still works in 15.21; change it when upgrading.
