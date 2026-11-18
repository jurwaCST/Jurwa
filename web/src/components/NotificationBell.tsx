// The bell in the header for students (SCRUM-144). Owner: Gayley.
// It reads the newest message (GET /me/notifications?pageSize=1) on every page
// change and shows a dot when it is newer than the last one seen. On the
// notifications page itself there is no dot.
import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { api } from '../api/client'
import { isNewer, lastSeen, SEEN_EVENT } from '../notifications/seen'
import Icon from './Icon'

type Page = { notifications?: { createdAt: string }[] }

export default function NotificationBell({ userId }: { userId: string }) {
  const { pathname } = useLocation()
  const [newest, setNewest] = useState<string | null>(null)
  const [seen, setSeen] = useState(() => lastSeen(userId))

  useEffect(() => {
    let alive = true
    api
      .get<Page>('/me/notifications?pageSize=1')
      .then((res) => alive && setNewest(res.notifications?.[0]?.createdAt ?? null))
      .catch(() => {
        // No dot when the list cannot be read; the page shows the error.
      })
    return () => {
      alive = false
    }
  }, [pathname])

  useEffect(() => {
    const update = () => setSeen(lastSeen(userId))
    window.addEventListener(SEEN_EVENT, update)
    return () => window.removeEventListener(SEEN_EVENT, update)
  }, [userId])

  const unread = newest !== null && pathname !== '/notifications' && isNewer(newest, seen)

  return (
    <Link
      to="/notifications"
      className="bell"
      aria-label={unread ? 'Notifications, new' : 'Notifications'}
      aria-current={pathname === '/notifications' ? 'page' : undefined}
    >
      <Icon name="bell" size={22} />
      {unread && <span className="bell-dot" aria-hidden="true" />}
    </Link>
  )
}
