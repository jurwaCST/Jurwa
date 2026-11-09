import { useEffect, useState } from 'react'
import { api, ApiError } from '../api/client'
import ErrorState from '../components/ErrorState'
import Icon from '../components/Icon'
import PageHeader from '../components/PageHeader'
import Skeleton from '../components/Skeleton'
import StatTile from '../components/StatTile'
import './PresidentDashboard.css'

// The College President's read-only dashboard (SCRUM-99, FR-29, FR-30). It only
// reads GET /dashboard/summary and GET /dashboard/categories (analytics through
// the gateway); there is nothing on it that changes data.

// int64 counts arrive as strings (protojson); 0 may be left out.
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

const categoryNames: Record<string, string> = {
  HOSTEL: 'Hostel',
  ACADEMIC_BLOCK: 'Academic Block',
  OTHERS: 'Others',
}

type Period = 'week' | 'month' | 'all'
const periods: { id: Period; label: string }[] = [
  { id: 'week', label: 'This week' },
  { id: 'month', label: 'This month' },
  { id: 'all', label: 'All time' },
]

// The query for a period. Times are sent in UTC ("Z"): no "+" or space, so the
// value is safe in a URL as it is (a "+06:00" offset would need encoding).
function query(p: Period): string {
  if (p === 'all') return ''
  const now = new Date()
  const from = p === 'week' ? new Date(now.getTime() - 7 * 24 * 3600 * 1000) : new Date(now.getFullYear(), now.getMonth(), 1)
  return `?from=${from.toISOString()}`
}

function averageTime(hours: number): string {
  if (hours <= 0) return 'Nothing resolved yet'
  if (hours < 1) return 'Under an hour'
  if (hours < 48) return `${Math.round(hours)} hours`
  return `${Math.round(hours / 24)} days`
}

/** A ring showing the share of reports resolved. */
function Ring({ percent }: { percent: number }) {
  const r = 42
  const length = 2 * Math.PI * r
  return (
    <svg viewBox="0 0 100 100" className="ring" aria-hidden="true" focusable="false">
      <circle cx="50" cy="50" r={r} className="ring-track" />
      <circle
        cx="50"
        cy="50"
        r={r}
        className="ring-value"
        strokeDasharray={`${(percent / 100) * length} ${length}`}
        transform="rotate(-90 50 50)"
      />
    </svg>
  )
}

/** A thin bar split by status, so a row's mix shows at a glance. */
function MixBar({ c }: { c: Counts }) {
  const total = n(c.total)
  if (total === 0) return <span className="muted">No reports</span>
  const parts: [string, number][] = [
    ['open', n(c.open)],
    ['assigned', n(c.assigned)],
    ['progress', n(c.inProgress)],
    ['resolved', n(c.resolved)],
  ]
  return (
    <span className="mix" aria-hidden="true">
      {parts.map(([tone, v]) =>
        v > 0 ? <span key={tone} className={`mix-part tone-${tone}`} style={{ flexGrow: v }} /> : null,
      )}
    </span>
  )
}

export default function PresidentDashboard() {
  const [period, setPeriod] = useState<Period>('all')
  const [attempt, setAttempt] = useState(0)
  // What came back, and for which period and attempt; anything else is loading.
  const [result, setResult] = useState<
    { key: string; summary: Summary; breakdown: Breakdown } | { key: string; error: string } | null
  >(null)
  const key = `${period}-${attempt}`

  useEffect(() => {
    let current = true
    const q = query(period)
    Promise.all([api.get<Summary>(`/dashboard/summary${q}`), api.get<Breakdown>(`/dashboard/categories${q}`)])
      .then(([summary, breakdown]) => {
        if (current) setResult({ key, summary, breakdown })
      })
      .catch((err) => {
        if (current) setResult({ key, error: err instanceof ApiError ? err.message : 'The dashboard could not load.' })
      })
    return () => {
      current = false
    }
  }, [period, key])

  const loaded = result?.key === key ? result : null
  const error = loaded && 'error' in loaded ? loaded.error : null
  const data = loaded && 'summary' in loaded ? loaded : null

  const c = data?.summary.counts ?? {}
  const total = n(c.total)
  const rate = total > 0 ? Math.round((n(c.resolved) / total) * 100) : 0

  return (
    <>
      <PageHeader
        title="President dashboard"
        description="How reports across campus are being handled."
        aside={
          <span className="tag">
            <Icon name="readonly" size={16} />
            Read-only
          </span>
        }
      />

      <ul className="chips period-chips" aria-label="Period">
        {periods.map((p) => (
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
        <>
          <section className="dash-tiles" aria-label="Report counts">
            <StatTile label="Open" value={String(n(c.open))} icon="open" tone="open" />
            <StatTile label="Assigned" value={String(n(c.assigned))} icon="assigned" tone="assigned" />
            <StatTile label="In progress" value={String(n(c.inProgress))} icon="progress" tone="progress" />
            <StatTile label="Resolved" value={String(n(c.resolved))} icon="resolved" tone="resolved" />
            <StatTile label="Escalated" value={String(n(c.escalated))} icon="escalated" tone="escalated" />
          </section>

          <section className="dash-pair" aria-label="Resolution">
            <div className="card dash-rate">
              <Ring percent={rate} />
              <div>
                <h2>Resolution rate</h2>
                <p className="dash-big num">{rate}%</p>
                <p className="muted">
                  {n(c.resolved)} of {total} reports resolved
                </p>
              </div>
            </div>
            <div className="card">
              <h2>Average time to resolve</h2>
              <p className="dash-big">{averageTime(data.summary.averageResolutionHours ?? 0)}</p>
              <p className="muted">From the report being sent to it being marked resolved.</p>
            </div>
          </section>

          <table className="stack-table dash-table">
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
                <th scope="col">Mix</th>
              </tr>
            </thead>
            <tbody>
              {(data.breakdown.categories ?? []).map((row) => {
                const rc = row.counts ?? {}
                return (
                  <tr key={row.category}>
                    <th scope="row" data-label="Category">
                      {categoryNames[row.category] ?? row.category}
                    </th>
                    <td className="num" data-label="Open">{n(rc.open)}</td>
                    <td className="num" data-label="Assigned">{n(rc.assigned)}</td>
                    <td className="num" data-label="In progress">{n(rc.inProgress)}</td>
                    <td className="num" data-label="Resolved">{n(rc.resolved)}</td>
                    <td className="num" data-label="Escalated">{n(rc.escalated)}</td>
                    <td className="num" data-label="Total"><strong>{n(rc.total)}</strong></td>
                    <td data-label="Mix"><MixBar c={rc} /></td>
                  </tr>
                )
              })}
            </tbody>
          </table>

          <p className="muted dash-note">
            Escalated counts each report that was escalated at least once. A list of the escalated reports needs a new
            analytics route and comes later.
          </p>
        </>
      )}
    </>
  )
}
