import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import MyReports from './MyReports'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn(), blob: vi.fn() } }
})

const get = vi.mocked(api.get)

const DAY = 24 * 3600 * 1000
const ago = (ms: number) => new Date(Date.now() - ms).toISOString()

// GET /me/reports as the gateway sends it (report-history ListReportsByReporter).
function summary(reportId: string, status: string, category: string, building: string, area: string, age: number) {
  return {
    reportId,
    category,
    location: { building, wing: category === 'OTHERS' ? undefined : 'Floor 1', area },
    status,
    createdAt: ago(age),
    updatedAt: ago(age),
  }
}

const mine = {
  reports: [
    summary('89533b76-a375-47b2-aa7f-56032148bdea', 'ASSIGNED', 'HOSTEL', 'HC', 'Room 101', 2 * DAY),
    summary('c8826785-8732-4d65-88a0-c2f4ea10a926', 'RESOLVED', 'OTHERS', 'Campus', 'Library', 5 * DAY),
    summary('4f891548-a13c-43ad-a1c3-2a31fad1b3e9', 'IN_PROGRESS', 'ACADEMIC_BLOCK', 'Academic Block', 'LH1', 3 * 3600 * 1000),
  ],
  nextPageToken: '',
}

function renderPage() {
  return render(
    <MemoryRouter>
      <MyReports />
    </MemoryRouter>,
  )
}

beforeEach(() => {
  get.mockReset()
})

describe('SCRUM-143 My reports: the list', () => {
  it('shows a card per report with its number, status, category, place and age', async () => {
    get.mockResolvedValue(mine)
    renderPage()

    const cards = await screen.findAllByRole('listitem')
    expect(cards).toHaveLength(3)

    const first = within(cards[0])
    expect(first.getByRole('link', { name: /HC, Floor 1, Room 101/ })).toHaveAttribute(
      'href',
      '/reports/89533b76-a375-47b2-aa7f-56032148bdea',
    )
    expect(first.getByText('#89533B76')).toBeInTheDocument()
    expect(first.getByText('Assigned')).toBeInTheDocument()
    expect(first.getByText(/Hostel/)).toBeInTheDocument()
    expect(first.getByText(/2 days ago/)).toBeInTheDocument()

    expect(within(cards[2]).getByText(/3 hours ago/)).toBeInTheDocument()
    expect(get).toHaveBeenCalledWith('/me/reports')
  })

  it('filters by status with chips that say which one is pressed', async () => {
    const user = userEvent.setup()
    get.mockResolvedValue(mine)
    renderPage()

    const all = await screen.findByRole('button', { name: /All/ })
    expect(all).toHaveAttribute('aria-pressed', 'true')

    const resolved = screen.getByRole('button', { name: /Resolved/ })
    await user.click(resolved)
    expect(resolved).toHaveAttribute('aria-pressed', 'true')
    expect(all).toHaveAttribute('aria-pressed', 'false')
    expect(screen.getAllByRole('listitem')).toHaveLength(1)
    expect(screen.getByRole('link', { name: /Library/ })).toBeInTheDocument()
  })

  it('says when there are no reports yet, with a link to report one', async () => {
    get.mockResolvedValue({ reports: [], nextPageToken: '' })
    renderPage()

    expect(await screen.findByRole('heading', { name: 'No reports yet' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Report an issue' })).toHaveAttribute('href', '/report/new')
  })

  it('offers to try again when the list does not load, and loads it again', async () => {
    const user = userEvent.setup()
    get.mockRejectedValueOnce(new ApiError(503, 'unavailable')).mockResolvedValueOnce(mine)
    renderPage()

    expect(await screen.findByText(/service is busy or unreachable/i)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findAllByRole('listitem')).toHaveLength(3)
    expect(get).toHaveBeenCalledTimes(2)
  })

  it('loads the next page when there is one', async () => {
    const user = userEvent.setup()
    get.mockImplementation(async (path: string) => {
      if (path === '/me/reports') return { reports: mine.reports.slice(0, 2), nextPageToken: 'p2' }
      if (path === '/me/reports?pageToken=p2') return { reports: mine.reports.slice(2), nextPageToken: '' }
      throw new Error(`unexpected ${path}`)
    })
    renderPage()

    expect(await screen.findAllByRole('listitem')).toHaveLength(2)
    await user.click(screen.getByRole('button', { name: 'Load more' }))
    expect(await screen.findAllByRole('listitem')).toHaveLength(3)
    expect(screen.queryByRole('button', { name: 'Load more' })).not.toBeInTheDocument()
  })
})
