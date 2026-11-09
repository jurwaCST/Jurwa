import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import StaffView from './StaffView'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn() } }
})

// The signed-in Category Head.
const ME = '00000000-0000-4000-8000-000000000001'
vi.mock('../auth/useAuth', () => ({
  useAuth: () => ({
    user: { id: ME, email: 'dsa.demo@jurwa.example', fullName: 'Dorji DSA', roles: ['CATEGORY_HEAD'] },
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

const get = vi.mocked(api.get)
const post = vi.mocked(api.post)

// Shapes as the gateway sends them (protojson of report-history, status-lifecycle
// and internal-notes; see services/api-gateway/README.md).
function summary(reportId: string, status: string, building: string, area: string) {
  return {
    reportId,
    reporterId: 'student-1',
    category: 'HOSTEL',
    location: { building, wing: 'Wing 1', area },
    status,
    currentAssigneeId: ME,
    createdAt: '2026-11-08T08:00:00Z',
    updatedAt: '2026-11-08T09:00:00Z',
  }
}

const assigned = {
  reports: [
    summary('r1', 'ASSIGNED', 'Hostel A', 'Bathroom 1'),
    summary('r2', 'IN_PROGRESS', 'Hostel B', 'Kitchen'),
    summary('r3', 'ESCALATED', 'Hostel A', 'Corridor'),
    summary('r4', 'RESOLVED', 'Hostel C', 'Room 12'),
  ],
  nextPageToken: '',
}

let statuses: Record<string, { status: string; version: number }>

function stubGet(overrides: Record<string, unknown> = {}) {
  get.mockImplementation(async (path: string) => {
    if (path in overrides) {
      const v = overrides[path]
      if (v instanceof Error) throw v
      return v
    }
    if (path === '/me/assigned-reports') return assigned
    let m = path.match(/^\/reports\/(r\d)$/)
    if (m) {
      return {
        report: {
          id: m[1],
          category: 'HOSTEL',
          location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
          description: 'The tap has been leaking since Monday.',
          createdAt: '2026-11-08T08:00:00Z',
        },
      }
    }
    m = path.match(/^\/reports\/(r\d)\/status$/)
    if (m) return { status: { reportId: m[1], ...statuses[m[1]] } }
    m = path.match(/^\/reports\/(r\d)\/timeline$/)
    if (m) {
      return {
        entries: [
          { id: 'e1', eventType: 'report.submitted', summary: 'Report submitted', occurredAt: '2026-11-08T08:00:00Z' },
          { id: 'e2', eventType: 'report.routed', summary: 'Assigned to DSA', occurredAt: '2026-11-08T08:00:01Z' },
          {
            id: 'e3',
            eventType: 'status.changed',
            fromStatus: 'OPEN',
            toStatus: 'ASSIGNED',
            summary: 'Routed to the Category Head',
            occurredAt: '2026-11-08T08:00:02Z',
          },
        ],
      }
    }
    m = path.match(/^\/reports\/(r\d)\/notes$/)
    if (m) {
      return {
        notes: [{ id: 'n1', reportId: m[1], authorId: 'staff-9', body: 'Plumber called.', createdAt: '2026-11-08T10:00:00Z' }],
      }
    }
    throw new ApiError(404, 'not found')
  })
}

function renderView() {
  return render(
    <MemoryRouter>
      <StaffView />
    </MemoryRouter>,
  )
}

type User = ReturnType<typeof userEvent.setup>

// openReport opens the card of a report by its place.
async function openReport(user: User, place: RegExp) {
  await user.click(await screen.findByRole('button', { name: place }))
  return screen.findByRole('heading', { level: 2, name: place })
}

beforeEach(() => {
  statuses = {
    r1: { status: 'ASSIGNED', version: 2 },
    r2: { status: 'IN_PROGRESS', version: 3 },
    r3: { status: 'ESCALATED', version: 3 },
    r4: { status: 'RESOLVED', version: 4 },
  }
  stubGet()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('SCRUM-36 assigned reports', () => {
  it('lists the reports assigned to me with their place and status', async () => {
    renderView()
    expect(screen.getByRole('status', { name: 'Loading reports' })).toBeInTheDocument()
    const card = await screen.findByRole('button', { name: /Hostel A, Wing 1, Bathroom 1/ })
    expect(card).toHaveTextContent('Assigned')
    expect(screen.getAllByRole('listitem')).toHaveLength(4)
    expect(get).toHaveBeenCalledWith('/me/assigned-reports')
  })

  it('counts the reports per status on the chips and filters by one', async () => {
    const user = userEvent.setup()
    renderView()
    await screen.findByRole('button', { name: /Bathroom 1/ })
    expect(screen.getByRole('button', { name: 'All 4' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Escalated 1' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Escalated 1' }))
    expect(screen.getByRole('button', { name: 'Escalated 1' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getAllByRole('listitem')).toHaveLength(1)
    expect(screen.getByRole('button', { name: /Corridor/ })).toBeInTheDocument()
  })

  it('says so when nothing is assigned', async () => {
    stubGet({ '/me/assigned-reports': { reports: [], nextPageToken: '' } })
    renderView()
    expect(await screen.findByRole('heading', { name: 'Nothing assigned to you' })).toBeInTheDocument()
  })

  it('offers to try again when the list does not load', async () => {
    stubGet({ '/me/assigned-reports': new ApiError(503, 'report-history unavailable') })
    const user = userEvent.setup()
    renderView()
    expect(await screen.findByRole('alert')).toHaveTextContent('The service is busy or unreachable. Please try again.')
    stubGet()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findByRole('button', { name: /Bathroom 1/ })).toBeInTheDocument()
  })

  it('loads the next page when there are more reports', async () => {
    stubGet({
      '/me/assigned-reports': { reports: [summary('r1', 'ASSIGNED', 'Hostel A', 'Bathroom 1')], nextPageToken: 't2' },
      '/me/assigned-reports?pageToken=t2': { reports: [summary('r2', 'IN_PROGRESS', 'Hostel B', 'Kitchen')], nextPageToken: '' },
    })
    const user = userEvent.setup()
    renderView()
    await user.click(await screen.findByRole('button', { name: 'Load more' }))
    expect(await screen.findByRole('button', { name: /Kitchen/ })).toBeInTheDocument()
    expect(screen.getAllByRole('listitem')).toHaveLength(2)
    expect(screen.queryByRole('button', { name: 'Load more' })).not.toBeInTheDocument()
  })
})

describe('SCRUM-36 one report: details, history and notes', () => {
  it('opens a report on the Details tab with its description', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Hostel A, Wing 1, Bathroom 1/)
    expect(screen.getByRole('tab', { name: 'Details' })).toHaveAttribute('aria-selected', 'true')
    expect(await screen.findByText('The tap has been leaking since Monday.')).toBeInTheDocument()
    expect(get).toHaveBeenCalledWith('/reports/r1')
  })

  it('moves between the tabs with the arrow keys', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    screen.getByRole('tab', { name: 'Details' }).focus()
    await user.keyboard('{ArrowRight}')
    expect(screen.getByRole('tab', { name: 'History' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('tab', { name: 'History' })).toHaveFocus()
  })

  it('shows the history of status changes, oldest first', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(screen.getByRole('tab', { name: 'History' }))
    const panel = screen.getByRole('tabpanel', { name: 'History' })
    await waitFor(() => expect(within(panel).getAllByRole('listitem')).toHaveLength(2))
    const items = within(panel).getAllByRole('listitem')
    expect(items[0]).toHaveTextContent('Open')
    expect(items[1]).toHaveTextContent('Assigned')
    expect(items[1]).toHaveTextContent('Routed to the Category Head')
  })

  it('lists the notes and adds one', async () => {
    post.mockResolvedValueOnce({
      note: { id: 'n2', reportId: 'r1', authorId: ME, body: 'Spare part ordered.', createdAt: '2026-11-08T11:00:00Z' },
    })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(screen.getByRole('tab', { name: 'Notes' }))
    expect(await screen.findByText('Plumber called.')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Add note' }))
    expect(await screen.findByText('Write the note first.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()

    await user.type(screen.getByLabelText('New note'), 'Spare part ordered.')
    await user.click(screen.getByRole('button', { name: 'Add note' }))
    expect(await screen.findByText('Spare part ordered.')).toBeInTheDocument()
    expect(post).toHaveBeenCalledWith('/reports/r1/notes', { body: 'Spare part ordered.' })
    expect(screen.getByLabelText('New note')).toHaveValue('')
  })

  it('goes back to the list', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(screen.getByRole('button', { name: 'Back to the list' }))
    expect(screen.getByRole('button', { name: 'All 4' })).toBeInTheDocument()
  })
})

describe('SCRUM-36 update status', () => {
  it('offers only the next allowed status and sends it with the version', async () => {
    post.mockResolvedValueOnce({ status: { reportId: 'r1', status: 'IN_PROGRESS', version: 3 } })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))

    const sheet = screen.getByRole('dialog', { name: 'Update status' })
    expect(within(sheet).getAllByRole('radio').map((r) => r.closest('label')?.textContent)).toEqual(['In progress'])
    await user.click(within(sheet).getByRole('radio', { name: 'In progress' }))
    await user.type(within(sheet).getByLabelText('Note for the history (optional)'), 'Plumber on the way')
    await user.click(within(sheet).getByRole('button', { name: 'Save status' }))

    expect(post).toHaveBeenCalledWith('/reports/r1/status', {
      toStatus: 'IN_PROGRESS',
      expectedVersion: 2,
      reason: 'Plumber on the way',
    })
    expect(await screen.findByText('Status updated')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: /Bathroom 1/ }).closest('section')).toHaveTextContent('In progress')
  })

  it('keeps every letter typed in the sheet', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))
    const reason = screen.getByLabelText('Note for the history (optional)')
    await user.type(reason, 'abc')
    expect(reason).toHaveValue('abc')
    expect(reason).toHaveFocus()
  })

  it('says when someone else changed the report first, and shows the new status', async () => {
    post.mockRejectedValueOnce(new ApiError(409, 'version conflict'))
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))
    statuses.r1 = { status: 'IN_PROGRESS', version: 3 }
    await user.click(screen.getByRole('radio', { name: 'In progress' }))
    await user.click(screen.getByRole('button', { name: 'Save status' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Someone changed this report first. It now shows the latest status.',
    )
    await waitFor(() => expect(get).toHaveBeenCalledTimes(4)) // list, report, status, status again
  })

  it('needs a status to be chosen', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))
    await user.click(screen.getByRole('button', { name: 'Save status' }))
    expect(await screen.findByText('Choose the new status.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()
  })

  it('has nothing to update on a resolved report', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Room 12/)
    expect(await screen.findByText('Resolved. Nothing more to do.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Update status' })).not.toBeInTheDocument()
  })
})
