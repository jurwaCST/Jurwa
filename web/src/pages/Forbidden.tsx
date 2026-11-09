import { Link } from 'react-router-dom'
import { useAuth } from '../auth/useAuth'
import { LostArt } from '../components/Art'
import { roleNames } from '../components/labels'

/** Forbidden is the 403 page: signed in, but the role may not open this page. */
export default function Forbidden() {
  const { user } = useAuth()
  const roles = user?.roles.map((r) => roleNames[r] ?? r).join(' and ')

  return (
    <section className="empty notice-page">
      <div className="empty-art">
        <LostArt />
      </div>
      <h1>Not allowed</h1>
      <p>
        {roles ? `This page is not for your role (${roles}).` : 'Your role cannot open this page.'} Your own pages are
        on the home page.
      </p>
      <Link to="/" className="action">
        Go to the home page
      </Link>
    </section>
  )
}
