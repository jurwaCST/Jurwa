// The College President's read-only dashboard (SCRUM-99, FR-29, FR-30).
// Written before the page. fetch is faked per URL, like the gateway answers.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from '../App'
import { SESSION_KEY } from '../auth/AuthContext'
import PresidentDashboard from './PresidentDashboard'

const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

// Counts arrive as strings: they are int64 in the contract (protojson).
const summary = {
  counts: { total: '6', open: '1', assigned: '1', inProgress: '1', resolved: '2', escalated: '1' },
  averageResolutionHours: 9,
}
const categories = {
  categories: [
    { category: 'HOSTEL', counts: { total: '2', open: '1', assigned: '0', inProgress: '0', resolved: '1', escalated: '0' } },
    { category: 'ACADEMIC_BLOCK', counts: { total: '2', open: '0', assigned: '0', inProgress: '1', resolved: '0', escalated: '1' } },
    { category: 'OTHERS', counts: { total: '2', open: '0', assigned: '1', inProgress: '0', resolved: '1', escalated: '0' } },
  ],
}

let fetchMock: ReturnType<typeof vi.fn>

beforeEach(() => {
  fetchMock = vi.fn((url: string) =>
    Promise.resolve(json(url.includes('/dashboard/categories') ? categories : summary)),
  )
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function renderPage() {
  return render(
    <MemoryRouter>
      <PresidentDashboard />
    </MemoryRouter>,
  )
}

describe('President dashboard', () => {
  it('shows the counts, the resolution rate and the average time from analytics', async () => {
    renderPage()

    const tiles = await screen.findByRole('region', { name: 'Report counts' })
    expect(within(tiles).getByText('Escalated').closest('.stat')).toHaveTextContent('1')
    expect(within(tiles).getByText('Resolved').closest('.stat')).toHaveTextContent('2')
    // 2 resolved of 6 reports = 33%
    expect(screen.getByText('33%')).toBeInTheDocument()
    expect(screen.getByText('9 hours')).toBeInTheDocument()
  })

  it('shows every category with its counts per status', async () => {
    renderPage()

    const table = await screen.findByRole('table', { name: 'Reports by category and status' })
    const hostel = within(table).getByRole('row', { name: /Hostel/ })
    expect(hostel).toHaveTextContent('Hostel')
    expect(within(table).getByRole('row', { name: /Academic Block/ })).toBeInTheDocument()
    expect(within(table).getByRole('row', { name: /Others/ })).toBeInTheDocument()
  })

  it('is read-only: no form, no field, and the only buttons are the period chips', async () => {
    const { container } = renderPage()
    await screen.findByRole('table')

    expect(screen.getByText('Read-only')).toBeInTheDocument()
    expect(container.querySelector('form, input, select, textarea')).toBeNull()
    const buttons = screen.getAllByRole('button').map((b) => b.textContent)
    expect(buttons).toEqual(['This week', 'This month', 'All time'])
    // Every call is a read.
    for (const [, init] of fetchMock.mock.calls as [string, RequestInit][]) {
      expect(init?.method ?? 'GET').toBe('GET')
    }
  })

  it('asks analytics for the chosen period', async () => {
    const user = userEvent.setup()
    renderPage()
    await screen.findByRole('table')
    // All time first: no period sent.
    expect(fetchMock.mock.calls[0][0]).toMatch(/\/dashboard\/summary$/)

    await user.click(screen.getByRole('button', { name: 'This week' }))

    expect(screen.getByRole('button', { name: 'This week' })).toHaveAttribute('aria-pressed', 'true')
    const lastUrl = fetchMock.mock.calls.at(-1)![0] as string
    expect(lastUrl).toMatch(/\?from=\d{4}-\d{2}-\d{2}T[\d:.]+Z$/)
  })

  it('shows an error with a way to try again when analytics cannot be reached', async () => {
    fetchMock.mockImplementation(() => Promise.resolve(json({ error: 'analytics unavailable' }, 503)))
    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByRole('alert')).toHaveTextContent('analytics unavailable')
    fetchMock.mockImplementation((url: string) =>
      Promise.resolve(json(url.includes('/dashboard/categories') ? categories : summary)),
    )
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findByRole('table')).toBeInTheDocument()
  })

  it('is only for the College President: a student gets the 403 page', () => {
    sessionStorage.setItem(
      SESSION_KEY,
      JSON.stringify({
        token: 't',
        expiresAt: new Date(Date.now() + 3600_000).toISOString(),
        user: { id: 'u1', email: 's@cst.edu.bt', fullName: 'Student', roles: ['STUDENT'] },
      }),
    )
    render(
      <MemoryRouter initialEntries={['/president']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByRole('heading', { name: 'Not allowed' })).toBeInTheDocument()
  })
})
