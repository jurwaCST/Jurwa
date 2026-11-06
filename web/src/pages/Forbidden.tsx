import { Link } from 'react-router-dom'

/** Forbidden is the 403 page: signed in, but the role may not open this page. */
export default function Forbidden() {
  return (
    <>
      <h1>Not allowed</h1>
      <p>Your role cannot open this page.</p>
      <p>
        <Link to="/">Go to the home page</Link>
      </p>
    </>
  )
}
