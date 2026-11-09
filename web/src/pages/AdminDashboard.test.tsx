import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import AdminDashboard from './AdminDashboard'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn() } }
})

const get = vi.mocked(api.get)

// Shapes as the gateway sends them: analytics counts are int64, so protojson
// sends them as strings, and a count of 0 may be left out.
const summary = {
  counts: { total: '11', open: '1', assigned: '2', inProgress: '3', resolved: '4', escalated: '1' },
  averageResolutionHours: 5.2,
}
const categories = {
  categories: [
    { category: 'HOSTEL', counts: { total: '7', open: '1', assigned: '1', inProgress: '2', resolved: '3', escalated: '1' } },
    { category: 'ACADEMIC_BLOCK', counts: { total: '3', assigned: '1', inProgress: '1', resolved: '1' } },
    { category: 'OTHERS', counts: { total: '1' } },
  ],
}

function stub(over: { summary?: unknown; categories?: unknown } = {}) {
  get.mockImplementation(async (path: string) => {
    if (path.startsWith('/dashboard/summary')) {
      const v = over.summary ?? summary
      if (v instanceof Error) throw v
      return v
    }
    if (path.startsWith('/dashboard/categories')) return over.categories ?? categories
    throw new ApiError(404, 'not found')
  })
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDashboard />
    </MemoryRouter>,
  )
}

beforeEach(() => stub())
afterEach(() => vi.clearAllMocks())

describe('SCRUM-37 admin dashboard: counts', () => {
  it('shows a tile per status with its count', async () => {
    renderPage()
    expect(screen.getByRole('heading', { level: 1, name: 'Admin dashboard' })).toBeInTheDocument()
    expect(screen.getByRole('status', { name: 'Loading the dashboard' })).toBeInTheDocument()
    const tiles = await screen.findByRole('region', { name: 'Report counts' })
    expect(within(tiles).getByText('Open').nextSibling).toHaveTextContent('1')
    expect(within(tiles).getByText('In progress').nextSibling).toHaveTextContent('3')
    expect(within(tiles).getByText('Escalated').nextSibling).toHaveTextContent('1')
    expect(within(tiles).getByText('Total').nextSibling).toHaveTextContent('11')
    expect(get).toHaveBeenCalledWith('/dashboard/summary')
    expect(get).toHaveBeenCalledWith('/dashboard/categories')
  })

  it('lists every category with its counts, a missing count being 0', async () => {
    renderPage()
    const table = await screen.findByRole('table', { name: 'Reports by category and status' })
    const academic = within(table).getByRole('row', { name: /Academic Block/ })
    expect(within(academic).getAllByRole('cell').map((c) => c.textContent)).toEqual(['0', '1', '1', '1', '0', '3'])
    expect(within(table).getByRole('row', { name: /Others/ })).toHaveTextContent('1')
  })

  it('points out what needs attention: escalated and not yet assigned', async () => {
    renderPage()
    const card = await screen.findByRole('region', { name: 'Needs attention' })
    expect(card).toHaveTextContent('1 report escalated')
    expect(card).toHaveTextContent('1 report not assigned yet')
  })

  it('says when nothing needs attention', async () => {
    stub({ summary: { counts: { total: '2', resolved: '2' }, averageResolutionHours: 3 } })
    renderPage()
    const card = await screen.findByRole('region', { name: 'Needs attention' })
    expect(card).toHaveTextContent('Nothing needs attention.')
  })

  it('reloads the counts for a period', async () => {
    const user = userEvent.setup()
    renderPage()
    await screen.findByRole('region', { name: 'Report counts' })
    await user.click(screen.getByRole('button', { name: 'This week' }))
    expect(screen.getByRole('button', { name: 'This week' })).toHaveAttribute('aria-pressed', 'true')
    await screen.findByRole('region', { name: 'Report counts' })
    expect(get).toHaveBeenCalledWith(expect.stringMatching(/^\/dashboard\/summary\?from=\d{4}-\d\d-\d\dT.*Z$/))
  })

  it('offers to try again when the counts do not load', async () => {
    stub({ summary: new ApiError(503, 'analytics unavailable') })
    const user = userEvent.setup()
    renderPage()
    expect(await screen.findByRole('alert')).toHaveTextContent('The service is busy or unreachable. Please try again.')
    stub()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findByRole('region', { name: 'Report counts' })).toBeInTheDocument()
  })

  it('tells the admin that rules and time limits are not editable here yet', async () => {
    renderPage()
    await screen.findByRole('region', { name: 'Report counts' })
    expect(screen.getByRole('region', { name: 'Routing rules and time limits' })).toHaveTextContent(
      'Editing them here comes next',
    )
  })
})
