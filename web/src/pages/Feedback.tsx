// Feedback stars on a resolved report (SCRUM-145, FR-10). Owner: Gayley.
// GET /reports/{id}/feedback: the saved rating, or 404 when there is none yet.
// POST /reports/{id}/feedback: 1 to 5 stars and an optional comment. The
// feedback service refuses a second rating (409) and a report that is not
// RESOLVED (422).
import { useEffect, useId, useState, type FormEvent } from 'react'
import { api, ApiError } from '../api/client'
import Button from '../components/Button'
import Icon from '../components/Icon'
import Skeleton from '../components/Skeleton'
import TextArea from '../components/TextArea'
import { messageFor } from './reportText'

type Saved = { rating: number; comment?: string }

const MAX_COMMENT = 500 // the feedback service's limit
const STARS = [1, 2, 3, 4, 5]
const starWord = (n: number) => (n === 1 ? '1 star' : `${n} stars`)

/** Five stars, filled up to the rating. Decorative: the text says the rating. */
function Stars({ rating }: { rating: number }) {
  return (
    <span className="fb-stars" aria-hidden="true">
      {STARS.map((n) => (
        <span key={n} className={n <= rating ? 'fb-star on' : 'fb-star'}>
          <Icon name="star" size={28} />
        </span>
      ))}
    </span>
  )
}

export default function Feedback({ reportId }: { reportId: string }) {
  const headingId = useId()
  const [saved, setSaved] = useState<Saved | null>(null)
  const [state, setState] = useState<'loading' | 'form' | 'saved' | 'error'>('loading')
  const [attempt, setAttempt] = useState(0)
  const [rating, setRating] = useState(0)
  const [hover, setHover] = useState(0)
  const [comment, setComment] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  function show(f: Saved) {
    setSaved(f)
    setState('saved')
  }

  useEffect(() => {
    let alive = true
    api
      .get<{ feedback: Saved }>(`/reports/${reportId}/feedback`)
      .then((res) => alive && show(res.feedback))
      .catch((err) => {
        if (!alive) return
        if (err instanceof ApiError && err.status === 404) setState('form')
        else {
          setError(messageFor(err))
          setState('error')
        }
      })
    return () => {
      alive = false
    }
  }, [reportId, attempt])

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    if (!rating) {
      setError('Choose from 1 to 5 stars.')
      return
    }
    if ([...comment].length > MAX_COMMENT) {
      setError(`Keep the comment to ${MAX_COMMENT} characters.`)
      return
    }
    setBusy(true)
    try {
      const res = await api.post<{ feedback: Saved }>(`/reports/${reportId}/feedback`, {
        rating,
        comment: comment.trim(),
      })
      show(res.feedback)
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        // Already rated, for example in another tab: show that rating.
        try {
          show((await api.get<{ feedback: Saved }>(`/reports/${reportId}/feedback`)).feedback)
        } catch (again) {
          setError(messageFor(again))
        }
      } else if (err instanceof ApiError && err.status === 422) {
        setError('You can rate a report once it is resolved.')
      } else {
        setError(messageFor(err))
      }
    } finally {
      setBusy(false)
    }
  }

  let body
  if (state === 'loading') {
    body = <Skeleton lines={2} label="Loading your feedback" />
  } else if (state === 'error') {
    body = (
      <>
        <p role="alert" className="error">
          {error}
        </p>
        <Button
          variant="secondary"
          onClick={() => {
            setError(null)
            setState('loading')
            setAttempt((n) => n + 1)
          }}
        >
          Try again
        </Button>
      </>
    )
  } else if (state === 'saved' && saved) {
    body = (
      <div className="fb-saved">
        <Stars rating={saved.rating} />
        <p>You rated this {saved.rating} out of 5 stars.</p>
        {saved.comment && <blockquote className="fb-comment">{saved.comment}</blockquote>}
        <p className="muted">Thank you. Your rating helps the Category Head see what works.</p>
      </div>
    )
  } else {
    const shown = hover || rating
    body = (
      <form className="stack" onSubmit={onSubmit} noValidate>
        <div role="radiogroup" aria-labelledby={headingId} className="fb-stars fb-pick" onMouseLeave={() => setHover(0)}>
          {STARS.map((n) => (
            <label key={n} className={n <= shown ? 'fb-star on' : 'fb-star'} onMouseEnter={() => setHover(n)}>
              <input
                type="radio"
                name={`rating-${reportId}`}
                className="visually-hidden"
                value={n}
                checked={rating === n}
                onChange={() => setRating(n)}
              />
              <Icon name="star" size={36} />
              <span className="visually-hidden">{starWord(n)}</span>
            </label>
          ))}
        </div>
        <p className="muted fb-choice" aria-hidden="true">
          {rating ? `${starWord(rating)} out of 5` : 'Tap a star.'}
        </p>
        <TextArea label="Comment (optional)" value={comment} onChange={setComment} max={MAX_COMMENT} rows={3} />
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        <div>
          <Button type="submit" loading={busy} loadingText="Sending...">
            Send feedback
          </Button>
        </div>
      </form>
    )
  }

  return (
    <section className="card fb-card" aria-labelledby={headingId}>
      <h2 id={headingId}>How was the fix?</h2>
      {body}
    </section>
  )
}
