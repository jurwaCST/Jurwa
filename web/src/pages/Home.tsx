import { Link } from 'react-router-dom'
import type { Role } from '../auth/types'
import { useAuth } from '../auth/useAuth'

// The page each role starts from. A user can hold more than one role.
const linksByRole: { role: Role; to: string; label: string }[] = [
  { role: 'STUDENT', to: '/report/new', label: 'Report an issue' },
  { role: 'CATEGORY_HEAD', to: '/staff', label: 'Staff view' },
  { role: 'STAFF', to: '/staff', label: 'Staff view' },
  { role: 'SYSTEM_ADMIN', to: '/admin', label: 'Admin dashboard' },
  { role: 'PRESIDENT', to: '/president', label: 'President dashboard' },
]

export default function Home() {
  const { user } = useAuth()

  // Category Head and Staff share one page, so drop the repeated link.
  const links = user
    ? linksByRole
        .filter((l) => user.roles.includes(l.role))
        .filter((l, i, all) => all.findIndex((x) => x.to === l.to) === i)
    : []

  return (
    <>
      <h1>Jurwa</h1>
      <p className="muted">Report campus facility issues at CST and follow them until they are fixed.</p>
      {user ? (
        <nav aria-label="Your pages">
          <ul>
            {links.map((l) => (
              <li key={l.to}>
                <Link to={l.to}>{l.label}</Link>
              </li>
            ))}
          </ul>
        </nav>
      ) : (
        <p>
          <Link to="/login">Sign in</Link>
        </p>
      )}
    </>
  )
}
