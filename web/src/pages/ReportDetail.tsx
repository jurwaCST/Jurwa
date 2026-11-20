// One of a student's own reports (SCRUM-143). Owner: Gayley.
// Reads GET /reports/{id}, its status and its timeline, and the photo from
// GET /photos/{photoId}. The gateway refuses a report that is not the student's
// (403); that and a missing report (404) show the same kind message.
import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { useAuth } from '../auth/useAuth'
import BottomSheet from '../components/BottomSheet'
import Button from '../components/Button'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import Icon from '../components/Icon'
import Skeleton from '../components/Skeleton'
import StatusBadge from '../components/StatusBadge'
import TextArea from '../components/TextArea'
import Timeline, { type TimelineEntry } from '../components/Timeline'
import { LostArt } from '../components/Art'
import Feedback from './Feedback'
import type { Status } from '../components/labels'
import { categoryNames, messageFor, place, reportNumber, type Location } from './reportText'
import './MyReports.css'

type Report = { id: string; category: string; location: Location; description: string; photoId?: string; createdAt?: string }
type HistoryEntry = {
  eventType: string
  toStatus?: Status | 'STATUS_UNSPECIFIED'
  actorId?: string
  summary?: string
  occurredAt: string
}

const when = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' })

/** The report's photo, fetched with the sign-in token and shown from a blob URL. */
function Photo({ photoId }: { photoId: string }) {
  const [url, setUrl] = useState<string | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let alive = true
    let made: string | null = null
    api
      .blob(`/photos/${photoId}`)
      .then((b) => {
        if (!alive) return
        made = URL.createObjectURL(b)
        setUrl(made)
      })
      .catch(() => alive && setFailed(true))
    return () => {
      alive = false
      if (made) URL.revokeObjectURL(made)
    }
  }, [photoId])

  if (failed) return <p className="muted">The photo could not be loaded.</p>
  if (!url) return <Skeleton lines={1} label="Loading the photo" />
  return <img src={url} alt="The photo you added" className="rd-photo" />
}

// SCRUM-163: what the student sees when work started before their change arrived.
const WORK_STARTED = 'Someone has started work on it, so it can no longer be changed.'
const MAX_DESCRIPTION = 500

function changeError(err: unknown): string {
  if (err instanceof ApiError && err.status === 422) return WORK_STARTED
  return messageFor(err)
}

interface ChangeProps {
  reportId: string
  description: string
  onEdited: (text: string) => void
}

