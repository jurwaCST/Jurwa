import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { SESSION_KEY } from './auth/AuthContext'
import { setToken } from './api/client'

// renderAt shows the app at one URL, as if the user typed it in the address bar.
function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  )
}

// signInAs stores a session for a test user, as a real login would.
function signInAs(role: string, fullName = 'Test User') {
  sessionStorage.setItem(
    SESSION_KEY,
    JSON.stringify({
      token: 'test-token',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
      user: { id: 'u1', email: 'test@cst.edu.bt', fullName, roles: [role] },
    }),
  )
}

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn())
})

afterEach(() => {
  setToken(null)
  vi.unstubAllGlobals()
})

describe('routes', () => {
  it('home shows the app name and a sign in link', () => {
    renderAt('/')
    expect(screen.getByRole('heading', { level: 1, name: 'Jurwa' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Sign in' })).toBeInTheDocument()
  })

  it('an unknown route shows a not found page', () => {
    renderAt('/no-such-page')
    expect(screen.getByRole('heading', { name: 'Page not found' })).toBeInTheDocument()
  })

  it('/login shows the login form', () => {
    renderAt('/login')
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toHaveAttribute('type', 'password')
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeInTheDocument()
  })
})

describe('role guards', () => {
  it('sends a visitor who is not signed in to /login', () => {
    renderAt('/report/new')
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeInTheDocument()
  })

  it('shows the report form placeholder to a student', () => {
    signInAs('STUDENT')
    renderAt('/report/new')
    expect(screen.getByText('Report form: Nyendrak, SCRUM-35')).toBeInTheDocument()
  })

  it('refuses the report form to staff with a 403 page', () => {
    signInAs('STAFF')
    renderAt('/report/new')
    expect(screen.getByRole('heading', { name: 'Not allowed' })).toBeInTheDocument()
  })

  it.each([
    ['/staff', 'CATEGORY_HEAD', 'Staff view'],
    ['/staff', 'STAFF', 'Staff view'],
    ['/admin', 'SYSTEM_ADMIN', 'Admin dashboard'],
    ['/president', 'PRESIDENT', 'President dashboard'],
  ])('%s is open to %s', (path, role, heading) => {
    signInAs(role)
    renderAt(path)
    expect(screen.getByRole('heading', { name: heading })).toBeInTheDocument()
  })

  it.each([
    ['/staff', 'STUDENT'],
    ['/admin', 'CATEGORY_HEAD'],
    ['/president', 'SYSTEM_ADMIN'],
  ])('%s is refused to %s', (path, role) => {
    signInAs(role)
    renderAt(path)
    expect(screen.getByRole('heading', { name: 'Not allowed' })).toBeInTheDocument()
  })

  it('treats an expired session as signed out', () => {
    sessionStorage.setItem(
      SESSION_KEY,
      JSON.stringify({
        token: 'old',
        expiresAt: new Date(Date.now() - 1000).toISOString(),
        user: { id: 'u1', email: 'a@cst.edu.bt', fullName: 'Old', roles: ['STUDENT'] },
      }),
    )
    renderAt('/report/new')
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeInTheDocument()
  })
})

describe('login', () => {
  it('signs in through the gateway and shows the links for the role', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          accessToken: 'jwt-123',
          expiresAt: new Date(Date.now() + 3600_000).toISOString(),
          user: { id: 'u1', email: 'pema@cst.edu.bt', fullName: 'Pema', roles: ['STUDENT'] },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    renderAt('/login')

    await user.type(screen.getByLabelText('Email'), 'pema@cst.edu.bt')
    await user.type(screen.getByLabelText('Password'), 'secret-pass')
    await user.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByText('Signed in as Pema')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Report an issue' })).toBeInTheDocument()

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toMatch(/\/auth\/login$/)
    expect(init.body).toBe(JSON.stringify({ email: 'pema@cst.edu.bt', password: 'secret-pass' }))
    // The password is never kept in the browser.
    expect(sessionStorage.getItem(SESSION_KEY)).not.toContain('secret-pass')
  })

  it('shows the gateway error when the password is wrong', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: 'invalid email or password' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )
    const user = userEvent.setup()
    renderAt('/login')

    await user.type(screen.getByLabelText('Email'), 'pema@cst.edu.bt')
    await user.type(screen.getByLabelText('Password'), 'wrong')
    await user.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('invalid email or password')
  })

  it('signing out goes back to the signed-out home page', async () => {
    signInAs('STUDENT', 'Pema')
    const user = userEvent.setup()
    renderAt('/')
    await user.click(screen.getByRole('button', { name: 'Sign out' }))
    expect(screen.getByRole('link', { name: 'Sign in' })).toBeInTheDocument()
    expect(sessionStorage.getItem(SESSION_KEY)).toBeNull()
  })
})
