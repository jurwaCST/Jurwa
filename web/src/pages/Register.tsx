import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { useAuth } from '../auth/useAuth'
import AuthLayout from '../components/AuthLayout'
import PasswordField from '../components/PasswordField'
import './Auth.css'

// identity's rule (services/identity/domain/user.go): at least 8 characters.
const MIN_PASSWORD = 8

/**
 * Register creates a Student account (POST /auth/register always gives the
 * Student role), then signs in with it. Staff accounts are made by the system.
 */
export default function Register() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    if (!fullName.trim()) {
      setError('Enter your full name.')
      return
    }
    if ([...password].length < MIN_PASSWORD) {
      setError(`Choose a password of at least ${MIN_PASSWORD} characters.`)
      return
    }
    setBusy(true)
    try {
      await api.post('/auth/register', { email, fullName: fullName.trim(), password })
    } catch (err) {
      // The gateway's own answers (400 bad email, 409 already registered) are
      // sentences for people; anything else gets a plain message.
      const known = err instanceof ApiError && (err.status === 400 || err.status === 409)
      setError(known ? err.message : 'Could not create the account. Try again.')
      setBusy(false)
      return
    }
    try {
      await login(email, password)
      navigate('/', { replace: true })
    } catch {
      // The account exists; only the automatic sign-in failed.
      navigate('/login', { replace: true })
    }
  }

  return (
    <AuthLayout title="Create a student account" intro="Use your college email. It takes a minute.">
      <form className="stack" onSubmit={onSubmit} noValidate>
        <label>
          Full name
          <input type="text" autoComplete="name" required value={fullName} onChange={(e) => setFullName(e.target.value)} />
        </label>
        <label>
          Email
          <input type="email" autoComplete="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        </label>
        <PasswordField
          value={password}
          onChange={setPassword}
          autoComplete="new-password"
          hint={`At least ${MIN_PASSWORD} characters.`}
        />
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        <button type="submit" className="auth-submit" disabled={busy}>
          {busy ? 'Creating account...' : 'Create account'}
        </button>
      </form>

      <p className="auth-switch">
        Already have an account? <Link to="/login">Sign in</Link>
      </p>
      <p className="muted auth-note">
        This creates a student account. Category Heads, staff and administrators get their accounts from the System
        Administrator.
      </p>
    </AuthLayout>
  )
}
