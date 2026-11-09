// Admin dashboard (SCRUM-37). Owner: Nyendrak.
// The System Administrator sees how many reports are in each status, per
// category, and what needs attention (escalated, not assigned yet), from
// GET /dashboard/summary and GET /dashboard/categories (analytics).
// Editing routing rules and SLA hours comes when the gateway has the
// /admin/routing-rules and /admin/sla-policies routes.
// Built from Gayley's design: PageHeader, StatTile, Skeleton, ErrorState, Icon,
// the chip, card and stack-table classes and the tokens in index.css.
import { useEffect, useState } from 'react'
import { api, ApiError } from '../api/client'
import ErrorState from '../components/ErrorState'
import Icon from '../components/Icon'
import PageHeader from '../components/PageHeader'
import Skeleton from '../components/Skeleton'
import StatTile from '../components/StatTile'
import './AdminDashboard.css'

// int64 counts arrive as strings (protojson); a count of 0 may be left out.
interface Counts {
  total?: string
  open?: string
  assigned?: string
  inProgress?: string
  resolved?: string
  escalated?: string
}
interface Summary {
  counts?: Counts
  averageResolutionHours?: number
}
interface Breakdown {
  categories?: { category: string; counts?: Counts }[]
}

const n = (v?: string) => Number(v ?? 0)
const reports = (count: number) => (count === 1 ? '1 report' : `${count} reports`)

const CATEGORY_NAMES: Record<string, string> = { HOSTEL: 'Hostel', ACADEMIC_BLOCK: 'Academic Block', OTHERS: 'Others' }

type Period = 'week' | 'month' | 'all'
const PERIODS: { id: Period; label: string }[] = [
  { id: 'week', label: 'This week' },
  { id: 'month', label: 'This month' },
  { id: 'all', label: 'All time' },
]

// The query for a period, in UTC ("Z") so it is safe in a URL as it is.
function query(p: Period): string {
  if (p === 'all') return ''
  const now = new Date()
  const from = p === 'week' ? new Date(now.getTime() - 7 * 24 * 3600 * 1000) : new Date(now.getFullYear(), now.getMonth(), 1)
  return `?from=${from.toISOString()}`
}

function messageFor(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 401) return 'Your sign-in has expired. Please sign in again.'
    if (err.status === 0 || err.status === 502 || err.status === 503 || err.status === 504) {
      return 'The service is busy or unreachable. Please try again.'
    }
    return err.message
  }
  return 'The dashboard could not load.'
}

type Result = { key: string; summary: Summary; breakdown: Breakdown } | { key: string; error: string }

export default function AdminDashboard() {
  const [period, setPeriod] = useState<Period>('all')
  const [attempt, setAttempt] = useState(0)
  const [result, setResult] = useState<Result | null>(null)
  // A result for another period or attempt counts as still loading.
  const key = `${period}-${attempt}`

  useEffect(() => {
    let alive = true
    const q = query(period)
    Promise.all([api.get<Summary>(`/dashboard/summary${q}`), api.get<Breakdown>(`/dashboard/categories${q}`)])
      .then(([summary, breakdown]) => alive && setResult({ key, summary, breakdown }))
      .catch((err) => alive && setResult({ key, error: messageFor(err) }))
    return () => {
      alive = false
    }
  }, [period, key])

  const loaded = result?.key === key ? result : null
  const error = loaded && 'error' in loaded ? loaded.error : null
  const data = loaded && 'summary' in loaded ? loaded : null
  const c = data?.summary.counts ?? {}
  const escalated = n(c.escalated)
  const notAssigned = n(c.open)

  return (
    <>
      <PageHeader title="Admin dashboard" description="Reports across campus, and what needs your attention." />

      <ul className="chips admin-periods" aria-label="Period">
        {PERIODS.map((p) => (
          <li key={p.id}>
            <button type="button" className="chip" aria-pressed={period === p.id} onClick={() => setPeriod(p.id)}>
              {p.label}
            </button>
          </li>
        ))}
      </ul>

      {error ? (
        <ErrorState message={error} onRetry={() => setAttempt((a) => a + 1)} />
      ) : !data ? (
        <Skeleton lines={6} label="Loading the dashboard" />
      ) : (
        <div className="admin-grid">
          <section className="admin-tiles" aria-label="Report counts">
            <StatTile label="Total" value={String(n(c.total))} icon="dashboard" tone="open" />
            <StatTile label="Open" value={String(n(c.open))} icon="open" tone="open" />
            <StatTile label="Assigned" value={String(n(c.assigned))} icon="assigned" tone="assigned" />
            <StatTile label="In progress" value={String(n(c.inProgress))} icon="progress" tone="progress" />
            <StatTile label="Resolved" value={String(n(c.resolved))} icon="resolved" tone="resolved" />
            <StatTile label="Escalated" value={String(escalated)} icon="escalated" tone="escalated" />
          </section>

          <section className="card admin-attention" aria-labelledby="admin-attention-title">
            <h2 id="admin-attention-title">Needs attention</h2>
            {escalated === 0 && notAssigned === 0 ? (
              <p className="admin-ok">
                <Icon name="resolved" />
                Nothing needs attention.
              </p>
            ) : (
              <ul className="admin-alerts">
                {escalated > 0 && (
                  <li className="tone-escalated">
                    <Icon name="escalated" />
                    <span>
                      <strong>{reports(escalated)} escalated.</strong>
                      <span className="muted"> The time limit passed, so it went back to the Category Head.</span>
                    </span>
                  </li>
                )}
                {notAssigned > 0 && (
                  <li className="tone-open">
                    <Icon name="open" />
                    <span>
                      <strong>{reports(notAssigned)} not assigned yet.</strong>
                      <span className="muted"> Routing has not given it to anyone yet.</span>
                    </span>
                  </li>
                )}
              </ul>
            )}
          </section>

          <table className="stack-table admin-table">
            <caption>Reports by category and status</caption>
            <thead>
              <tr>
                <th scope="col">Category</th>
                <th scope="col" className="num">Open</th>
                <th scope="col" className="num">Assigned</th>
                <th scope="col" className="num">In progress</th>
                <th scope="col" className="num">Resolved</th>
                <th scope="col" className="num">Escalated</th>
                <th scope="col" className="num">Total</th>
              </tr>
            </thead>
            <tbody>
              {(data.breakdown.categories ?? []).map((row) => {
                const rc = row.counts ?? {}
                return (
                  <tr key={row.category}>
                    <th scope="row" data-label="Category">
                      {CATEGORY_NAMES[row.category] ?? row.category}
                    </th>
                    <td className="num" data-label="Open">{n(rc.open)}</td>
                    <td className="num" data-label="Assigned">{n(rc.assigned)}</td>
                    <td className="num" data-label="In progress">{n(rc.inProgress)}</td>
                    <td className="num" data-label="Resolved">{n(rc.resolved)}</td>
                    <td className="num" data-label="Escalated">{n(rc.escalated)}</td>
                    <td className="num" data-label="Total">
                      <strong>{n(rc.total)}</strong>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>

          <section className="card admin-settings" aria-labelledby="admin-settings-title">
            <h2 id="admin-settings-title">Routing rules and time limits</h2>
            <p className="muted">
              Reports go by category: Hostel to the DSA, Academic Block to the DAA, Others to the SSO. Each category has a
              time limit; when it passes, the report is escalated. Editing them here comes next, once the gateway has the
              admin routes.
            </p>
          </section>
        </div>
      )}
    </>
  )
}
