// Student report form (SCRUM-35). Owner: Nyendrak.
// Category and location come from the gateway; the photo is uploaded first and
// its photoId is sent with the report (report-submission SubmitReport).
// Built from Gayley's design (SCRUM-138, SCRUM-141): PageHeader, Field, Button,
// Icon, StatusBadge, the card class and the tokens in index.css.
// The category is a set of tiles (radio buttons, so the arrow keys move between
// them); every other field stays inside its <label> with the same text, so
// tests find it by its label.
import { useEffect, useMemo, useState, type ChangeEvent, type DragEvent, type FormEvent, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import Button from '../components/Button'
import Field from '../components/Field'
import Icon, { type IconName } from '../components/Icon'
import PageHeader from '../components/PageHeader'
import StatusBadge from '../components/StatusBadge'
import './ReportForm.css'

/** MAX_DESCRIPTION is report-submission's limit. */
export const MAX_DESCRIPTION = 500
/** MAX_PHOTO_BYTES is photo-media's limit: 5 MB, and exactly 5 MB is allowed. */
export const MAX_PHOTO_BYTES = 5 * 1024 * 1024
const PHOTO_TYPES = ['image/jpeg', 'image/png']

type CategoryInfo = { category: string; name: string; description?: string; sortOrder?: number; active?: boolean }
type Location = { building: string; wing?: string; area?: string }
type LocationEntry = { id: string; category: string; location: Location }

// A problem to show: the text, and whether the student must sign in again.
type Problem = { text: string; signIn: boolean }
type FieldName = 'category' | 'location' | 'description'

const EXPIRED: Problem = { text: 'Your sign-in has expired. Please sign in again.', signIn: true }

// Icons this form needs that are not in components/Icon.tsx yet. Same size,
// stroke and markup as Gayley's, so they look the same.
const formPaths = {
  category: 'M3 12V4h8l10 10-8 8-10-10ZM7.5 8h.01',
  hostel: 'M3 19V6M3 15h18M21 19v-6a3 3 0 0 0-3-3h-8v5M6.5 12.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Z',
  academic: 'M3 21h18M4 10h16M12 3 3.5 8h17L12 3ZM6 10v8M10 10v8M14 10v8M18 10v8',
  others: 'M12 21v-5M6 16h12L12 3 6 16Z',
  place: 'M12 21s-7-6.2-7-11.5a7 7 0 0 1 14 0C19 14.8 12 21 12 21ZM12 12a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
  describe: 'M4 20h4L19 9l-4-4L4 16v4ZM13.5 6.5l4 4',
  camera: 'M4 8h3l2-3h6l2 3h3v11H4V8ZM12 16.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z',
  upload: 'M12 16V4M7 9l5-5 5 5M4 16v4h16v-4',
  send: 'M21 3 10 14M21 3l-7 18-4-7-7-4 18-7Z',
  copy: 'M9 9h11v11H9V9ZM5 15H4V4h11v1',
  check: 'M5 12.5l4.5 4.5L19 7',
  alert: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18ZM12 8v4.5M12 16h.01',
  info: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18ZM12 11v5M12 8h.01',
  remove: 'M4 7h16M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3',
} as const

type FormIconName = keyof typeof formPaths

// The icon on each category tile; a new category gets the general tag icon.
const categoryIcons: Record<string, FormIconName> = {
  HOSTEL: 'hostel',
  ACADEMIC_BLOCK: 'academic',
  OTHERS: 'others',
}

// FormIcon draws one of formPaths. Decorative: the text next to it says the same.
function FormIcon({ name, size = 20 }: { name: FormIconName; size?: number }) {
  return (
    <svg
      className="icon"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
      focusable="false"
    >
      <path d={formPaths[name]} />
    </svg>
  )
}

// unique keeps the first copy of each non-empty value, in order.
function unique(values: (string | undefined)[]): string[] {
  return values.filter((v, i): v is string => !!v && values.indexOf(v) === i)
}

// problemFor turns an API error into a message for the student. An expired
// sign-in always says so; otherwise fallback (when given) is the message.
function problemFor(err: unknown, fallback: string): Problem {
  if (err instanceof ApiError && err.status === 401) return EXPIRED
  if (fallback) return { text: fallback, signIn: false }
  if (err instanceof ApiError) {
    if (err.status === 0 || err.status === 502 || err.status === 503 || err.status === 504) {
      return { text: 'The service is busy or unreachable. Please try again.', signIn: false }
    }
    return { text: err.message, signIn: false }
  }
  return { text: 'Something went wrong. Please try again.', signIn: false }
}

// fileSize gives "2 KB" or "3.4 MB".
function fileSize(bytes: number): string {
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function ProblemText({ problem }: { problem: Problem }) {
  return (
    <p role="alert" className="report-problem">
      <FormIcon name="alert" />
      <span>
        {problem.text}
        {problem.signIn && (
          <>
            {' '}
            <Link to="/login">Sign in again</Link>
          </>
        )}
      </span>
    </p>
  )
}

// Step is one part of the form, with its icon in a round badge.
function Step({ icon, title, children }: { icon: FormIconName; title: string; children: ReactNode }) {
  return (
    <section className="report-section">
      <h2 className="report-step">
        <span className="report-step-badge">
          <FormIcon name={icon} />
        </span>
        {title}
      </h2>
      {children}
    </section>
  )
}

// What happens after sending: the first stops of Gayley's route on Home.
const nextSteps: { icon: IconName; title: string; text: string }[] = [
  { icon: 'open', title: 'Open', text: 'Your report gets a number as soon as it is sent.' },
  { icon: 'assigned', title: 'Assigned', text: 'It goes to the Category Head for its category, by category and place only.' },
  { icon: 'escalated', title: 'Not handled in time?', text: 'Every category has a time limit. When it passes, the report is escalated.' },
]

function NextSteps() {
  return (
    <aside className="card report-next" aria-labelledby="report-next-title">
      <h2 id="report-next-title">
        <FormIcon name="info" />
        What happens next
      </h2>
      <ol>
        {nextSteps.map((s) => (
          <li key={s.title}>
            <span className="report-next-icon">
              <Icon name={s.icon} size={18} />
            </span>
            <div>
              <strong>{s.title}</strong>
              <p className="muted">{s.text}</p>
            </div>
          </li>
        ))}
      </ol>
    </aside>
  )
}

export default function ReportForm() {
  const [categories, setCategories] = useState<CategoryInfo[]>([])
  const [loadProblem, setLoadProblem] = useState<Problem | null>(null)
  const [category, setCategory] = useState('')
  const [locations, setLocations] = useState<LocationEntry[]>([])
  const [building, setBuilding] = useState('')
  const [wing, setWing] = useState('')
  const [area, setArea] = useState('')
  const [description, setDescription] = useState('')
  const [photo, setPhoto] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [photoError, setPhotoError] = useState<string | null>(null)
  const [dragging, setDragging] = useState(false)
  const [fieldError, setFieldError] = useState<{ field: FieldName; text: string } | null>(null)
  const [sendProblem, setSendProblem] = useState<Problem | null>(null)
  const [busy, setBusy] = useState(false)
  const [sentId, setSentId] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  // Load the categories once.
  useEffect(() => {
    let alive = true
    api
      .get<{ categories?: CategoryInfo[] }>('/categories')
      .then((res) => {
        if (!alive) return
        const list = (res.categories ?? [])
          .filter((c) => c.active !== false)
          .sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0))
        setCategories(list)
      })
      .catch((err) => alive && setLoadProblem(problemFor(err, 'Could not load the categories. Please try again.')))
    return () => {
      alive = false
    }
  }, [])

  // A new category starts the location from scratch.
  function chooseCategory(next: string) {
    setCategory(next)
    setLocations([])
    setBuilding('')
    setWing('')
    setArea('')
  }

  // Load the locations whenever the category changes.
  useEffect(() => {
    if (!category) return
    let alive = true
    api
      .get<{ locations?: LocationEntry[] }>(`/categories/${category}/locations`)
      .then((res) => alive && setLocations(res.locations ?? []))
      .catch((err) => alive && setLoadProblem(problemFor(err, 'Could not load the locations. Please try again.')))
    return () => {
      alive = false
    }
  }, [category])

  // Free the preview URL when the photo changes or the page closes.
  useEffect(() => {
    return () => {
      if (preview) URL.revokeObjectURL(preview)
    }
  }, [preview])

  const buildings = useMemo(() => unique(locations.map((l) => l.location.building)), [locations])
  const inBuilding = locations.filter((l) => l.location.building === building)
  const wings = unique(inBuilding.map((l) => l.location.wing))
  const inWing = wings.length ? inBuilding.filter((l) => l.location.wing === wing) : inBuilding
  const areas = unique(inWing.map((l) => l.location.area))
  const chosen = categories.find((c) => c.category === category)

  // The chosen location is complete when every step that exists has a value.
  const locationDone = !!building && (!wings.length || !!wing) && (!areas.length || !!area)
  const stepsDone = [!!category, locationDone, !!description.trim()].filter(Boolean).length

  // errorText is a field's message, or null.
  function errorText(field: FieldName): string | null {
    return fieldError?.field === field ? fieldError.text : null
  }

  // The location message goes under the first step still missing.
  const locationError = errorText('location')
  const missing = !building ? 'building' : wings.length && !wing ? 'wing' : 'area'

  // acceptFile checks a chosen or dropped photo the same way.
  function acceptFile(file: File | null): boolean {
    setPhotoError(null)
    setPhoto(null)
    setPreview(null)
    if (!file) return true
    if (!PHOTO_TYPES.includes(file.type)) {
      setPhotoError('The photo must be a JPG or PNG image.')
      return false
    }
    if (file.size > MAX_PHOTO_BYTES) {
      setPhotoError('The photo must be 5 MB or smaller.')
      return false
    }
    setPhoto(file)
    setPreview(URL.createObjectURL(file))
    return true
  }

  function onPhoto(e: ChangeEvent<HTMLInputElement>) {
    if (!acceptFile(e.target.files?.[0] ?? null)) e.target.value = ''
  }

  function onDragOver(e: DragEvent<HTMLLabelElement>) {
    e.preventDefault()
    setDragging(true)
  }

  function onDrop(e: DragEvent<HTMLLabelElement>) {
    e.preventDefault()
    setDragging(false)
    acceptFile(e.dataTransfer.files?.[0] ?? null)
  }

  function removePhoto() {
    setPhoto(null)
    setPreview(null)
  }

  function startAgain() {
    chooseCategory('')
    setDescription('')
    removePhoto()
    setPhotoError(null)
    setFieldError(null)
    setSendProblem(null)
    setSentId(null)
    setCopied(false)
  }

  // The pilot showed students cannot remember the 36-character number, so it
  // can be copied. The clipboard needs HTTPS; without it the button is hidden.
  async function copyNumber() {
    if (!sentId) return
    try {
      await navigator.clipboard.writeText(sentId)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (busy) return
    setFieldError(null)
    setSendProblem(null)
    if (!category) return setFieldError({ field: 'category', text: 'Choose a category.' })
    if (!locationDone) {
      return setFieldError({
        field: 'location',
        text: wings.length ? 'Choose the full location: building, floor and area.' : 'Choose the full location: building and area.',
      })
    }
    const text = description.trim()
    if (!text) return setFieldError({ field: 'description', text: 'Describe the problem.' })

    setBusy(true)
    try {
      const body: Record<string, unknown> = {
        category,
        location: { building, ...(wing && { wing }), ...(area && { area }) },
        description: text,
      }
      // Upload first: the report refers to the photo by its id.
      if (photo) {
        const uploaded = await api.upload<{ photoId: string }>('/photos', photo, 'photo')
        body.photoId = uploaded.photoId
      }
      const res = await api.post<{ report?: { id?: string } }>('/reports', body)
      setSentId(res.report?.id ?? '')
    } catch (err) {
      setSendProblem(problemFor(err, ''))
    } finally {
      setBusy(false)
    }
  }

  if (sentId !== null) {
    const canCopy = !!sentId && typeof navigator !== 'undefined' && !!navigator.clipboard
    return (
      <div className="report-page">
        <PageHeader title="Report an issue" />
        <section className="card report-done">
          <span className="report-done-icon">
            <Icon name="resolved" size={36} />
          </span>
          <p role="status" className="report-done-text">
            Report sent. Your report number is <code className="report-number">{sentId}</code>.
          </p>
          <StatusBadge status="OPEN" />
          {canCopy && (
            <Button variant="secondary" onClick={copyNumber}>
              <FormIcon name={copied ? 'check' : 'copy'} size={18} />
              {copied ? 'Copied' : 'Copy report number'}
            </Button>
          )}
          <p className="muted">It goes to the Category Head for this category. Keep the number to ask about it later.</p>
          <div className="report-actions">
            <Button onClick={startAgain}>
              <Icon name="report" />
              Report another issue
            </Button>
            <Link to="/" className="report-secondary">
              <Icon name="home" />
              Back to home
            </Link>
          </div>
        </section>
      </div>
    )
  }

  const categoryError = errorText('category')

  return (
    <div className="report-page">
      <PageHeader title="Report an issue" description="Tell us what is wrong and where. It takes about two minutes." />
      {loadProblem && <ProblemText problem={loadProblem} />}

      <div className="report-layout">
        <form className="card report-form" onSubmit={onSubmit} noValidate>
          <Step icon="category" title="What kind of problem?">
            <p id="category-label" className="report-label">
              Category
            </p>
            <div
              role="radiogroup"
              aria-labelledby="category-label"
              aria-invalid={categoryError ? true : undefined}
              aria-describedby={categoryError ? 'category-error' : undefined}
              className="report-tiles"
            >
              {categories.map((c) => (
                <label key={c.category} className={c.category === category ? 'report-tile is-chosen' : 'report-tile'}>
                  <input
                    type="radio"
                    name="category"
                    className="visually-hidden"
                    value={c.category}
                    checked={c.category === category}
                    onChange={() => chooseCategory(c.category)}
                  />
                  <span className="report-tile-icon">
                    <FormIcon name={categoryIcons[c.category] ?? 'category'} size={26} />
                  </span>
                  <span className="report-tile-name">{c.name}</span>
                </label>
              ))}
            </div>
            {chosen?.description && <p className="field-hint">{chosen.description}</p>}
            {categoryError && (
              <p id="category-error" role="alert" className="field-error">
                {categoryError}
              </p>
            )}
          </Step>

          <Step icon="place" title="Where is it?">
            {!category && <p className="field-hint">Choose a category first.</p>}
            <div className="report-place">
              {category && buildings.length > 0 && (
                <Field label="Building" error={missing === 'building' ? locationError : null}>
                  <select
                    value={building}
                    onChange={(e) => {
                      setBuilding(e.target.value)
                      setWing('')
                      setArea('')
                    }}
                  >
                    <option value="">Choose a building</option>
                    {buildings.map((b) => (
                      <option key={b} value={b}>
                        {b}
                      </option>
                    ))}
                  </select>
                </Field>
              )}

              {building && wings.length > 0 && (
                // Pilot (SCRUM-113): students found the wing step confusing. Hostels
                // keep the floor in the wing field (SCRUM-150), so it says Floor.
                <Field
                  label="Floor"
                  hint="The floor the problem is on. Room 101 is on Floor 1, Room 201 on Floor 2."
                  error={missing === 'wing' ? locationError : null}
                >
                  <select
                    value={wing}
                    onChange={(e) => {
                      setWing(e.target.value)
                      setArea('')
                    }}
                  >
                    <option value="">Choose a floor</option>
                    {wings.map((w) => (
                      <option key={w} value={w}>
                        {w}
                      </option>
                    ))}
                  </select>
                </Field>
              )}

              {building && (!wings.length || wing) && areas.length > 0 && (
                <Field label="Area" error={missing === 'area' ? locationError : null}>
                  <select value={area} onChange={(e) => setArea(e.target.value)}>
                    <option value="">Choose an area</option>
                    {areas.map((a) => (
                      <option key={a} value={a}>
                        {a}
                      </option>
                    ))}
                  </select>
                </Field>
              )}
            </div>
            {/* No select on screen yet to put the message under. */}
            {locationError && !buildings.length && (
              <p role="alert" className="field-error">
                {locationError}
              </p>
            )}
          </Step>

          <Step icon="describe" title="What is wrong?">
            <Field
              label="Description"
              error={errorText('description')}
              after={
                <p
                  id="description-count"
                  className={description.length >= MAX_DESCRIPTION ? 'field-counter over' : 'field-counter'}
                >
                  {description.length} / {MAX_DESCRIPTION}
                </p>
              }
            >
              <textarea
                rows={5}
                maxLength={MAX_DESCRIPTION}
                value={description}
                placeholder="For example: the tap in Bathroom 1 has been leaking since Monday."
                onChange={(e) => setDescription(e.target.value)}
              />
            </Field>
          </Step>

          <Step icon="camera" title="Add a photo">
            {!preview && (
              <label
                className={dragging ? 'report-photo-drop is-dragging' : 'report-photo-drop'}
                onDragOver={onDragOver}
                onDragLeave={() => setDragging(false)}
                onDrop={onDrop}
              >
                <span className="report-photo-drop-icon">
                  <FormIcon name="upload" size={24} />
                </span>
                Photo (optional)
                <input type="file" accept="image/jpeg,image/png" onChange={onPhoto} aria-describedby="photo-help" />
              </label>
            )}
            <p id="photo-help" className="field-hint">
              Drag a photo here or choose one. JPG or PNG, at most 5 MB.
            </p>
            {photoError && <p className="field-error">{photoError}</p>}
            {preview && photo && (
              <div className="report-photo-preview">
                <img src={preview} alt="Photo preview" />
                <div>
                  <p className="report-photo-name">
                    {photo.name}, {fileSize(photo.size)}
                  </p>
                  <Button variant="secondary" className="report-remove" onClick={removePhoto}>
                    <FormIcon name="remove" size={18} />
                    Remove photo
                  </Button>
                </div>
              </div>
            )}
          </Step>

          {sendProblem && <ProblemText problem={sendProblem} />}

          {/* Stays at the bottom of the screen while the form scrolls. */}
          <div className="report-submit-bar">
            <p className="report-progress" aria-live="polite">
              {stepsDone === 3 ? 'Ready to send' : `${stepsDone} of 3 steps done`}
            </p>
            <Button type="submit" loading={busy} loadingText="Sending...">
              <FormIcon name="send" />
              Send report
            </Button>
          </div>
        </form>

        <NextSteps />
      </div>
    </div>
  )
}