// ChangeReport lets the student change the description or delete the report
// while nobody works on it yet (SCRUM-163): Open or Assigned only.
function ChangeReport({ reportId, description, onEdited }: ChangeProps) {
  const navigate = useNavigate()
  const [editing, setEditing] = useState(false)
  const [text, setText] = useState(description)
  const [asking, setAsking] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function save() {
    if (busy) return
    setError(null)
    if (!text.trim()) return setError('Describe the problem.')
    setBusy(true)
    try {
      const saved = await api.patch<{ description?: string }>(`/reports/${reportId}`, { description: text })
      onEdited(saved?.description ?? text)
      setEditing(false)
    } catch (err) {
      setError(changeError(err))
    } finally {
      setBusy(false)
    }
  }

  async function remove() {
    if (busy) return
    setError(null)
    setBusy(true)
    try {
      await api.del(`/reports/${reportId}`)
      navigate('/reports', { replace: true })
    } catch (err) {
      setError(changeError(err))
      setAsking(false)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="rd-change">
      {editing ? (
        <>
          <TextArea label="Description" value={text} onChange={setText} max={MAX_DESCRIPTION} rows={4} />
          <div className="rd-change-actions">
            <Button onClick={save} loading={busy} loadingText="Saving...">
              Save description
            </Button>
            <Button variant="secondary" onClick={() => (setEditing(false), setText(description), setError(null))}>
              Cancel
            </Button>
          </div>
        </>
      ) : (
        <div className="rd-change-actions">
          <Button variant="secondary" onClick={() => (setText(description), setEditing(true))}>
            Edit description
          </Button>
          <Button variant="secondary" className="rd-delete" onClick={() => setAsking(true)}>
            Delete report
          </Button>
        </div>
      )}
      {error && (
        <p role="alert" className="field-error">
          {error}
        </p>
      )}
      <BottomSheet open={asking} title="Delete this report?" onClose={() => setAsking(false)}>
        <p>It is removed for you and for the person who would fix it. This cannot be undone.</p>
        <div className="rd-change-actions">
          <Button onClick={remove} loading={busy} loadingText="Deleting...">
            Yes, delete it
          </Button>
          <Button variant="secondary" onClick={() => setAsking(false)}>
            Keep it
          </Button>
        </div>
      </BottomSheet>
    </div>
  )
}

export default function ReportDetail() {
  const { reportId = '' } = useParams()
  const { user } = useAuth()
  const me = user?.id ?? ''
  // What came back for which attempt; anything else is loading.
  const [loaded, setLoaded] = useState<
    { key: string; report: Report; status: Status | null } | { key: string; notFound: true } | { key: string; error: string } | null
  >(null)
  const [history, setHistory] = useState<{ key: string; entries?: HistoryEntry[]; error?: string } | null>(null)
  const [attempt, setAttempt] = useState(0)
  const [historyAttempt, setHistoryAttempt] = useState(0)
  // SCRUM-163: a description the student changed on this page, for this load.
  const [edited, setEdited] = useState<{ key: string; text: string } | null>(null)
  const key = `${reportId}-${attempt}`
  const historyKey = `${reportId}-${historyAttempt}`

  useEffect(() => {
    let alive = true
    // The status comes from another service; when it does not answer, the
    // report is still shown, without its badge.
    // GET /reports/{id} sends the report itself (POST /reports wraps it in { report }).
    Promise.all([
      api.get<Report>(`/reports/${reportId}`),
      api
        .get<{ status: { status: Status } }>(`/reports/${reportId}/status`)
        .then((s) => s.status.status)
        .catch(() => null),
    ])
      .then(([r, status]) => alive && setLoaded({ key, report: r, status }))
      .catch((err) => {
        if (!alive) return
        if (err instanceof ApiError && (err.status === 403 || err.status === 404)) setLoaded({ key, notFound: true })
        else setLoaded({ key, error: messageFor(err) })
      })
    return () => {
      alive = false
    }
  }, [reportId, key])

  useEffect(() => {
    let alive = true
    api
      .get<{ entries?: HistoryEntry[] }>(`/reports/${reportId}/timeline`)
      .then((res) => alive && setHistory({ key: historyKey, entries: res.entries ?? [] }))
      .catch((err) => alive && setHistory({ key: historyKey, error: messageFor(err) }))
    return () => {
      alive = false
    }
  }, [reportId, historyKey])

  const back = (
    <Link to="/reports" className="btn btn-ghost rd-back">
      <Icon name="back" />
      My reports
    </Link>
  )

  const current = loaded?.key === key ? loaded : null
  if (!current) return <Skeleton lines={6} label="Loading the report" />

  if ('notFound' in current) {
    return (
      <EmptyState title="Report not found" art={<LostArt />}>
        <p>This report does not exist, or it is not one of yours.</p>
        <Link to="/reports" className="action">
          Back to My reports
        </Link>
      </EmptyState>
    )
  }

  if ('error' in current) {
    return (
      <>
        {back}
        <ErrorState message={current.error} onRetry={() => setAttempt((n) => n + 1)} />
      </>
    )
  }

  const { report, status } = current

  // Status changes, as in the staff view (SCRUM-36): a submitted report starts
  // Open; events that are not a status change (routing) are left out.
  const shownHistory = history?.key === historyKey ? history : null
  const timeline: TimelineEntry[] = (shownHistory?.entries ?? []).flatMap((h) => {
    const s: Status | null =
      h.toStatus && h.toStatus !== 'STATUS_UNSPECIFIED' ? h.toStatus : h.eventType === 'report.submitted' ? 'OPEN' : null
    if (!s) return []
    const by = !h.actorId ? 'the system' : h.actorId === me ? 'you' : undefined
    return [{ status: s, at: h.occurredAt, by, note: h.summary }]
  })

  return (
    <div className="rd-page">
      {back}

      <div className="page-header rd-head">
        <div>
          <h1>{place(report.location)}</h1>
          <p className="muted">
            <span className="myr-number">{reportNumber(report.id)}</span>
            {categoryNames[report.category] ?? report.category}
            {report.createdAt && <>, sent {when.format(new Date(report.createdAt))}</>}
          </p>
        </div>
        {status && <StatusBadge status={status} />}
      </div>

      <div className="rd-layout">
        <section className="card" aria-labelledby="rd-what">
          <h2 id="rd-what">What you reported</h2>
          <p className="rd-description">{edited?.key === key ? edited.text : report.description}</p>
          {/* SCRUM-163: changeable until someone starts work on it. */}
          {(status === 'OPEN' || status === 'ASSIGNED') && (
            <ChangeReport
              reportId={reportId}
              description={edited?.key === key ? edited.text : report.description}
              onEdited={(text) => setEdited({ key, text })}
            />
          )}
          {report.photoId ? <Photo photoId={report.photoId} /> : <p className="muted">No photo added.</p>}
        </section>

        <section className="card" aria-labelledby="rd-history">
          <h2 id="rd-history">What happened</h2>
          {shownHistory?.error ? (
            <ErrorState message={shownHistory.error} onRetry={() => setHistoryAttempt((n) => n + 1)} />
          ) : !shownHistory ? (
            <Skeleton lines={3} label="Loading what happened" />
          ) : (
            <Timeline entries={timeline} />
          )}
        </section>
      </div>

      {/* SCRUM-145: only a resolved report can be rated. */}
      {status === 'RESOLVED' && <Feedback reportId={reportId} />}
    </div>
  )
}
