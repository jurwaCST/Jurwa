import { useEffect, useState } from 'react'
import { api } from '../api/client'

// RatingTile shows the students' average rating on the admin and President
// dashboards (SCRUM-164), from GET /dashboard/ratings, refreshed every 15
// seconds like the counts. The count is int64, so protojson sends a string,
// and leaves it out when it is 0. If it cannot load, the tile is not shown.

type Ratings = { count?: string; averageRating?: number }

const REFRESH_MS = 15_000

export default function RatingTile() {
  const [ratings, setRatings] = useState<Ratings | null>(null)

  useEffect(() => {
    let alive = true
    const load = () =>
      api
        .get<Ratings>('/dashboard/ratings')
        .then((r) => alive && setRatings(r))
        .catch(() => {})
    load()
    const timer = setInterval(load, REFRESH_MS)
    return () => {
      alive = false
      clearInterval(timer)
    }
  }, [])

  if (!ratings) return null
  const count = Number(ratings.count ?? 0)
  return (
    <section className="card rating-tile" aria-label="Student ratings">
      <h2>Student ratings</h2>
      {count === 0 ? (
        <p className="muted">No ratings yet.</p>
      ) : (
        <>
          <p className="rating-value">{`${(ratings.averageRating ?? 0).toFixed(1)} / 5`}</p>
          <p className="muted">{`${count} rated ${count === 1 ? 'report' : 'reports'}`}</p>
        </>
      )}
    </section>
  )
}
