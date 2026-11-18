import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from '../App'
import { SESSION_KEY } from '../auth/AuthContext'

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn())
})

afterEach(() => {
  vi.unstubAllGlobals()
})

// SCRUM-150: the home page describes Others with the same text as category-taxonomy.
describe('home page categories', () => {
  it('describes Others as campus places, without Wi-Fi', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByText('Campus places such as the library, mess, gym and sports grounds.')).toBeInTheDocument()
    expect(screen.queryByText(/Wi-Fi, roads and grounds/)).not.toBeInTheDocument()
  })
})

// SCRUM-156: only students report, so only they see where a report goes.
// The other roles see their own pages, and a Category Head is told that late
// reports come back to them.
function signInAs(role: string) {
  sessionStorage.setItem(
    SESSION_KEY,
    JSON.stringify({
      token: 'test-token',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
      user: { id: 'u1', email: 'test@cst.edu.bt', fullName: 'Test User', roles: [role] },
    }),
  )
}

function renderHome() {
  return render(
    <MemoryRouter initialEntries={['/']}>
      <App />
    </MemoryRouter>,
  )
}

const routeHeading = { level: 2, name: 'Where your report goes' }
const escalatedToYou = /passes its time limit, it is escalated to you automatically/

describe('SCRUM-156 signed-in home by role', () => {
  afterEach(() => {
    sessionStorage.clear()
  })

  it('shows a student where their report goes', () => {
    signInAs('STUDENT')
    renderHome()
    expect(screen.getByRole('link', { name: 'Report an issue' })).toBeInTheDocument()
    expect(screen.getByRole('heading', routeHeading)).toBeInTheDocument()
  })

  it('tells a Category Head that late reports come to them, without the student route', () => {
    signInAs('CATEGORY_HEAD')
    renderHome()
    expect(screen.getByRole('link', { name: 'Staff view' })).toBeInTheDocument()
    expect(screen.getByText(escalatedToYou)).toBeInTheDocument()
    expect(screen.queryByRole('heading', routeHeading)).not.toBeInTheDocument()
  })

  it.each([
    ['STAFF', 'Staff view'],
    ['SYSTEM_ADMIN', 'Admin dashboard'],
    ['PRESIDENT', 'President dashboard'],
  ])('gives %s only their page', (role, tile) => {
    signInAs(role)
    renderHome()
    expect(screen.getByRole('link', { name: tile })).toBeInTheDocument()
    expect(screen.queryByRole('heading', routeHeading)).not.toBeInTheDocument()
    expect(screen.queryByText(escalatedToYou)).not.toBeInTheDocument()
  })
})
