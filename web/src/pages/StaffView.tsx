// Staff view (SCRUM-36). Owner: Nyendrak.
// Category Heads and staff see the reports assigned to them, open one, read its
// details, history and notes, and move it on (Assigned > In progress > Resolved).
// Built from Gayley's design (SCRUM-138, SCRUM-141): PageHeader, Button, TextArea,
// BottomSheet, Toast, Timeline, Skeleton, EmptyState, ErrorState, StatusBadge,
// the chip and tone classes and the tokens in index.css.
// Reassign uses Zangmo's GET /positions (the picker) and
// POST /reports/{reportId}/reassign.
import { useCallback, useEffect, useRef, useState, type FormEvent, type KeyboardEvent } from 'react'
import { api, ApiError } from '../api/client'
import { useAuth } from '../auth/useAuth'
import BottomSheet from '../components/BottomSheet'
import Button from '../components/Button'
import EmptyState from '../components/EmptyState'
import Field from '../components/Field'
import ErrorState from '../components/ErrorState'
import Icon from '../components/Icon'
import PageHeader from '../components/PageHeader'
import Skeleton from '../components/Skeleton'
import StatusBadge from '../components/StatusBadge'
import TextArea from '../components/TextArea'
import Timeline, { type TimelineEntry } from '../components/Timeline'
import Toast from '../components/Toast'
import { statusLooks, type Status } from '../components/labels'
import './StaffView.css'

// Shapes as the gateway sends them (protojson, lowerCamelCase, enums by name).
type Location = { building: string; wing?: string; area?: string }
type Summary = {
  reportId: string
  category: string
  location: Location
  status: Status
  createdAt?: string
  updatedAt?: string
}
type Report = { id: string; category: string; location: Location; description: string; photoId?: string; createdAt?: string }
type ReportStatus = { status: Status; version?: number }
type HistoryEntry = {
  id: string
  eventType: string
  toStatus?: Status | 'STATUS_UNSPECIFIED'
  actorId?: string
  summary?: string
  occurredAt: string
}
type Note = { id: string; authorId: string; body: string; createdAt: string }
type Position = { code: string; title: string; kind?: string; holderUserId?: string }

/** The changes a person may make (status-lifecycle domain/transition.go). */
const NEXT: Record<Status, Status[]> = {
  OPEN: ['ASSIGNED'],
  ASSIGNED: ['IN_PROGRESS'],
  IN_PROGRESS: ['RESOLVED'],
  ESCALATED: ['IN_PROGRESS'],
  RESOLVED: [],
}

const CHIP_ORDER: Status[] = ['OPEN', 'ASSIGNED', 'IN_PROGRESS', 'ESCALATED', 'RESOLVED']
const CATEGORY_NAMES: Record<string, string> = { HOSTEL: 'Hostel', ACADEMIC_BLOCK: 'Academic Block', OTHERS: 'Others' }
const NOTE_MAX = 1000
const REASON_MAX = 500

const when = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' })

function place(l: Location): string {
  return [l.building, l.wing, l.area].filter(Boolean).join(', ')
}

function dateText(iso?: string): string {
  return iso ? when.format(new Date(iso)) : ''
}

// messageFor turns an API error into a sentence for staff.
function messageFor(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 401) return 'Your sign-in has expired. Please sign in again.'
    if (err.status === 0 || err.status === 502 || err.status === 503 || err.status === 504) {
      return 'The service is busy or unreachable. Please try again.'
    }
    return err.message
  }
  return 'Something went wrong. Please try again.'
}

// ---- Reassign ----

interface ReassignProps {
  reportId: string
  me: string
  open: boolean
  onClose: () => void
  onDone: (reportId: string, toTitle: string) => void
}

