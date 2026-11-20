import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import type { Role } from '../auth/types'
import { useAuth } from '../auth/useAuth'
import Icon, { type IconName } from './Icon'
import { roleNames } from './labels'
import NotificationBell from './NotificationBell'
import './components.css'

interface Tab {
  to: string
  label: string
  icon: IconName
}

// Tabs per role. Only pages that exist are listed (routes in App.tsx).
const tabsByRole: Record<Role, Tab[]> = {
  STUDENT: [
    { to: '/report/new', label: 'Report', icon: 'report' },
    { to: '/reports', label: 'My reports', icon: 'queue' },
  ],
  CATEGORY_HEAD: [{ to: '/staff', label: 'Queue', icon: 'queue' }],
  STAFF: [{ to: '/staff', label: 'Queue', icon: 'queue' }],
  SYSTEM_ADMIN: [{ to: '/admin', label: 'Overview', icon: 'admin' }],
  PRESIDENT: [{ to: '/president', label: 'Dashboard', icon: 'dashboard' }],
}

function tabsFor(roles: Role[]): Tab[] {
  const tabs: Tab[] = [{ to: '/', label: 'Home', icon: 'home' }]
  for (const r of roles) {
    for (const t of tabsByRole[r] ?? []) {
      if (!tabs.some((x) => x.to === t.to)) tabs.push(t)
    }
  }
  return tabs
}

// initials takes the first letters of the capitalised words, so "Dean of Student
// Affairs" is "DS" and "Block Councillor (HA)" is "BC". A name in lower case
// falls back to its first words.
function initials(name: string): string {
  const words = name.split(/\s+/).filter((w) => /^[A-Za-z]/.test(w))
  const capital = words.filter((w) => /^[A-Z]/.test(w))
  return (capital.length ? capital : words)
    .slice(0, 2)
    .map((w) => w[0]!.toUpperCase())
    .join('')
}

// The page each role starts on. It has no back button: there is nothing before it.
const startPages = new Set(['/', '/staff', '/admin', '/president'])

// BackButton goes to the page before (SCRUM-159): the installed app on a phone has
// no browser back button. Opened straight from a link, it goes home.
function BackButton() {
  const navigate = useNavigate()
  const location = useLocation()
  if (startPages.has(location.pathname)) return null
  return (
    <button
      type="button"
      className="back-button"
      onClick={() => (location.key === 'default' ? navigate('/') : navigate(-1))}
    >
      <Icon name="back" size={18} />
      <span>Back</span>
    </button>
  )
}

function TabLinks({ tabs }: { tabs: Tab[] }) {
  return (
    <ul>
      {tabs.map((t) => (
        <li key={t.to}>
          <NavLink to={t.to} end={t.to === '/'}>
            <Icon name={t.icon} size={22} />
            <span>{t.label}</span>
          </NavLink>
        </li>
      ))}
    </ul>
  )
}

/** Layout is the frame around every page: bar, navigation, content, footer. */
export default function Layout() {
  const { user, logout } = useAuth()
  const tabs = user ? tabsFor(user.roles) : []

  return (
    <>
      <a href="#main" className="skip-link">
        Skip to content
      </a>
      <header className="topbar">
        <div className="topbar-inner">
          <Link to="/" className="brand">
            <img
              src="/cst-logo.jpg"
              alt=""
              className="brand-logo"
              width={36}
              height={36}
              onError={(e) => (e.currentTarget.style.display = 'none')}
            />
            <span className="brand-name">Jurwa</span>{' '}
            <span className="brand-place">College of Science and Technology</span>
          </Link>
          {user && (
            <nav className="topnav" aria-label="Main">
              <TabLinks tabs={tabs} />
            </nav>
          )}
          {user && (
            <div className="profile">
              {/* SCRUM-162: the bell is for every role. */}
              <NotificationBell userId={user.id} />
              <span className="avatar" aria-hidden="true">
                {initials(user.fullName)}
              </span>
              <span className="profile-text">
                <span className="profile-name">Signed in as {user.fullName}</span>
                {/* SCRUM-159: an HoD or councillor is named by position, so "Staff" adds nothing. */}
                {user.roles.some((r) => r !== 'STAFF') && (
                  <span className="profile-role">
                    {user.roles
                      .filter((r) => r !== 'STAFF')
                      .map((r) => roleNames[r] ?? r)
                      .join(', ')}
                  </span>
                )}
              </span>
              <button type="button" className="signout" onClick={logout}>
                <Icon name="signout" size={18} />
                <span>Sign out</span>
              </button>
            </div>
          )}
        </div>
      </header>

      <main id="main" className="page" tabIndex={-1}>
        <BackButton />
        <Outlet />
      </main>

      <footer className="footer">
        <div className="footer-inner">
          <p>
            <strong>Jurwa</strong>, campus facility reporting for the College of Science and Technology, Royal
            University of Bhutan.
          </p>
          <p className="muted">Group 6 project for SWE302, SWE303 and WEB303.</p>
        </div>
      </footer>

      {user && (
        <nav className="tabbar" aria-label="Main">
          <TabLinks tabs={tabs} />
        </nav>
      )}
    </>
  )
}
