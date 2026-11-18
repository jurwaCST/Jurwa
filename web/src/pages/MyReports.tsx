// My reports (SCRUM-143). Owner: Gayley.
// A student's own reports, newest first, from GET /me/reports, with chips to
// filter by status. Each card opens the report page (ReportDetail.tsx).
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import Button from '../components/Button'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import PageHeader from '../components/PageHeader'
import Skeleton from '../components/Skeleton'
import StatusBadge from '../components/StatusBadge'
import { statusLooks, type Status } from '../components/labels'
import { age, categoryNames, messageFor, place, reportNumber, type Location } from './reportText'
import './MyReports.css'

// As the gateway sends it (report-history ListReportsByReporter).
type Summary = {
  reportId: string
  category: string
  location: Location
  status: Status
  createdAt?: string
  updatedAt?: string
}
type Page = { reports?: Summary[]; nextPageToken?: string }

const CHIP_ORDER: Status[] = ['OPEN', 'ASSIGNED', 'IN_PROGRESS', 'ESCALATED', 'RESOLVED']

export default function MyReports() {
  const [reports, setReports] = useState<Summary[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [nextToken, setNextToken] = useState('')
  const [moreBusy, setMoreBusy] = useState(false)
  const [moreError, setMoreError] = useState<string | null>(null)
  const [filter, setFilter] = useState<Status | 'ALL'>('ALL')
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let alive = true
    api
      .get<Page>('/me/reports')
      .then((res) => {
        if (!alive) return
        setReports(res.reports ?? [])
        setNextToken(res.nextPageToken ?? '')
      })
      .catch((err) => alive && setError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [attempt])

  function retry() {
    setError(null)
    setReports(null)
    setAttempt((n) => n + 1)
  }

  async function loadMore() {
    if (!nextToken || moreBusy) return
    setMoreBusy(true)
    setMoreError(null)
    try {
      const res = await api.get<Page>(`/me/reports?pageToken=${encodeURIComponent(nextToken)}`)
      setReports((list) => [...(list ?? []), ...(res.reports ?? [])])
      setNextToken(res.nextPageToken ?? '')
    } catch (err) {
      setMoreError(messageFor(err))
    } finally {
      setMoreBusy(false)
    }
  }

  const counts = (reports ?? []).reduce<Partial<Record<Status, number>>>((acc, r) => {
    acc[r.status] = (acc[r.status] ?? 0) + 1
    return acc
  }, {})
  const shown = (reports ?? []).filter((r) => filter === 'ALL' || r.status === filter)

  let body
  if (error) {
    body = <ErrorState message={error} onRetry={retry} />
  } else if (!reports) {
    body = <Skeleton lines={4} label="Loading your reports" />
  } else if (reports.length === 0) {
    body = (
      <EmptyState title="No reports yet">
        <p>When you report something broken, you can follow it here.</p>
        <Link to="/report/new" className="action">
          Report an issue
        </Link>
      </EmptyState>
    )
  } else {
    body = (
      <>
        <div className="chips myr-chips" role="group" aria-label="Filter by status">
          <button type="button" className="chip" aria-pressed={filter === 'ALL'} onClick={() => setFilter('ALL')}>
            All <span className="myr-count">{reports.length}</span>
          </button>
          {CHIP_ORDER.filter((s) => counts[s]).map((s) => (
            <button
              key={s}
              type="button"
              className={`chip tone-${statusLooks[s].tone}`}
              aria-pressed={filter === s}
              onClick={() => setFilter(s)}
            >
              {statusLooks[s].label} <span className="myr-count">{counts[s]}</span>
            </button>
          ))}
        </div>

        <ul className="myr-list">
          {shown.map((r) => (
            <li key={r.reportId} className={`myr-card tone-${statusLooks[r.status]?.tone ?? 'open'}`}>
              <Link to={`/reports/${r.reportId}`} className="myr-place">
                {place(r.location)}
              </Link>
              <span className="myr-meta muted">
                <span className="myr-number">{reportNumber(r.reportId)}</span>
                {categoryNames[r.category] ?? r.category}, sent {age(r.createdAt)}
              </span>
              <StatusBadge status={r.status} />
            </li>
          ))}
        </ul>

        {nextToken && (
          <Button variant="secondary" loading={moreBusy} loadingText="Loading..." onClick={loadMore}>
            Load more
          </Button>
        )}
        {moreError && (
          <p role="alert" className="field-error">
            {moreError}
          </p>
        )}
      </>
    )
  }

  return (
    <div className="myr-page">
      <PageHeader title="My reports" description="Everything you have reported, newest first. Open one to see what has happened." />
      {body}
    </div>
  )
}
