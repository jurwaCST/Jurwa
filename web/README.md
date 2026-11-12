# web (Jurwa PWA)

Card: SCRUM-118. Setup owner: Gayley Choden.

React 19 and TypeScript, built with Vite. The app talks only to the api-gateway (`docs/design/api-routes.md`).

## Run

```bash
cd web
npm install        # first time, and after every git pull that changes package-lock.json
npm run dev        # http://localhost:5173
npm test           # Vitest, once
npm run build      # type check and production build into dist/
npm run lint       # oxlint
```

Start the gateway and the services it calls first (identity, authorization, category-taxonomy, location-taxonomy, report-submission). The gateway needs `JWT_SECRET` set for identity.

### Gateway address

| Setting | Where | Default |
|---|---|---|
| `VITE_API_URL` | `web/.env.local` (copy `.env.example`) | `http://localhost:8080` |
| `GATEWAY_URL` | environment, for the dev proxy only | `http://localhost:8080` |

The gateway sends no CORS headers yet, so a browser on port 5173 cannot call port 8080 directly. Use the dev proxy: put `VITE_API_URL=/api` in `web/.env.local`, and the dev server forwards `/api/...` to the gateway. Put it in the file, not on the command line: Git Bash turns `VITE_API_URL=/api` into `C:/Program Files/Git/api`, and every call then fails with "Cannot reach the server". If port 8080 is taken on your laptop, run the gateway with `HTTP_PORT=8081` and start the app with `GATEWAY_URL=http://localhost:8081 npm run dev`.

There are no seeded accounts yet. Register one with `POST /auth/register` (a new account is a Student):

```bash
curl -X POST http://localhost:5173/api/auth/register -H "Content-Type: application/json" \
  -d '{"email":"pema.test@cst.edu.bt","fullName":"Pema Test","password":"pema-pass-123"}'
```

## Container image

`Dockerfile` (SCRUM-140) builds the app with `VITE_API_URL=/api` and serves it with nginx on port **8080** as a non-root user (uid 101). `nginx.conf` sends every path that is not a file to `index.html`, so a reload on `/login` or `/report/new` still works, and caches the hashed files in `/assets/` for a year. The Kubernetes Deployment and the Traefik `/` route are SCRUM-139 (Sonam Choki).

```bash
docker build -f web/Dockerfile -t jurwa-web web     # from the repo root; the context is web/
docker run --rm -p 8088:8080 jurwa-web              # http://localhost:8088
```

## Install as an app

The app is a PWA (SCRUM-98, NFR-07): `public/manifest.webmanifest` (name, pine theme colour, icons made from `cst-logo.jpg`) and a hand-written service worker in `public/sw.js`, registered by `src/pwa/register.ts` in the built app only (never in `npm run dev`).

- **Install:** on a phone, open the HTTPS address and use "Add to home screen" (Android Chrome also offers "Install app"). On a laptop, Chrome shows an install icon in the address bar.
- **Offline:** pages open from the saved app shell, and a banner says "You are offline. Showing saved pages." (`src/pwa/OfflineBanner.tsx`). API calls (`/api/...`) are never cached, so reports and statuses are never stale; without a connection they fail as usual.
- **Updating:** change `CACHE_VERSION` in `sw.js` when the list of shell files changes; older caches are deleted when the new worker starts. nginx never caches `sw.js`, so browsers see a new version on the next visit.
- **Check:** `npm run build && npm run preview`, then Chrome DevTools, Application tab: Manifest (no errors, "Installable") and Service workers (activated and running).

## Folders

| Path | Content | Owner |
|---|---|---|
| `src/api/client.ts` | The API client (below) | Gayley |
| `src/auth/` | `AuthProvider` (login, logout, session), `useAuth`, `RequireRole`, the `User` and `Role` types | Gayley |
| `src/components/` | Shared pieces (below) and the `Layout` frame | Gayley; ask before changing one |
| `src/pages/` | One file per page, so nobody edits the same file | see below |
| `src/test/setup.ts` | Test setup (jest-dom matchers) | Gayley |
| `src/index.css` | Design tokens: every colour, space, radius, shadow and type size | Gayley |

## Design and shared components

Colours come from the CST campus: forest green (`--pine-*`) for the bar and text, roof teal (`--teal-*`) for links, saffron walls (`--saffron-*`, `--action`) for the **one** main button of a screen, pillar red for Escalated. Use the variables in `src/index.css`, never a hex value, so light and dark mode both work. The report form's names (`--bg`, `--text`, `--muted`, `--border`, `--accent`, `--accent-text`, `--danger`) stay. Spacing is an 8 px grid: `--s1` (8 px) to `--s8` (64 px).

Import from `src/components/`. Every piece is tested in `components.test.tsx`.

