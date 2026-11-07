// Student report form (SCRUM-35). Owner: Nyendrak.
// Category and location come from the gateway; the photo is uploaded first and
// its photoId is sent with the report (report-submission SubmitReport).
// Every field stays inside its <label> with the same text, so tests (Vitest and
// Gayley's Cypress findLabel) find it by its label.
import { useEffect, useMemo, useState, type ChangeEvent, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
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
type Field = 'category' | 'location' | 'description'

const EXPIRED: Problem = { text: 'Your sign-in has expired. Please sign in again.', signIn: true }

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
      {problem.text}
      {problem.signIn && (
        <>
          {' '}
          <Link to="/login">Sign in again</Link>
        </>
      )}
    </p>
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
  const [fieldError, setFieldError] = useState<{ field: Field; text: string } | null>(null)
  const [sendProblem, setSendProblem] = useState<Problem | null>(null)
  const [busy, setBusy] = useState(false)
  const [sentId, setSentId] = useState<string | null>(null)

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

  // errorFor gives a field's error props: aria-invalid and the id of its message.
  function errorFor(field: Field, describedBy?: string) {
    const invalid = fieldError?.field === field
    const ids = [describedBy, invalid ? `${field}-error` : undefined].filter(Boolean).join(' ')
    return { 'aria-invalid': invalid || undefined, 'aria-describedby': ids || undefined }
  }

  // fieldErrorText shows a field's message under it, or nothing.
  function fieldErrorText(field: Field) {
    if (fieldError?.field !== field) return null
    return (
      <p id={`${field}-error`} role="alert" className="report-field-error">
        {fieldError.text}
      </p>
    )
  }

  function onPhoto(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0] ?? null
    setPhotoError(null)
    setPhoto(null)
    setPreview(null)
    if (!file) return
    if (!PHOTO_TYPES.includes(file.type)) {
      setPhotoError('The photo must be a JPG or PNG image.')
      e.target.value = ''
      return
    }
    if (file.size > MAX_PHOTO_BYTES) {
      setPhotoError('The photo must be 5 MB or smaller.')
      e.target.value = ''
      return
    }
    setPhoto(file)
    setPreview(URL.createObjectURL(file))
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
        text: wings.length ? 'Choose the full location: building, wing and area.' : 'Choose the full location: building and area.',
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
    return (
      <div className="report-page">
        <h1>Report an issue</h1>
        <section className="report-card report-done">
          <div className="report-done-icon" aria-hidden="true">
            ✓
          </div>
          <p role="status" className="report-done-text">
            Report sent. Your report number is <code className="report-number">{sentId}</code>.
          </p>
          <p className="muted">The right Category Head has received it. You can follow it under your reports.</p>
          <div className="report-actions">
            <button type="button" onClick={startAgain}>
              Report another issue
            </button>
            <Link to="/" className="report-secondary">
              Back to home
            </Link>
          </div>
        </section>
      </div>
    )
  }

  return (
    <div className="report-page">
      <h1>Report an issue</h1>
      <p className="muted report-intro">Tell us what is wrong and where. It takes about a minute.</p>
      {loadProblem && <ProblemText problem={loadProblem} />}

      <form className="report-card report-form" onSubmit={onSubmit} noValidate>
        <section className="report-section">
          <h2 className="report-step">
            <span aria-hidden="true">1</span> What kind of problem?
          </h2>
          <label>
            Category
            <select value={category} onChange={(e) => chooseCategory(e.target.value)} {...errorFor('category')}>
              <option value="">Choose a category</option>
              {categories.map((c) => (
                <option key={c.category} value={c.category}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          {chosen?.description && <p className="muted report-hint">{chosen.description}</p>}
          {fieldErrorText('category')}
        </section>

        <section className="report-section">
          <h2 className="report-step">
            <span aria-hidden="true">2</span> Where is it?
          </h2>
          {!category && <p className="muted report-hint">Choose a category first.</p>}
          {category && buildings.length > 0 && (
            <label>
              Building
              <select
                value={building}
                onChange={(e) => {
                  setBuilding(e.target.value)
                  setWing('')
                  setArea('')
                }}
                {...errorFor('location')}
              >
                <option value="">Choose a building</option>
                {buildings.map((b) => (
                  <option key={b} value={b}>
                    {b}
                  </option>
                ))}
              </select>
            </label>
          )}

          {building && wings.length > 0 && (
            <label>
              Wing
              <select
                value={wing}
                onChange={(e) => {
                  setWing(e.target.value)
                  setArea('')
                }}
              >
                <option value="">Choose a wing</option>
                {wings.map((w) => (
                  <option key={w} value={w}>
                    {w}
                  </option>
                ))}
              </select>
            </label>
          )}

          {building && (!wings.length || wing) && areas.length > 0 && (
            <label>
              Area
              <select value={area} onChange={(e) => setArea(e.target.value)}>
                <option value="">Choose an area</option>
                {areas.map((a) => (
                  <option key={a} value={a}>
                    {a}
                  </option>
                ))}
              </select>
            </label>
          )}
          {fieldErrorText('location')}
        </section>

        <section className="report-section">
          <h2 className="report-step">
            <span aria-hidden="true">3</span> What is wrong?
          </h2>
          <label>
            Description
            <textarea
              rows={5}
              maxLength={MAX_DESCRIPTION}
              value={description}
              placeholder="For example: the tap in Bathroom 1 has been leaking since Monday."
              onChange={(e) => setDescription(e.target.value)}
              {...errorFor('description', 'description-count')}
            />
          </label>
          <p
            id="description-count"
            className={`muted report-counter${description.length >= MAX_DESCRIPTION ? ' report-counter-full' : ''}`}
          >
            {description.length} / {MAX_DESCRIPTION}
          </p>
          {fieldErrorText('description')}
        </section>

        <section className="report-section">
          <h2 className="report-step">
            <span aria-hidden="true">4</span> Add a photo
          </h2>
          {!preview && (
            <label className="report-photo-drop">
              Photo (optional)
              <input type="file" accept="image/jpeg,image/png" onChange={onPhoto} aria-describedby="photo-help" />
            </label>
          )}
          <p id="photo-help" className="muted report-hint">
            JPG or PNG, at most 5 MB.
          </p>
          {photoError && <p className="report-field-error">{photoError}</p>}
          {preview && photo && (
            <div className="report-photo-preview">
              <img src={preview} alt="Photo preview" />
              <div>
                <p className="report-photo-name">
                  {photo.name}, {fileSize(photo.size)}
                </p>
                <button type="button" className="link" onClick={removePhoto}>
                  Remove photo
                </button>
              </div>
            </div>
          )}
        </section>

        {sendProblem && <ProblemText problem={sendProblem} />}
        <button type="submit" className="report-submit" disabled={busy}>
          {busy ? 'Sending...' : 'Send report'}
        </button>
      </form>
    </div>
  )
}
