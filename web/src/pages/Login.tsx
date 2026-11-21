import { useState, type FormEvent } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { ApiError } from '../api/client'
import { useAuth } from '../auth/useAuth'
import AuthLayout from '../components/AuthLayout'
import PasswordField from '../components/PasswordField'
import './Auth.css'

// The demo accounts from the identity seed (SCRUM-126). Their password is
// never written in the app; the team gives it out for the demo.
const demoAccounts = [
  // SCRUM-158: Hostel and Others go to the SSO, who hands them to a councillor;
  // the Academic Block goes to the department's HoD. Late reports go up to the
  // DSA or the DAA. Every HoD and councillor has an account like these two.
  ['Student', '02240300.cst@rub.edu.bt'],
  ['SSO, Hostel and Others', 'sso.cst@rub.edu.bt'],
  ['Block Councillor (HA)', 'ha.cst@rub.edu.bt'],
  ['HoD, Computing Technologies', 'hod.ct.cst@rub.edu.bt'],
  ['DSA, late Hostel and Others reports', 'dsa.cst@rub.edu.bt'],
  ['DAA, late Academic Block reports', 'daa.cst@rub.edu.bt'],
  ['System Administrator', 'admin.cst@rub.edu.bt'],
  ['College President', 'president.cst@rub.edu.bt'],
]

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  // RequireRole puts the page the user wanted in state.from.
  const from = (location.state as { from?: string } | null)?.from ?? '/'

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await login(email, password)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Sign in failed. Try again.')
      setBusy(false)
    }
  }

  return (
    <AuthLayout title="Sign in" intro="Use your college email and password.">
      <form className="stack" onSubmit={onSubmit}>
        <label>
          Email
          <input
            type="email"
            autoComplete="username"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </label>
        <PasswordField value={password} onChange={setPassword} autoComplete="current-password" />
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        <button type="submit" className="auth-submit" disabled={busy}>
          {busy ? 'Signing in...' : 'Sign in'}
        </button>
      </form>

      <p className="auth-switch">
        New student? <Link to="/register">Create a student account</Link>
      </p>

      <details className="demo-accounts">
        <summary>Demo accounts</summary>
        <ul>
          {demoAccounts.map(([role, mail]) => (
            <li key={mail}>
              <span className="muted">{role}</span>
              <code>{mail}</code>
            </li>
          ))}
        </ul>
        <p className="muted">Ask the team for the demo password.</p>
      </details>
    </AuthLayout>
  )
}