// ReassignSheet moves a report to the holder of another position. The report
// then leaves this person's list.
function ReassignSheet({ reportId, me, open, onClose, onDone }: ReassignProps) {
  const [positions, setPositions] = useState<Position[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [toUserId, setToUserId] = useState('')
  const [reason, setReason] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [loadTry, setLoadTry] = useState(0)

  // The picker loads the first time the sheet opens.
  useEffect(() => {
    if (!open || positions) return
    let alive = true
    api
      .get<{ positions?: Position[] }>('/positions')
      .then((res) => alive && setPositions(res.positions ?? []))
      .catch((err) => alive && setLoadError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [open, positions, loadTry])

  // Only positions someone holds, never my own, by title.
  const choices = (positions ?? [])
    .filter((p) => p.holderUserId && p.holderUserId !== me)
    .sort((a, b) => a.title.localeCompare(b.title))

  async function send(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (busy) return
    setError(null)
    if (!toUserId) return setError('Choose who to reassign it to.')
    const text = reason.trim()
    if (!text) return setError('Write why you are reassigning it.')
    if (reason.length > REASON_MAX) return setError(`The reason can be at most ${REASON_MAX} characters.`)
    setBusy(true)
    try {
      await api.post(`/reports/${reportId}/reassign`, { toUserId, reason: text })
      const to = choices.find((p) => p.holderUserId === toUserId)
      onDone(reportId, to?.title ?? 'the new person')
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        setError('Only the person this report is assigned to can reassign it.')
      } else if (err instanceof ApiError && err.status === 409) {
        setError('Someone reassigned this report first. Go back to the list to see it.')
      } else {
        setError(messageFor(err))
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <BottomSheet open={open} title="Reassign" onClose={onClose}>
      {loadError ? (
        <ErrorState
          message={loadError}
          onRetry={() => {
            setLoadError(null)
            setLoadTry((n) => n + 1)
          }}
        />
      ) : (
        <form className="staff-sheet-form" onSubmit={send} noValidate>
          <Field label="Reassign to" hint="The report leaves your list and goes to this person.">
            <select value={toUserId} onChange={(e) => setToUserId(e.target.value)} disabled={!positions}>
              <option value="">{positions ? 'Choose a person' : 'Loading...'}</option>
              {choices.map((p) => (
                <option key={p.code} value={p.holderUserId}>
                  {p.title}
                </option>
              ))}
            </select>
          </Field>
          <TextArea
            label="Why (the new person sees this)"
            value={reason}
            onChange={setReason}
            max={REASON_MAX}
            rows={3}
          />
          {error && (
            <p role="alert" className="field-error">
              {error}
            </p>
          )}
          <Button type="submit" loading={busy} loadingText="Reassigning...">
            Reassign report
          </Button>
        </form>
      )}
    </BottomSheet>
  )
}

// ---- One report ----

type Tab = 'details' | 'history' | 'notes'
const TABS: { id: Tab; label: string }[] = [
  { id: 'details', label: 'Details' },
  { id: 'history', label: 'History' },
  { id: 'notes', label: 'Notes' },
]

interface DetailProps {
  summary: Summary
  me: string
  onBack: () => void
  onStatus: (reportId: string, status: Status) => void
  onToast: (message: string) => void
  onReassigned: (reportId: string, toTitle: string) => void
}

function ReportDetail({ summary, me, onBack, onStatus, onToast, onReassigned }: DetailProps) {
  const id = summary.reportId
  const [report, setReport] = useState<Report | null>(null)
  const [current, setCurrent] = useState<ReportStatus>({ status: summary.status })
  const [loadError, setLoadError] = useState<string | null>(null)
  const [tab, setTab] = useState<Tab>('details')
  const [history, setHistory] = useState<HistoryEntry[] | null>(null)
  const [historyError, setHistoryError] = useState<string | null>(null)
  const [notes, setNotes] = useState<Note[] | null>(null)
  const [notesError, setNotesError] = useState<string | null>(null)
  const [historyTry, setHistoryTry] = useState(0)
  const [notesTry, setNotesTry] = useState(0)
  const [noteText, setNoteText] = useState('')
  const [noteError, setNoteError] = useState<string | null>(null)
  const [noteBusy, setNoteBusy] = useState(false)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [choice, setChoice] = useState<Status | ''>('')
  const [reason, setReason] = useState('')
  const [sheetError, setSheetError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [reassignOpen, setReassignOpen] = useState(false)
  const tabRefs = useRef<Record<Tab, HTMLButtonElement | null>>({ details: null, history: null, notes: null })

  const loadStatus = useCallback(
    () =>
      api.get<{ status: ReportStatus }>(`/reports/${id}/status`).then((res) => {
        setCurrent(res.status)
        onStatus(id, res.status.status)
        return res.status
      }),
    [id, onStatus],
  )

  // The report and its status, when the report is opened.
  useEffect(() => {
    let alive = true
    Promise.all([api.get<{ report: Report }>(`/reports/${id}`), loadStatus()])
      .then(([res]) => alive && setReport(res.report))
      .catch((err) => alive && setLoadError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [id, loadStatus])

  // History and notes load the first time their tab is shown.
  useEffect(() => {
    if (tab !== 'history' || history) return
    let alive = true
    api
      .get<{ entries?: HistoryEntry[] }>(`/reports/${id}/timeline`)
      .then((res) => alive && setHistory(res.entries ?? []))
      .catch((err) => alive && setHistoryError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [tab, history, id, historyTry])

  useEffect(() => {
    if (tab !== 'notes' || notes) return
    let alive = true
    api
      .get<{ notes?: Note[] }>(`/reports/${id}/notes`)
      .then((res) => alive && setNotes(res.notes ?? []))
      .catch((err) => alive && setNotesError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [tab, notes, id, notesTry])

  // Arrow keys, Home and End move between the tabs (WAI-ARIA tabs pattern).
  function onTabKey(e: KeyboardEvent<HTMLDivElement>) {
    const i = TABS.findIndex((t) => t.id === tab)
    let next = i
    if (e.key === 'ArrowRight') next = (i + 1) % TABS.length
    else if (e.key === 'ArrowLeft') next = (i - 1 + TABS.length) % TABS.length
    else if (e.key === 'Home') next = 0
    else if (e.key === 'End') next = TABS.length - 1
    else return
    e.preventDefault()
    setTab(TABS[next].id)
    tabRefs.current[TABS[next].id]?.focus()
  }

  // A stable function, so the sheet does not move the focus on every key press.
  const closeSheet = useCallback(() => setSheetOpen(false), [])
  const closeReassign = useCallback(() => setReassignOpen(false), [])

  function openSheet() {
    setChoice('')
    setReason('')
    setSheetError(null)
    setSheetOpen(true)
  }

  async function saveStatus(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (saving) return
    setSheetError(null)
    if (!choice) return setSheetError('Choose the new status.')
    if (reason.length > REASON_MAX) return setSheetError(`The note can be at most ${REASON_MAX} characters.`)
    setSaving(true)
    try {
      const text = reason.trim()
      const res = await api.post<{ status: ReportStatus }>(`/reports/${id}/status`, {
        toStatus: choice,
        expectedVersion: current.version ?? 0,
        ...(text && { reason: text }),
      })
      setCurrent(res.status)
      onStatus(id, res.status.status)
      setHistory(null) // the history has a new entry
      setSheetOpen(false)
      onToast('Status updated')
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setSheetError('Someone changed this report first. It now shows the latest status.')
        setChoice('')
        loadStatus().catch(() => undefined)
      } else {
        setSheetError(messageFor(err))
      }
    } finally {
      setSaving(false)
    }
  }

  async function addNote() {
    if (noteBusy) return
    setNoteError(null)
    const body = noteText.trim()
    if (!body) return setNoteError('Write the note first.')
    if (noteText.length > NOTE_MAX) return setNoteError(`A note can be at most ${NOTE_MAX} characters.`)
    setNoteBusy(true)
    try {
      const res = await api.post<{ note: Note }>(`/reports/${id}/notes`, { body })
      setNotes((list) => [...(list ?? []), res.note])
      setNoteText('')
      onToast('Note added')
    } catch (err) {
      setNoteError(messageFor(err))
    } finally {
      setNoteBusy(false)
    }
  }

  // Status changes for Gayley's Timeline: a submitted report starts Open;
  // events that are not a status change (routing, reassignment) are left out.
  const timeline: TimelineEntry[] = (history ?? []).flatMap((h) => {
    const status: Status | null =
      h.toStatus && h.toStatus !== 'STATUS_UNSPECIFIED' ? h.toStatus : h.eventType === 'report.submitted' ? 'OPEN' : null
    if (!status) return []
    const by = !h.actorId ? 'the system' : h.actorId === me ? 'you' : undefined
    return [{ status, at: h.occurredAt, by, note: h.summary }]
  })

  const options = NEXT[current.status] ?? []
  const title = place(summary.location)

  return (
    <section className="card staff-detail" aria-labelledby="staff-detail-title">
      <button type="button" className="btn btn-ghost staff-back" onClick={onBack}>
        <Icon name="back" />
        Back to the list
      </button>

      <div className="staff-detail-head">
        <div>
          <h2 id="staff-detail-title">{title}</h2>
          <p className="muted">
            {CATEGORY_NAMES[summary.category] ?? summary.category}
            {summary.createdAt && <>, reported {dateText(summary.createdAt)}</>}
          </p>
        </div>
        <StatusBadge status={current.status} />
      </div>

      {current.status === 'RESOLVED' ? (
        <p className="staff-done">
          <Icon name="resolved" />
          Resolved. Nothing more to do.
        </p>
      ) : (
        <div className="staff-actions">
          {options.length > 0 && <Button onClick={openSheet}>Update status</Button>}
          <Button variant="secondary" onClick={() => setReassignOpen(true)}>
            Reassign
          </Button>
        </div>
      )}

      <div role="tablist" aria-label="Report" className="staff-tabs" onKeyDown={onTabKey}>
        {TABS.map((t) => (
          <button
            key={t.id}
            ref={(el) => {
              tabRefs.current[t.id] = el
            }}
            type="button"
            role="tab"
            id={`tab-${t.id}`}
            aria-selected={tab === t.id}
            aria-controls={`panel-${t.id}`}
            tabIndex={tab === t.id ? 0 : -1}
            className="staff-tab"
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div role="tabpanel" id={`panel-${tab}`} aria-labelledby={`tab-${tab}`} className="staff-panel" tabIndex={0}>
        {tab === 'details' &&
          (loadError ? (
            <ErrorState message={loadError} />
          ) : !report ? (
            <Skeleton lines={3} label="Loading the report" />
          ) : (
            <dl className="staff-facts">
              <div>
                <dt>What is wrong</dt>
                <dd className="staff-description">{report.description}</dd>
              </div>
              <div>
                <dt>Where</dt>
                <dd>{place(report.location)}</dd>
              </div>
              <div>
                <dt>Photo</dt>
                <dd>{report.photoId ? 'The student added a photo.' : 'No photo.'}</dd>
              </div>
            </dl>
          ))}

        {tab === 'history' &&
          (historyError ? (
            <ErrorState
              message={historyError}
              onRetry={() => {
                setHistoryError(null)
                setHistoryTry((n) => n + 1)
              }}
            />
          ) : !history ? (
            <Skeleton lines={3} label="Loading the history" />
          ) : (
            <Timeline entries={timeline} />
          ))}

        {tab === 'notes' && (
          <div className="staff-notes">
            {notesError ? (
              <ErrorState
                message={notesError}
                onRetry={() => {
                  setNotesError(null)
                  setNotesTry((n) => n + 1)
                }}
              />
            ) : !notes ? (
              <Skeleton lines={2} label="Loading the notes" />
            ) : notes.length === 0 ? (
              <p className="muted">No notes yet. Notes are seen by staff only, never by the student.</p>
            ) : (
              <ul className="staff-note-list">
                {notes.map((n) => (
                  <li key={n.id} className="staff-note">
                    <p className="staff-note-meta">
                      <strong>{n.authorId === me ? 'You' : 'Staff'}</strong>
                      <time dateTime={n.createdAt} className="muted">
                        {dateText(n.createdAt)}
                      </time>
                    </p>
                    <p className="staff-note-body">{n.body}</p>
                  </li>
                ))}
              </ul>
            )}
            <TextArea label="New note" value={noteText} onChange={setNoteText} max={NOTE_MAX} rows={3} error={noteError} />
            <Button variant="secondary" loading={noteBusy} loadingText="Adding..." onClick={addNote}>
              Add note
            </Button>
          </div>
        )}
      </div>

      <BottomSheet open={sheetOpen} title="Update status" onClose={closeSheet}>
        <form className="staff-sheet-form" onSubmit={saveStatus} noValidate>
          <fieldset className="staff-choices">
            <legend>New status</legend>
            {options.map((s) => (
              <label key={s} className={`staff-choice tone-${statusLooks[s].tone}`}>
                <input
                  type="radio"
                  name="to-status"
                  value={s}
                  checked={choice === s}
                  onChange={() => setChoice(s)}
                />
                <Icon name={statusLooks[s].icon} />
                {statusLooks[s].label}
              </label>
            ))}
          </fieldset>
          <TextArea
            label="Note for the history (optional)"
            value={reason}
            onChange={setReason}
            max={REASON_MAX}
            rows={3}
          />
          {sheetError && (
            <p role="alert" className="field-error">
              {sheetError}
            </p>
          )}
          <Button type="submit" loading={saving} loadingText="Saving...">
            Save status
          </Button>
        </form>
      </BottomSheet>

      <ReassignSheet reportId={id} me={me} open={reassignOpen} onClose={closeReassign} onDone={onReassigned} />
    </section>
  )
}

// ---- The list ----

export default function StaffView() {
  const { user } = useAuth()
  const me = user?.id ?? ''
  const [reports, setReports] = useState<Summary[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [nextToken, setNextToken] = useState('')
  const [moreBusy, setMoreBusy] = useState(false)
  const [filter, setFilter] = useState<Status | 'ALL'>('ALL')
  const [openId, setOpenId] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)
  const [attempt, setAttempt] = useState(0)

  // The first page; "Try again" loads it again.
  useEffect(() => {
    let alive = true
    api
      .get<{ reports?: Summary[]; nextPageToken?: string }>('/me/assigned-reports')
      .then((res) => {
        if (!alive) return
        setReports(res.reports ?? [])
        setNextToken(res.nextPageToken ?? '')
      })
      .catch((err) => alive && setLoadError(messageFor(err)))
    return () => {
      alive = false
    }
  }, [attempt])

  function retry() {
    setLoadError(null)
    setReports(null)
    setAttempt((n) => n + 1)
  }

  async function loadMore() {
    if (!nextToken || moreBusy) return
    setMoreBusy(true)
    try {
      const res = await api.get<{ reports?: Summary[]; nextPageToken?: string }>(
        `/me/assigned-reports?pageToken=${encodeURIComponent(nextToken)}`,
      )
      setReports((list) => [...(list ?? []), ...(res.reports ?? [])])
      setNextToken(res.nextPageToken ?? '')
    } catch (err) {
      setToast(messageFor(err))
    } finally {
      setMoreBusy(false)
    }
  }

  // Stable functions: the detail and the toast keep them in effects.
  const updateStatus = useCallback((reportId: string, status: Status) => {
    setReports((list) => list?.map((r) => (r.reportId === reportId ? { ...r, status } : r)) ?? list)
  }, [])
  const showToast = useCallback((message: string) => setToast(message), [])
  // A reassigned report is no longer mine: it leaves the list.
  const reassigned = useCallback((reportId: string, toTitle: string) => {
    setReports((list) => list?.filter((r) => r.reportId !== reportId) ?? list)
    setOpenId(null)
    setToast(`Reassigned to ${toTitle}`)
  }, [])
  const hideToast = useCallback(() => setToast(null), [])
  const closeDetail = useCallback(() => setOpenId(null), [])

  const counts = (reports ?? []).reduce<Record<string, number>>((acc, r) => {
    acc[r.status] = (acc[r.status] ?? 0) + 1
    return acc
  }, {})
  const shown = (reports ?? []).filter((r) => filter === 'ALL' || r.status === filter)
  const opened = reports?.find((r) => r.reportId === openId) ?? null

  let list
  if (loadError) {
    list = <ErrorState message={loadError} onRetry={retry} />
  } else if (!reports) {
    list = <Skeleton lines={4} label="Loading reports" />
  } else if (reports.length === 0) {
    list = (
      <EmptyState title="Nothing assigned to you">
        <p>Reports routed or reassigned to you appear here.</p>
      </EmptyState>
    )
  } else {
    list = (
      <>
        <div className="chips staff-chips" aria-label="Filter by status" role="group">
          <button type="button" className="chip" aria-pressed={filter === 'ALL'} onClick={() => setFilter('ALL')}>
            All <span className="chip-count">{reports.length}</span>
          </button>
          {CHIP_ORDER.filter((s) => counts[s]).map((s) => (
            <button
              key={s}
              type="button"
              className={`chip tone-${statusLooks[s].tone}`}
              aria-pressed={filter === s}
              onClick={() => setFilter(s)}
            >
              {statusLooks[s].label} <span className="chip-count">{counts[s]}</span>
            </button>
          ))}
        </div>

        <ul className="staff-list">
          {shown.map((r) => (
            <li key={r.reportId}>
              <button
                type="button"
                className={`staff-card tone-${statusLooks[r.status]?.tone ?? 'open'}${r.reportId === openId ? ' is-open' : ''}`}
                aria-current={r.reportId === openId || undefined}
                onClick={() => setOpenId(r.reportId)}
              >
                <span className="staff-card-place">{place(r.location)}</span>
                <span className="staff-card-meta muted">
                  {CATEGORY_NAMES[r.category] ?? r.category}
                  {r.updatedAt && <>, updated {dateText(r.updatedAt)}</>}
                </span>
                <StatusBadge status={r.status} />
              </button>
            </li>
          ))}
        </ul>

        {nextToken && (
          <Button variant="secondary" loading={moreBusy} loadingText="Loading..." onClick={loadMore}>
            Load more
          </Button>
        )}
      </>
    )
  }

  return (
    <div className={opened ? 'staff-page has-detail' : 'staff-page'}>
      <PageHeader title="Staff view" description="Reports assigned to you. Open one to read it, add a note or update its status." />
      <div className="staff-layout">
        <div className="staff-list-pane">{list}</div>
        {opened && (
          <ReportDetail
            key={opened.reportId}
            summary={opened}
            me={me}
            onBack={closeDetail}
            onStatus={updateStatus}
            onToast={showToast}
            onReassigned={reassigned}
          />
        )}
      </div>
      {toast && <Toast message={toast} onDone={hideToast} />}
    </div>
  )
}
