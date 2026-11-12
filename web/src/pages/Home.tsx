import { Link } from 'react-router-dom'
import type { Role } from '../auth/types'
import { useAuth } from '../auth/useAuth'
import Icon, { type IconName } from '../components/Icon'
import { roleNames } from '../components/labels'
import './Home.css'

// The page each role starts from. A user can hold more than one role.
const actionsByRole: { role: Role; to: string; label: string; hint: string; icon: IconName }[] = [
  { role: 'STUDENT', to: '/report/new', label: 'Report an issue', hint: 'Takes about two minutes', icon: 'report' },
  { role: 'CATEGORY_HEAD', to: '/staff', label: 'Staff view', hint: 'Reports assigned to you', icon: 'queue' },
  { role: 'STAFF', to: '/staff', label: 'Staff view', hint: 'Reports assigned to you', icon: 'queue' },
  { role: 'SYSTEM_ADMIN', to: '/admin', label: 'Admin dashboard', hint: 'Rules, time limits and counts', icon: 'admin' },
  { role: 'PRESIDENT', to: '/president', label: 'President dashboard', hint: 'Read-only overview', icon: 'dashboard' },
]

// The route a report travels. A real sequence, so it is drawn as one.
const stops: { name: string; icon: IconName; tone: string; text: string }[] = [
  { name: 'Open', icon: 'open', tone: 'open', text: 'You choose the category and the place, describe it and add a photo if you have one.' },
  { name: 'Assigned', icon: 'assigned', tone: 'assigned', text: 'It goes straight to the head for its category, by category and place only.' },
  { name: 'In progress', icon: 'progress', tone: 'progress', text: 'The Category Head or their staff take it on.' },
  { name: 'Resolved', icon: 'resolved', tone: 'resolved', text: 'It is fixed and marked resolved.' },
]

function Route() {
  return (
    <section className="route" aria-labelledby="route-title">
      <h2 id="route-title">Where your report goes</h2>
      <ol>
        {stops.map((s) => (
          <li key={s.name} className={`stop stop-${s.tone}`}>
            <span className="stop-dot">
              <Icon name={s.icon} size={20} />
            </span>
            <strong>{s.name}</strong>
            <p>{s.text}</p>
          </li>
        ))}
      </ol>
      <p className="route-escalation">
        <strong>Nobody acting on it?</strong> Every category has a time limit. When it passes, the report is escalated
        to the Category Head automatically.
      </p>
    </section>
  )
}

// The three categories and who receives them (FR-12), with photos of the
// places on the CST campus they cover.
const categories = [
  {
    name: 'Hostel',
    head: 'DSA',
    covers: 'Rooms, bathrooms, water and electricity in the hostels',
    photo: '/campus/hostels.jpg',
    alt: 'Red-roofed hostel blocks on the CST campus',
  },
  {
    name: 'Academic Block',
    head: 'DAA',
    covers: 'Classrooms, labs and equipment in the academic blocks',
    photo: '/campus/academic-block.jpg',
    alt: 'The yellow and red academic block beside the lamp-lined road',
  },
  {
    name: 'Others',
    head: 'SSO',
    covers: 'Campus places such as the library, mess, gym and sports grounds',
    photo: '/campus/aerial.jpg',
    alt: 'The CST campus on its hillside from above, with the football ground and winding roads',
  },
]

function Categories() {
  return (
    <section className="places" aria-labelledby="places-title">
      <h2 id="places-title">Who fixes what</h2>
      <ul>
        {categories.map((c) => (
          <li key={c.name} className="place">
            <img src={c.photo} alt={c.alt} loading="lazy" />
            <div className="place-text">
              <h3>{c.name}</h3>
              <p>{c.covers}.</p>
              <p className="place-head">
                Goes to the <strong>{c.head}</strong>
              </p>
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}

export default function Home() {
  const { user } = useAuth()

  if (!user) {
    return (
      <>
        <section className="hero">
          <img
            className="hero-photo"
            src="/campus/cst-roof.jpg"
            alt="The CST hall in Rinchending, with CST spelled out in solar panels on its teal roof"
            fetchPriority="high"
          />
          <div className="hero-text">
            <h1>See something broken on campus?</h1>
            <p className="hero-lead">
              A leaking tap, a dead light, a broken window. Report it in under two minutes, and Jurwa sends it to the
              person who can fix it.
            </p>
            <div className="hero-actions">
              <Link to="/login" className="action">
                Sign in
              </Link>
              <Link to="/register" className="hero-link">
                Create a student account
              </Link>
            </div>
          </div>
        </section>

        <Categories />
        <Route />
      </>
    )
  }

  // Category Head and Staff share one page, so drop the repeated action.
  const actions = actionsByRole
    .filter((a) => user.roles.includes(a.role))
    .filter((a, i, all) => all.findIndex((x) => x.to === a.to) === i)
  const firstName = user.fullName.split(' ')[0]

  return (
    <>
      <section className="greeting">
        <h1>Kuzuzangpo, {firstName}</h1>
        <p className="role-chips">
          {user.roles.map((r) => (
            <span key={r} className="role-chip">
              {roleNames[r] ?? r}
            </span>
          ))}
        </p>
      </section>

      <nav aria-label="Your pages">
        <ul className="tiles">
          {actions.map((a, i) => (
            // Only the title is the link, so its name is exactly the words it shows
            // (axe label-content-name-mismatch). The hint is its description, and
            // the link's ::after covers the tile so the whole tile is clickable.
            <li key={a.to} className={i === 0 ? 'tile tile-main' : 'tile'}>
              <span className="tile-icon">
                <Icon name={a.icon} size={26} />
              </span>
              <Link to={a.to} className="tile-label" aria-describedby={`tile-${i}-hint`}>
                {a.label}
              </Link>
              <span className="tile-hint" id={`tile-${i}-hint`}>
                {a.hint}
              </span>
            </li>
          ))}
        </ul>
      </nav>

      <Route />
    </>
  )
}
