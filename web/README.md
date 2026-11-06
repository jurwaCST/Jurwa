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

The gateway sends no CORS headers yet, so a browser on port 5173 cannot call port 8080 directly. Use the dev proxy: put `VITE_API_URL=/api` in `web/.env.local`, and the dev server forwards `/api/...` to the gateway. If port 8080 is taken on your laptop, run the gateway with `HTTP_PORT=8081` and start the app with `GATEWAY_URL=http://localhost:8081 npm run dev`.

There are no seeded accounts yet. Register one with `POST /auth/register` (a new account is a Student):

```bash
curl -X POST http://localhost:5173/api/auth/register -H "Content-Type: application/json" \
  -d '{"email":"pema.test@cst.edu.bt","fullName":"Pema Test","password":"pema-pass-123"}'
```

## Folders

| Path | Content | Owner |
|---|---|---|
| `src/api/client.ts` | The API client (below) | Gayley |
| `src/auth/` | `AuthProvider` (login, logout, session), `useAuth`, `RequireRole`, the `User` and `Role` types | Gayley |
| `src/components/` | Shared pieces, for example `Layout` (header with the signed-in user) | whoever adds one |
| `src/pages/` | One file per page, so nobody edits the same file | see below |
| `src/test/setup.ts` | Test setup (jest-dom matchers) | Gayley |

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
