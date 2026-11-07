// Student report form (SCRUM-35). Owner: Nyendrak.
// Category and location come from the gateway; the photo is uploaded first and
// its photoId is sent with the report (report-submission SubmitReport).
import { useEffect, useMemo, useState, type ChangeEvent, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import './ReportForm.css'

/** MAX_DESCRIPTION is report-submission's limit. */
export const MAX_DESCRIPTION = 500
/** MAX_PHOTO_BYTES is photo-media's limit: 5 MB, and exactly 5 MB is allowed. */
export const MAX_PHOTO_BYTES = 5 * 1024 * 1024
const PHOTO_TYPES = ['image/jpeg', 'image/png']

type CategoryInfo = { category: string; name: string; sortOrder?: number; active?: boolean }
type Location = { building: string; wing?: string; area?: string }
type LocationEntry = { id: string; category: string; location: Location }

// unique keeps the first copy of each non-empty value, in order.
function unique(values: (string | undefined)[]): string[] {
  return values.filter((v, i): v is string => !!v && values.indexOf(v) === i)
}

// errorText turns an API error into a message for the student.
function errorText(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 0 || err.status === 502 || err.status === 503 || err.status === 504) {
      return 'The service is busy or unreachable. Please try again.'
    }
    return err.message
  }
  return 'Something went wrong. Please try again.'
}

export default function ReportForm() {
  const [categories, setCategories] = useState<CategoryInfo[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)
  const [category, setCategory] = useState('')
  const [locations, setLocations] = useState<LocationEntry[]>([])
  const [building, setBuilding] = useState('')
  const [wing, setWing] = useState('')
  const [area, setArea] = useState('')
  const [description, setDescription] = useState('')
  const [photo, setPhoto] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [photoError, setPhotoError] = useState<string | null>(null)
  const [formError, setFormError] = useState<string | null>(null)
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
      .catch(() => alive && setLoadError('Could not load the categories. Please try again.'))
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
      .catch(() => alive && setLoadError('Could not load the locations. Please try again.'))
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

  // The chosen location is complete when every step that exists has a value.
  const locationDone = !!building && (!wings.length || !!wing) && (!areas.length || !!area)

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

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (busy) return
    setFormError(null)
    if (!category) return setFormError('Choose a category.')
    if (!locationDone) {
      return setFormError(
        wings.length ? 'Choose the full location: building, wing and area.' : 'Choose the full location: building and area.',
      )
    }
    const text = description.trim()
    if (!text) return setFormError('Describe the problem.')

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
      setFormError(errorText(err))
    } finally {
      setBusy(false)
    }
  }

  if (sentId !== null) {
    return (
      <>
        <h1>Report an issue</h1>
        <p role="status" className="report-sent">
          Report sent. Your report number is {sentId}.
        </p>
        <p>
          <Link to="/">Back to home</Link>
        </p>
      </>
    )
  }

  return (
    <>
      <h1>Report an issue</h1>
      {loadError && (
        <p role="alert" className="error">
          {loadError}
        </p>
      )}
      <form className="stack report-form" onSubmit={onSubmit} noValidate>
        <label>
          Category
          <select value={category} onChange={(e) => chooseCategory(e.target.value)}>
            <option value="">Choose a category</option>
            {categories.map((c) => (
              <option key={c.category} value={c.category}>
                {c.name}
              </option>
            ))}
          </select>
        </label>

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

        <label>
          Description
          <textarea
            rows={5}
            maxLength={MAX_DESCRIPTION}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            aria-describedby="description-count"
          />
        </label>
        <p id="description-count" className="muted counter">
          {description.length} / {MAX_DESCRIPTION}
        </p>

        <label>
          Photo (optional)
          <input type="file" accept="image/jpeg,image/png" onChange={onPhoto} aria-describedby="photo-help" />
        </label>
        <p id="photo-help" className="muted">
          JPG or PNG, at most 5 MB.
        </p>
        {photoError && <p className="error">{photoError}</p>}
        {preview && (
          <div className="photo-preview">
            <img src={preview} alt="Photo preview" />
            <button type="button" className="link" onClick={removePhoto}>
              Remove photo
            </button>
          </div>
        )}

        {formError && (
          <p role="alert" className="error">
            {formError}
          </p>
        )}
        <button type="submit" disabled={busy}>
          {busy ? 'Sending...' : 'Send report'}
        </button>
      </form>
    </>
  )
}
