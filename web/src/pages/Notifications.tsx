// A student's notifications (SCRUM-144, FR-09). Owner: Gayley.
// GET /me/notifications lists the messages newest first (sent by the
// notification service from report events, SCRUM-155). They are grouped by
// day. Opening the page marks the newest one as seen, so the bell dot goes.
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAuth } from '../auth/useAuth'
import Button from '../components/Button'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import PageHeader from '../components/PageHeader'
import Skeleton from '../components/Skeleton'
import { markSeen } from '../notifications/seen'
import { messageFor } from './reportText'
import './MyReports.css'

type Note = { id: string; reportId?: string; subject: string; body: string; createdAt: string }
type Page = { notifications?: Note[]; nextPageToken?: string }

const dayFormat = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium' })
const timeFormat = new Intl.DateTimeFormat('en-GB', { timeStyle: 'short' })

/** The local calendar day of a time, for grouping. */
function dayKey(d: Date): string {
  return `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`
}

/** "Today", "Yesterday" or the date, for a day heading. */
function dayLabel(d: Date): string {
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  if (dayKey(d) === dayKey(today)) return 'Today'
  if (dayKey(d) === dayKey(yesterday)) return 'Yesterday'
  return dayFormat.format(d)
}

/** Messages in order, cut into days. The list is already newest first. */
function byDay(list: Note[]): { label: string; notes: Note[] }[] {
  const groups: { key: string; label: string; notes: Note[] }[] = []
  for (const n of list) {
    const d = new Date(n.createdAt)
    const k = dayKey(d)
    const last = groups[groups.length - 1]
    if (last && last.key === k) last.notes.push(n)
    else groups.push({ key: k, label: dayLabel(d), notes: [n] })
  }
  return groups
}

export default function Notifications() {
  const { user } = useAuth()
  const me = user?.id ?? ''
  const [notes, setNotes] = useState<Note[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [nextToken, setNextToken] = useState('')
  const [moreBusy, setMoreBusy] = useState(false)
  const [moreError, setMoreError] = useState<string | null>(null)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let alive = true
    api
      .get<Page>('/me/notifications')
      .then((res) => {
        if (!alive) return
        const list = res.notifications ?? []
        setNotes(list)
        setNextToken(res.nextPageToken ?? '')
        if (me && list[0]) markSeen(me, list[0].createdAt)
      })
      .catch((err) => alive && setError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [attempt, me])

  function retry() {
    setError(null)
    setNotes(null)
    setAttempt((n) => n + 1)
  }

  async function loadMore() {
    if (!nextToken || moreBusy) return
    setMoreBusy(true)
    setMoreError(null)
    try {
      const res = await api.get<Page>(`/me/notifications?pageToken=${encodeURIComponent(nextToken)}`)
      setNotes((list) => [...(list ?? []), ...(res.notifications ?? [])])
      setNextToken(res.nextPageToken ?? '')
    } catch (err) {
      setMoreError(messageFor(err))
    } finally {
      setMoreBusy(false)
    }
  }

  let body
  if (error) {
    body = <ErrorState message={error} onRetry={retry} />
  } else if (!notes) {
    body = <Skeleton lines={4} label="Loading your notifications" />
  } else if (notes.length === 0) {
    body = (
      <EmptyState title="No notifications yet">
        <p>When something happens to a report you sent, you will see it here.</p>
        <Link to="/report/new" className="action">
          Report an issue
        </Link>
      </EmptyState>
    )
  } else {
    body = (
      <>
        {byDay(notes).map((g) => (
          <section key={g.label} className="nt-day" aria-label={g.label}>
            <h2>{g.label}</h2>
            <ul className="nt-list">
              {g.notes.map((n) => (
                <li key={n.id} className="nt-item card">
                  {n.reportId ? (
                    <Link to={`/reports/${n.reportId}`} className="nt-subject">
                      {n.subject}
                    </Link>
                  ) : (
                    <span className="nt-subject">{n.subject}</span>
                  )}
                  <p className="nt-body">{n.body}</p>
                  <time className="muted nt-time" dateTime={n.createdAt}>
                    {timeFormat.format(new Date(n.createdAt))}
                  </time>
                </li>
              ))}
            </ul>
          </section>
        ))}

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
    <div className="nt-page">
      <PageHeader title="Notifications" description="What has happened to the reports you sent, newest first." />
      {body}
    </div>
  )
}
