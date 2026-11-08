# Accessibility tests (SCRUM-110)

Owner: Gayley Choden. Requirement NFR-06 (the PWA meets WCAG basics, checked with axe-core and Lighthouse) and quality objective Q5. Code: `web/cypress/e2e/a11y.cy.ts`.

## Tools

| Tool | Version | What it checks |
|---|---|---|
| axe-core, through cypress-axe | axe-core 4.14.0, cypress-axe 1.7.0 | Automated WCAG rules on each page, in a real browser (Edge 157) |
| Lighthouse | 13.5.0, in Chrome 156 | The accessibility score of a page |
| Keyboard | by hand | Every control can be reached and used without a mouse |

The axe checks use the rule sets `wcag2a`, `wcag2aa`, `wcag21a` and `wcag21aa` (WCAG 2.0 and 2.1, levels A and AA). A test fails on any violation and prints the rule, its impact and the exact element, so the page's owner knows what to fix.

## 1. axe-core: every page

Run on 8 November 2026 (`docs/evidence/gayley/sprint3/a11y.txt`, video `a11y-run.mp4`):

| Page | Signed in as | Violations |
|---|---|---|
| `/login` | nobody | 0 |
| `/login` showing a wrong password error | nobody | 0 |
| `/` home | nobody | 0 |
| page not found | nobody | 0 |
| `/` home | a student | 0 |
| `/report/new` empty | a student | 0 |
| `/report/new` with a category and building chosen | a student | 0 |
| `/report/new` showing its error messages | a student | 0 |
| 403 (a student opening `/staff`) | a student | 0 |
| `/staff` | Category Head (`dsa.demo`) | 0 |
| `/admin` | System Administrator (`admin.demo`) | 0 |
| `/president` | College President (`president.demo`) | 0 |

**12 of 12 pass, 0 violations.** The pages are signed in through the real login page, so each one is checked as its user sees it.

`/staff`, `/admin` and `/president` are still placeholder pages (a heading and one line: SCRUM-36 and SCRUM-37 were moved to Sprint 4, SCRUM-99 is Stretch). Their result only shows the shared layout passes; the checks must be run again when the real pages are built. The real content checked today is the sign-in page, the home page and the report form. The report form was checked before Nyendrak's styling update (`SCRUM-35-form-polish`, not merged), so it is run again after that merge.

## 2. Lighthouse

| Page | Accessibility score | Scored checks | Evidence |
|---|---|---|---|
| `/login` | **100** | 13 of 13 pass | `docs/evidence/gayley/sprint3/lighthouse-login.report.html` and `.json` |

Target was 90 or more. Lighthouse also lists 10 checks it cannot do automatically (for example tab order and focus); the keyboard pass below covers them.

Lighthouse checks one page without signing in, so only `/login` is scored. The signed-in pages are covered by axe-core (section 1). The run used the dev server (`npx vite`), not a production build.

**The PWA check is not done.** Lighthouse 12 and later no longer has a PWA category, and the PWA manifest and service worker (SCRUM-98, Nyendrak) are not merged. When they are, installability is checked in Chrome DevTools, Application tab, Manifest.

## 3. Keyboard pass (report form)

Done by hand on 8 November 2026 in the browser, signed in as a new student, `/report/new`, no mouse:

| # | Check | Result |
|---|---|---|
| K1 | Tab reaches every field in order: Category, Building, Wing, Area, Description, Photo, Send report | Pass |
| K2 | The selected field is always visible | Pass (blue outline) |
| K3 | Arrow keys change the Category and Building choices | Pass |
| K4 | Enter on "Send report" with empty fields shows the error messages | Pass |
| K5 | The whole form can be filled in and sent, and the confirmation appears | Pass |
| K6 | Nothing needed the mouse | Pass |

## 4. Screen reader

**Not done.** No NVDA or phone screen reader pass was run in Sprint 3.

Found in the code instead: the photo error (for example "larger than 5 MB") is a plain `<p className="error">` in `web/src/pages/ReportForm.tsx`, while the form's other errors have `role="alert"`. A screen reader reads the other errors as they appear, but not the photo error. axe-core does not flag this, because the message is still visible. Sent to the page's owner, Nyendrak (SCRUM-35): add `role="alert"` to that line.

## 5. Open

| Item | Owner |
|---|---|
| `role="alert"` on the photo error message | Nyendrak (SCRUM-35) |
| PWA installability check | Nyendrak (SCRUM-98), then Gayley |
| Screen reader pass (NVDA) | Gayley, Sprint 4 |
| Lighthouse on the signed-in pages | Gayley, Sprint 4 |

## Run it

With the services and the web app running (`docs/testing/e2e-tests.md`, "Run it"):

```bash
cd web
CYPRESS_baseUrl=http://localhost:5173 CYPRESS_apiUrl=http://localhost:18080 CYPRESS_demoPassword=<SEED_DEMO_PASSWORD> npx cypress run --browser edge --spec cypress/e2e/a11y.cy.ts
npx lighthouse@13.5.0 http://localhost:5173/login --only-categories=accessibility --output=json --output=html --output-path=../docs/evidence/gayley/sprint3/lighthouse-login --chrome-flags="--headless=new"
```

The first run after starting the web app can fail once with a blank page: Vite reloads the browser after preparing a page for the first time. Run it again.