| Piece | Use it for | Example |
|---|---|---|
| `PageHeader` | The title block of a page | `<PageHeader title="My queue" description="Reports assigned to you" aside={<span className="tag">READ-ONLY</span>} />` |
| `Field` | Label, hint and error around one `input` or `select` (the label wraps it, so Cypress `findLabel` still works) | `<Field label="Building" hint="From the list" error={err}><select>…</select></Field>` |
| `TextArea` | Multi-line text with a live `12 / 500` counter (counts Dzongkha letters as characters) | `<TextArea label="Description" max={500} value={d} onChange={setD} />` |
| `Button` | `variant`: `primary` (saffron, one per screen), `secondary`, `ghost`, `danger`; `loading` disables it and shows `loadingText` | `<Button type="submit" loading={busy} loadingText="Sending...">Send report</Button>` |
| `StatusBadge` | A report status: colour, icon and word together; Escalated is filled red | `<StatusBadge status="IN_PROGRESS" />` |
| `Timeline` | A report's history, oldest first | `<Timeline entries={[{ status: 'OPEN', at: '2026-11-08T08:00:00Z' }, { status: 'ASSIGNED', at: '…', by: 'DSA' }]} />` |
| `BottomSheet` | "Update status", "Reassign": slides up on a phone, a dialog on a laptop; Escape closes it | `<BottomSheet open={open} title="Update status" onClose={() => setOpen(false)}>…</BottomSheet>` |
| `Toast` | A short confirmation after an action; remove it in `onDone` | `{msg && <Toast message="Status updated" onDone={() => setMsg(null)} />}` |
| `Skeleton` | Grey lines while a list loads | `<Skeleton lines={4} label="Loading reports" />` |
| `ErrorState` | A list or page that could not load, with "Try again" | `<ErrorState message={err.message} onRetry={load} />` |
| `EmptyState` | Nothing to show yet | `<EmptyState title="Nothing in your queue">New reports appear here.</EmptyState>` |
| `StatTile` | One big number on a dashboard | `<StatTile label="Escalated" value="3" icon="escalated" tone="escalated" />` |
| `Icon` | The app's icons (`report`, `queue`, `open`, `resolved`, …); decorative unless given a `label` | `<Icon name="queue" />` |

CSS classes for plain HTML: `card` (a white panel), `chips` with `<button className="chip" aria-pressed={on}>` (filter chips), `tag` (for READ-ONLY), `action` (a saffron link that looks like the main button), and `stack-table` for a table that turns into cards on a phone (put the column name on each cell: `<td data-label="Status">`).

Phones get a bottom tab bar and laptops a top bar; the tabs come from `src/components/Layout.tsx`. When a new page is added to `App.tsx`, add its tab there too.

## Pages and routes

| Route | Page file | Roles | Owner |
|---|---|---|---|
| `/` | `Home.tsx` | everyone (links by role) | Gayley |
| `/login` | `Login.tsx` | everyone | Gayley |
| `/report/new` | `ReportForm.tsx` | Student | Nyendrak, SCRUM-35 |
| `/staff` | `StaffView.tsx` | Category Head, Staff | Nyendrak, SCRUM-36 |
| `/admin` | `AdminDashboard.tsx` | System Administrator | Nyendrak, SCRUM-37 |
| `/president` | `PresidentDashboard.tsx` | College President | Gayley, SCRUM-99 |
| anything else | `NotFound.tsx` | everyone | Gayley |

The pages for SCRUM-35, 36, 37 and 99 are placeholders: replace the content of the file, keep the file name and the default export. The routes and roles are in `src/App.tsx`. `RequireRole` sends a visitor who is not signed in to `/login` and shows `Forbidden.tsx` (403) to a wrong role. It only hides pages: the gateway checks the role again on every call.

## API client

```ts
import { api, ApiError } from '../api/client'

const cats = await api.get<{ categories: Category[] }>('/categories')
const report = await api.post<SubmitReportResponse>('/reports', { category: 'HOSTEL', ... })
await api.put('/admin/sla-policies/HOSTEL', { ... })
const photo = await api.upload<UploadPhotoResponse>('/photos', file, 'photo')
```

| Function | Does |
|---|---|
| `api.get<T>(path)` | GET, returns the JSON body |
| `api.post<T>(path, body)` | POST with a JSON body |
| `api.put<T>(path, body)` | PUT with a JSON body |
| `api.upload<T>(path, file, field)` | POST one file as `multipart/form-data` under `field` |
| `setToken(token \| null)` | Sets the JWT sent as `Authorization: Bearer <token>`. `AuthProvider` calls it, pages do not |

Every failure throws `ApiError` with `status` (the HTTP code, or `0` when the server cannot be reached) and `message` (the gateway's `{"error": "..."}` text). A 204 answer returns `undefined`. JSON field names are lowerCamelCase and enums are their names (`"HOSTEL"`), as the gateway sends them.

Use the signed-in user in a page with `const { user, logout } = useAuth()` (`user.roles` is a list of `Role`).

## Login session

`POST /auth/login` returns the JWT, its expiry and the user. The app keeps them in memory and in `sessionStorage` (cleared when the tab closes, never the password). An expired session counts as signed out. Storage errors (private mode) are caught, so login still works for the open tab.

## Tests

Vitest with Testing Library, in jsdom. Test files sit next to the code (`*.test.ts`, `*.test.tsx`). `fetch` is replaced with `vi.fn()` in tests, so no gateway is needed. CI runs `npm ci`, `npm test` and `npm run build` in `.github/workflows/web.yml` on every pull request that changes `web/`.
