import { render, screen, within } from '@testing-library/react'
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
// The other roles see their own pages.
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

  // SCRUM-159: the admin and the President have one page each, so home takes
  // them straight to their dashboard.
  it.each([
    ['SYSTEM_ADMIN', 'Admin dashboard'],
    ['PRESIDENT', 'President dashboard'],
  ])('takes %s straight to their dashboard', async (role, heading) => {
    signInAs(role)
    renderHome()
    expect(await screen.findByRole('heading', { level: 1, name: heading })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: /Kuzuzangpo/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', routeHeading)).not.toBeInTheDocument()
  })
})

// SCRUM-158: a Category Head or a staff member (an HoD or a councillor) has one
// page, their board, so signing in takes them straight there, under their name.
describe('SCRUM-158 heads and staff land on their board', () => {
  afterEach(() => {
    sessionStorage.clear()
  })

  it.each(['CATEGORY_HEAD', 'STAFF'])('takes %s from home straight to their board', async (role) => {
    signInAs(role)
    renderHome()
    expect(await screen.findByRole('heading', { level: 1, name: 'Test User' })).toBeInTheDocument()
    expect(screen.getByText(/Reports assigned to you/)).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: /Kuzuzangpo/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', routeHeading)).not.toBeInTheDocument()
  })
})

// SCRUM-158: who receives each category, as the routing rules send it.
describe('SCRUM-158 who fixes what', () => {
  it('sends Hostel and Others to the SSO and the Academic Block to the department HoD', () => {
    renderHome()
    const places = screen.getByRole('region', { name: 'Who fixes what' })
    const hostel = within(places).getByRole('heading', { level: 3, name: 'Hostel' }).closest('li')!
    const block = within(places).getByRole('heading', { level: 3, name: 'Academic Block' }).closest('li')!
    const others = within(places).getByRole('heading', { level: 3, name: 'Others' }).closest('li')!
    expect(hostel).toHaveTextContent('Goes to the SSO')
    expect(block).toHaveTextContent("Goes to your department's HoD")
    expect(others).toHaveTextContent('Goes to the SSO')
  })

  it('says a late report goes up to the DSA or the DAA', () => {
    renderHome()
    expect(screen.getByText(/escalated to the DSA, or to the DAA for the Academic Block/)).toBeInTheDocument()
  })
})
