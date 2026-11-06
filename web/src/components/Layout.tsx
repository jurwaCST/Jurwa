import { Link, Outlet } from 'react-router-dom'
import { useAuth } from '../auth/useAuth'

/** Layout is the frame around every page: the app name and who is signed in. */
export default function Layout() {
  const { user, logout } = useAuth()

  return (
    <div className="layout">
      <header>
        <Link to="/" className="brand">
          Jurwa
        </Link>
        {user && (
          <div>
            <span className="muted">Signed in as {user.fullName}</span>{' '}
            <button type="button" className="link" onClick={logout}>
              Sign out
            </button>
          </div>
        )}
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  )
}
