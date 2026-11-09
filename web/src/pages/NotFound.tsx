import { Link } from 'react-router-dom'
import { LostArt } from '../components/Art'

/** NotFound is the 404 page: the address does not match any page. */
export default function NotFound() {
  return (
    <section className="empty notice-page">
      <div className="empty-art">
        <LostArt />
      </div>
      <h1>Page not found</h1>
      <p>This address is not a page in Jurwa. Check it, or start again from the home page.</p>
      <Link to="/" className="action">
        Go to the home page
      </Link>
    </section>
  )
}
