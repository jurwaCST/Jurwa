import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { SESSION_KEY } from './auth/AuthContext'
import { setToken } from './api/client'
import { markSeen } from './notifications/seen'

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
  localStorage.clear()
})

describe('routes', () => {
  it('home shows the app name, what it is for and a sign in link', () => {
    renderAt('/')
    expect(screen.getByRole('link', { name: 'Jurwa College of Science and Technology' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 1, name: 'See something broken on campus?' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Sign in' })).toBeInTheDocument()
  })

  // SCRUM-146: an escalated report goes back to its Category Head
  // (docs/design/escalation-saga.md), not to someone higher up.
  it('home says an escalated report goes to the Category Head', () => {
    renderAt('/')
    expect(screen.getByText(/the report is escalated to the Category Head automatically/)).toBeInTheDocument()
    expect(screen.queryByText(/next person up/)).not.toBeInTheDocument()
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

  it('shows the report form to a student', () => {
    signInAs('STUDENT')
    renderAt('/report/new')
    expect(screen.getByRole('heading', { name: 'Report an issue' })).toBeInTheDocument()
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
    ['/reports', 'STUDENT', 'My reports'],
    ['/notifications', 'STUDENT', 'Notifications'],
  ])('%s is open to %s', (path, role, heading) => {
    signInAs(role)
    renderAt(path)
    expect(screen.getByRole('heading', { name: heading })).toBeInTheDocument()
  })

  it.each([
    ['/staff', 'STUDENT'],
    ['/admin', 'CATEGORY_HEAD'],
    ['/president', 'SYSTEM_ADMIN'],
    ['/reports', 'STAFF'],
    ['/reports/r1', 'CATEGORY_HEAD'],
    ['/notifications', 'STAFF'],
  ])('%s is refused to %s', (path, role) => {
    signInAs(role)
    renderAt(path)
    expect(screen.getByRole('heading', { name: 'Not allowed' })).toBeInTheDocument()
  })

  it('gives a student a My reports tab', () => {
    signInAs('STUDENT')
    renderAt('/')
    for (const nav of screen.getAllByRole('navigation', { name: 'Main' })) {
      expect(within(nav).getByRole('link', { name: 'My reports' })).toHaveAttribute('href', '/reports')
    }
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
    expect(await screen.findByRole('link', { name: 'Report an issue' })).toBeInTheDocument()

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

  it('the sign-in page links to student sign-up', () => {
    renderAt('/login')
    expect(screen.getByRole('link', { name: 'Create a student account' })).toHaveAttribute('href', '/register')
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

describe('register', () => {
  const json = (status: number, body: unknown) =>
    new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

  async function fillIn(password = 'secret-pass') {
    const user = userEvent.setup()
    renderAt('/register')
    await user.type(screen.getByLabelText('Full name'), 'Pema Wangmo')
    await user.type(screen.getByLabelText('Email'), 'pema@cst.edu.bt')
    await user.type(screen.getByLabelText('Password'), password)
    await user.click(screen.getByRole('button', { name: 'Create account' }))
  }

  it('creates a student account, signs in and goes home', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        json(201, { user: { id: 'u1', email: 'pema@cst.edu.bt', fullName: 'Pema Wangmo', roles: ['STUDENT'] } }),
      )
      .mockResolvedValueOnce(
        json(200, {
          accessToken: 'jwt-123',
          expiresAt: new Date(Date.now() + 3600_000).toISOString(),
          user: { id: 'u1', email: 'pema@cst.edu.bt', fullName: 'Pema Wangmo', roles: ['STUDENT'] },
        }),
      )
    vi.stubGlobal('fetch', fetchMock)

    await fillIn()

    expect(await screen.findByText('Signed in as Pema Wangmo')).toBeInTheDocument()
    const [regUrl, regInit] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(regUrl).toMatch(/\/auth\/register$/)
    expect(regInit.body).toBe(
      JSON.stringify({ email: 'pema@cst.edu.bt', fullName: 'Pema Wangmo', password: 'secret-pass' }),
    )
    expect((fetchMock.mock.calls[1] as [string])[0]).toMatch(/\/auth\/login$/)
  })

  it('a password under 8 characters is refused before calling the server', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    await fillIn('short')

    expect(await screen.findByRole('alert')).toHaveTextContent('at least 8 characters')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('shows the server error when the email is already registered', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json(409, { error: 'email is already registered' })))

    await fillIn()

    expect(await screen.findByRole('alert')).toHaveTextContent('email is already registered')
  })

  // Something other than the gateway answered (a proxy page, another app on the
  // port): never show "Request failed (403)".
  it('shows a kind message, not the raw error, when something unexpected answers', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('<html>forbidden</html>', { status: 403 })))

    await fillIn()

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('Could not create the account. Try again.')
    expect(alert).not.toHaveTextContent('Request failed')
  })
})

// SCRUM-144: the bell in the header. There is no "mark as read" route, so the
// browser remembers the newest message the student has seen (per student).
describe('notifications bell', () => {
  const newest = {
    id: 'n1',
    eventType: 'status.changed',
    reportId: 'r1',
    subject: 'Your report is resolved',
    body: 'HC, Floor 1: it is marked as fixed.',
    createdAt: '2026-11-12T08:00:00Z',
  }
  const answer = (list: unknown[]) =>
    vi.fn().mockImplementation(async () =>
      new Response(JSON.stringify({ notifications: list, nextPageToken: '' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

  it('shows a student a bell with a dot when there is a new message', async () => {
    const fetchMock = answer([newest])
    vi.stubGlobal('fetch', fetchMock)
    signInAs('STUDENT')
    renderAt('/')

    expect(await screen.findByRole('link', { name: 'Notifications, new' })).toHaveAttribute('href', '/notifications')
    expect((fetchMock.mock.calls[0] as [string])[0]).toMatch(/\/me\/notifications\?pageSize=1$/)
  })

  it('shows no dot when the newest message was already seen', async () => {
    const fetchMock = answer([newest])
    vi.stubGlobal('fetch', fetchMock)
    signInAs('STUDENT')
    markSeen('u1', newest.createdAt)
    renderAt('/')

    await waitFor(() => expect(fetchMock).toHaveBeenCalled())
    expect(await screen.findByRole('link', { name: 'Notifications' })).toHaveAttribute('href', '/notifications')
    expect(screen.queryByRole('link', { name: 'Notifications, new' })).not.toBeInTheDocument()
  })

  it('shows no dot when there are no messages', async () => {
    const fetchMock = answer([])
    vi.stubGlobal('fetch', fetchMock)
    signInAs('STUDENT')
    renderAt('/')

    await waitFor(() => expect(fetchMock).toHaveBeenCalled())
    expect(await screen.findByRole('link', { name: 'Notifications' })).toBeInTheDocument()
  })

  it('gives staff no bell', () => {
    signInAs('STAFF')
    renderAt('/')
    expect(screen.queryByRole('link', { name: /^Notifications/ })).not.toBeInTheDocument()
  })
})
